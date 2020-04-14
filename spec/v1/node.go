package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/baetyl/baetyl-go/log"
	v1 "github.com/census-instrumentation/opencensus-proto/gen-go/agent/common/v1"
	"reflect"
	"strconv"
	"time"

	"github.com/evanphx/json-patch"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// maxJSONLevel the max level of json
const (
	maxJSONLevel   = 5
	milliPrecision = 1000
)

// ErrJSONLevelExceedsLimit the level of json exceeds the max limit
var ErrJSONLevelExceedsLimit = fmt.Errorf("the level of json exceeds the max limit (%d)", maxJSONLevel)

// Node the spec of node
type Node struct {
	Namespace         string            `json:"namespace,omitempty"`
	Name              string            `json:"name,omitempty" validate:"omitempty,resourceName"`
	Version           string            `json:"version,omitempty"`
	CreationTimestamp time.Time         `json:"createTime,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Report            Report            `json:"report,omitempty"`
	Desire            Desire            `json:"desire,omitempty"`
	Description       string            `json:"description,omitempty"`
}

type NodeReportView struct {
	Namespace string      `json:"namespace,omitempty"`
	Name      string      `json:"name,omitempty"`
	Version   string      `json:"version,omitempty"`
	Report    *ReportView `json:"report,omitempty"`
	Ready     bool        `json:"ready"`
}

type ReportView struct {
	Time       time.Time   `json:"time,omitempty"`
	Apps       []AppInfo   `json:"apps,omitempty"`
	Core       *CoreInfo   `json:"core,omitempty"`
	Appstats   []AppStatus `json:"appstats,omitempty"`
	Node       *NodeInfo   `json:"node,omitempty"`
	NodeStatus *NodeStatus `json:"nodestats,omitempty"`
}

// Report report data
type Report map[string]interface{}

// Desire desire data
type Desire map[string]interface{}

// AppInfos return app infos
func (r Report) AppInfos() []AppInfo {
	return getAppInfos("apps", r)
}

// AppInfos return sysapps infos
func (r Report) SysAppInfos() []AppInfo {
	return getAppInfos("sysapps", r)
}

// Merge merge new reported data
func (r Report) Merge(reported Report) error {
	return merge(r, reported, 1, maxJSONLevel)
}

// AppInfos return app infos
func (d Desire) AppInfos() []AppInfo {
	return getAppInfos("apps", d)
}

// AppInfos return sysapps infos
func (d Desire) SysAppInfos() []AppInfo {
	return getAppInfos("sysapps", d)
}

// Merge merge new reported data
func (d Desire) Merge(desired Desire) error {
	return merge(d, desired, 1, maxJSONLevel)
}

// Diff diff with reported data, return the delta fo desire
func (d Desire) Diff(reported Report) (Desire, error) {
	return diff(d, reported)
}

func getAppInfos(appType string, data map[string]interface{}) []AppInfo {
	if data == nil {
		return nil
	}
	apps, ok := data[appType]
	if !ok || apps == nil {
		return nil
	}
	res, ok := apps.([]AppInfo)
	if ok {
		return res
	}
	res = []AppInfo{}
	ais, ok := apps.([]interface{})
	if !ok {
		return nil
	}
	for _, ai := range ais {
		aim := ai.(map[string]interface{})
		if aim == nil {
			return nil
		}
		res = append(res, AppInfo{Name: aim["name"].(string), Version: aim["version"].(string)})
	}
	return res
}

// merge right map into left map
func merge(left, right map[string]interface{}, depth, maxDepth int) error {
	if depth >= maxDepth {
		return ErrJSONLevelExceedsLimit
	}
	for rk, rv := range right {
		lv, ok := left[rk]
		if !ok || rv == nil || reflect.TypeOf(rv).Kind() != reflect.Map || reflect.TypeOf(lv).Kind() != reflect.Map {
			left[rk] = rv
			continue
		}
		if err := merge(lv.(map[string]interface{}), rv.(map[string]interface{}), depth+1, maxDepth); err != nil {
			return err
		}
	}
	return nil
}

func diff(desired, reported map[string]interface{}) (map[string]interface{}, error) {
	var delta map[string]interface{}
	r, err := json.Marshal(reported)
	if err != nil {
		return delta, err
	}
	d, err := json.Marshal(desired)
	if err != nil {
		return delta, err
	}
	patch, err := jsonpatch.CreateMergePatch(r, d)
	if err != nil {
		return delta, err
	}
	err = json.Unmarshal(patch, &delta)
	if err != nil {
		return delta, err
	}
	clean(delta)
	return delta, nil
}

func clean(m map[string]interface{}) {
	for k, v := range m {
		if v == nil {
			delete(m, k)
			continue
		}
		bk := reflect.TypeOf(v).Kind()
		if bk != reflect.Map {
			continue
		}
		if vm, ok := v.(map[string]interface{}); ok {
			clean(vm)
		}
	}
}

func translateNodeToNodeReportView(node *v1.Node) (*NodeReportView, error) {
	view := new(NodeReportView)
	nodeStr, err := json.Marshal(node)
	if err != nil {
		log.L().Error("node to node report view error", log.Error(err))
		return nil, err
	}
	err = json.Unmarshal(nodeStr, view)
	if err != nil {
		log.L().Error("node to node report view error", log.Error(err))
		return nil, err
	}

	return view, nil
}

func processPercent(report *ReportView) error {
	if report == nil || report.NodeStatus == nil {
		return nil
	}
	nodeStatus := report.NodeStatus
	nodeStatus.Percent = map[string]string{}

	memory := string(coreV1.ResourceMemory)
	mPercent, err := getResourceUsagePercent(nodeStatus, memory)
	if err != nil {
		return err
	}
	nodeStatus.Percent[memory] = mPercent
	cpu := string(coreV1.ResourceCPU)
	cpuPercent, err := getResourceUsagePercent(nodeStatus, cpu)
	if err != nil {
		return err
	}
	nodeStatus.Percent[cpu] = cpuPercent
	return nil
}

func getResourceUsagePercent(status *NodeStatus, resourceType string) (string, error) {
	cap, capOk := status.Capacity[resourceType]
	usg, usageOk := status.Usage[resourceType]
	total := int64(0)
	usage := int64(0)
	var err error
	if capOk {
		switch resourceType {
		case string(coreV1.ResourceCPU):
			total, err = translateQuantityToDecimal(cap, true)
			if err != nil {
				return "", err
			}
			status.Capacity[resourceType] = strconv.FormatFloat(float64(total)/milliPrecision, 'f', -1, 64)
		case string(coreV1.ResourceMemory):
			total, err = translateQuantityToDecimal(cap, false)
			if err != nil {
				return "", err
			}
			status.Capacity[resourceType] = strconv.FormatInt(total, 10)
		default:
			return "", errors.New("unsupported resource type")
		}
	}

	if usageOk {
		switch resourceType {
		case string(coreV1.ResourceCPU):
			usage, err = translateQuantityToDecimal(usg, true)
			if err != nil {
				return "", err
			}
			status.Usage[resourceType] = strconv.FormatFloat(float64(usage)/milliPrecision, 'f', -1, 64)
		case string(coreV1.ResourceMemory):
			usage, err = translateQuantityToDecimal(usg, false)
			if err != nil {
				return "", err
			}
			status.Usage[resourceType] = strconv.FormatInt(usage, 10)
		default:
			return "", errors.New("unsupported resource type")
		}
	}

	ratio := float64(0)
	if capOk && usageOk {
		if total != 0 {
			ratio = float64(usage) / float64(total)
		}
	}

	return strconv.FormatFloat(ratio, 'f', -1, 64), nil
}

func translateQuantityToDecimal(q string, milli bool) (int64, error) {
	num, err := resource.ParseQuantity(q)
	if err != nil {
		return 0, err
	}
	if milli {
		return num.MilliValue(), nil
	}
	return num.Value(), nil
}
