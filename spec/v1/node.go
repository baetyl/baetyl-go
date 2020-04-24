package v1

import (
	"encoding/json"
	"fmt"
	"github.com/baetyl/baetyl-go/log"
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
	TimePattern    = "2006-01-02T15:04:05.999999999Z"
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

type NodeView struct {
	Namespace         string            `json:"namespace,omitempty"`
	Name              string            `json:"name,omitempty"`
	Version           string            `json:"version,omitempty"`
	CreationTimestamp time.Time         `json:"createTime,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Report            *ReportView       `json:"report,omitempty"`
	Desire            Desire            `json:"desire,omitempty"`
	Description       string            `json:"description,omitempty"`
	Ready             bool              `json:"ready"`
}

type ReportView struct {
	Time       time.Time   `json:"time,omitempty"`
	Apps       []AppInfo   `json:"apps,omitempty"`
	SysApps    []AppInfo   `json:"sysapps,omitempty"`
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

func (n *Node) View(timeout time.Duration) *NodeView {
	view := new(NodeView)
	nodeStr, err := json.Marshal(n)
	if err != nil {
		log.L().Error("failed to convert to node view", log.Error(err))
		return nil
	}
	err = json.Unmarshal(nodeStr, view)
	if err != nil {
		log.L().Error("failed to convert to node view", log.Error(err))
		return nil
	}
	if err = view.populateNodeStatus(timeout); err != nil {
		log.L().Error("failed to populate node status", log.Error(err))
		return nil
	}
	return view
}

func (view *NodeView) populateNodeStatus(timeout time.Duration) error {
	if view.Report == nil || view.Report.NodeStatus == nil {
		return nil
	}

	s := view.Report.NodeStatus
	s.Percent = map[string]string{}
	memory := string(coreV1.ResourceMemory)
	mPercent, err := s.processResourcePercent(s, memory, populateMemoryResource)
	if err != nil {
		return err
	}
	s.Percent[memory] = mPercent

	cpu := string(coreV1.ResourceCPU)
	cpuPercent, err := s.processResourcePercent(s, cpu, populateCPUResource)
	if err != nil {
		return err
	}
	s.Percent[cpu] = cpuPercent

	view.Ready = time.Now().Before(view.Report.Time.Add(timeout))
	return nil
}

func (s *NodeStatus) processResourcePercent(status *NodeStatus, resourceType string,
	populate func(usage string, resource map[string]string) (int64, error)) (string, error) {
	cap, capOk := status.Capacity[resourceType]
	usg, usageOk := status.Usage[resourceType]
	var total, usage int64
	var err error
	if capOk {
		if total, err = populate(cap, status.Capacity); err != nil {
			return "0", err
		}
	}
	if usageOk {
		if usage, err = populate(usg, status.Usage); err != nil {
			return "0", err
		}
	}

	if capOk && usageOk && total != 0 {
		return strconv.FormatFloat(float64(usage)/float64(total), 'f', -1, 64), nil
	}
	return "0", nil
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

func populateCPUResource(usage string, resource map[string]string) (int64, error) {
	usg, err := translateQuantityToDecimal(usage, true)
	if err != nil {
		return 0, err
	}
	resource[string(coreV1.ResourceCPU)] = strconv.FormatFloat(float64(usg)/milliPrecision, 'f', -1, 64)
	return usg, nil
}

func populateMemoryResource(usage string, resource map[string]string) (int64, error) {
	usg, err := translateQuantityToDecimal(usage, false)
	if err != nil {
		return 0, err
	}
	resource[string(coreV1.ResourceMemory)] = strconv.FormatInt(usg, 10)
	return usg, nil
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
