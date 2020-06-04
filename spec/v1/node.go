package v1

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/baetyl/baetyl-go/errors"
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
	Labels            map[string]string `json:"labels,omitempty" validate:"omitempty,validLabels"`
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
	Time        *time.Time  `json:"time,omitempty"`
	Apps        []AppInfo   `json:"apps,omitempty"`
	SysApps     []AppInfo   `json:"sysapps,omitempty"`
	Core        *CoreInfo   `json:"core,omitempty"`
	AppStats    []AppStatus `json:"appstats,omitempty"`
	SysAppStats []AppStatus `json:"sysappstats,omitempty"`
	Node        *NodeInfo   `json:"node,omitempty"`
	NodeStatus  *NodeStatus `json:"nodestats,omitempty"`
}

// Report report data
type Report map[string]interface{}

// Desire desire data
type Desire map[string]interface{}

// Merge merge new reported data
func (r Report) Merge(reported Report) error {
	return errors.Trace(merge(r, reported, 1, maxJSONLevel))
}

// Merge merge new reported data
func (d Desire) Merge(desired Desire) error {
	return errors.Trace(merge(d, desired, 1, maxJSONLevel))
}

// Diff diff with reported data, return the delta fo desire
func (d Desire) Diff(reported Report) (Desire, error) {
	res, err := diff(d, reported)
	return res, errors.Trace(err)
}

func (r Report) AppInfos(isSys bool) []AppInfo {
	if isSys {
		return getAppInfos("sysapps", r)
	} else {
		return getAppInfos("apps", r)
	}
}

func (r Report) SetAppInfos(isSys bool, apps []AppInfo) {
	if isSys {
		r["sysapps"] = apps
	} else {
		r["apps"] = apps
	}
}

func (d Desire) AppInfos(isSys bool) []AppInfo {
	if isSys {
		return getAppInfos("sysapps", d)
	} else {
		return getAppInfos("apps", d)
	}
}

func (d Desire) SetAppInfos(isSys bool, apps []AppInfo) {
	if isSys {
		d["sysapps"] = apps
	} else {
		d["apps"] = apps
	}
}

func (r Report) SetAppStats(isSys bool, stats []AppStatus) {
	if isSys {
		r["sysappstats"] = stats
	} else {
		r["appstats"] = stats
	}
}

func (d Desire) SetAppStats(isSys bool, stats []AppStatus) {
	if isSys {
		d["sysappstats"] = stats
	} else {
		d["appstats"] = stats
	}
}

func (r Report) AppStats(isSys bool) []AppStatus {
	if isSys {
		return getAppStats("sysappstats", r)
	} else {
		return getAppStats("appstats", r)
	}
}

func (d Desire) AppStats(isSys bool) []AppStatus {
	if isSys {
		return getAppStats("sysappstats", d)
	} else {
		return getAppStats("appstats", d)
	}
}

func (n *Node) View(timeout time.Duration) (*NodeView, error) {
	view := new(NodeView)
	nodeStr, err := json.Marshal(n)
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = json.Unmarshal(nodeStr, view)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if err = view.populateNodeStatus(timeout); err != nil {
		return nil, errors.Trace(err)
	}
	if report := view.Report; report != nil {
		if err = report.translateServiceResouceQuantity(); err != nil {
			return nil, errors.Trace(err)
		}
	}
	return view, nil
}

func (view *NodeView) populateNodeStatus(timeout time.Duration) (err error) {
	if view.Report == nil || view.Report.NodeStatus == nil {
		return nil
	}

	s := view.Report.NodeStatus
	s.Percent = map[string]string{}
	memory := string(coreV1.ResourceMemory)
	if s.Percent[memory], err = s.processResourcePercent(s, memory, populateMemoryResource); err != nil {
		return errors.Trace(err)
	}

	cpu := string(coreV1.ResourceCPU)
	if s.Percent[cpu], err = s.processResourcePercent(s, cpu, populateCPUResource); err != nil {
		return errors.Trace(err)
	}

	if view.Report.Time != nil {
		view.Ready = time.Now().Before(view.Report.Time.Add(timeout))
	}

	return
}

func (s *NodeStatus) processResourcePercent(status *NodeStatus, resourceType string,
	populate func(usage string, resource map[string]string) (int64, error)) (string, error) {
	cap, capOk := status.Capacity[resourceType]
	usg, usageOk := status.Usage[resourceType]
	var total, usage int64
	var err error
	if capOk {
		if total, err = populate(cap, status.Capacity); err != nil {
			return "0", errors.Trace(err)
		}
	}
	if usageOk {
		if usage, err = populate(usg, status.Usage); err != nil {
			return "0", errors.Trace(err)
		}
	}

	if capOk && usageOk && total != 0 {
		return strconv.FormatFloat(float64(usage)/float64(total), 'f', -1, 64), nil
	}
	return "0", nil
}

func (view *ReportView) translateServiceResouceQuantity() error {
	for idx := range view.AppStats {
		services := view.AppStats[idx].ServiceInfos
		if services == nil {
			continue
		}
		for _, v := range services {
			if err := v.translateResouceQuantity(); err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

func (s *ServiceInfo) translateResouceQuantity() error {
	if s.Usage == nil {
		return nil
	}

	if cpuUsage, cpuOk := s.Usage[string(coreV1.ResourceCPU)]; cpuOk {
		if _, err := populateCPUResource(cpuUsage, s.Usage); err != nil {
			return errors.Trace(err)
		}
	}

	if memoryUsage, mOk := s.Usage[string(coreV1.ResourceMemory)]; mOk {
		if _, err := populateMemoryResource(memoryUsage, s.Usage); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
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

func getAppStats(statsType string, data map[string]interface{}) []AppStatus {
	if data == nil {
		return nil
	}
	apps, ok := data[statsType]
	if !ok || apps == nil {
		return nil
	}
	if res, ok := apps.([]AppStatus); ok {
		return res
	} else {
		return nil
	}
}

// merge right map into left map
func merge(left, right map[string]interface{}, depth, maxDepth int) error {
	if depth >= maxDepth {
		return ErrJSONLevelExceedsLimit
	}
	for rk, rv := range right {
		lv, ok := left[rk]
		if !ok || lv == nil || rv == nil || reflect.TypeOf(rv).Kind() != reflect.Map || reflect.TypeOf(lv).Kind() != reflect.Map {
			left[rk] = rv
			continue
		}
		if err := merge(lv.(map[string]interface{}), rv.(map[string]interface{}), depth+1, maxDepth); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func diff(desired, reported map[string]interface{}) (map[string]interface{}, error) {
	var delta map[string]interface{}
	r, err := json.Marshal(reported)
	if err != nil {
		return delta, errors.Trace(err)
	}
	d, err := json.Marshal(desired)
	if err != nil {
		return delta, errors.Trace(err)
	}
	patch, err := jsonpatch.CreateMergePatch(r, d)
	if err != nil {
		return delta, errors.Trace(err)
	}
	err = json.Unmarshal(patch, &delta)
	if err != nil {
		return delta, errors.Trace(err)
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
		return 0, errors.Trace(err)
	}
	resource[string(coreV1.ResourceCPU)] = strconv.FormatFloat(float64(usg)/milliPrecision, 'f', -1, 64)
	return usg, nil
}

func populateMemoryResource(usage string, resource map[string]string) (int64, error) {
	usg, err := translateQuantityToDecimal(usage, false)
	if err != nil {
		return 0, errors.Trace(err)
	}
	resource[string(coreV1.ResourceMemory)] = strconv.FormatInt(usg, 10)
	return usg, nil
}

func translateQuantityToDecimal(q string, milli bool) (int64, error) {
	num, err := resource.ParseQuantity(q)
	if err != nil {
		return 0, errors.Trace(err)
	}
	if milli {
		return num.MilliValue(), nil
	}
	return num.Value(), nil
}
