package v1

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/evanphx/json-patch"
	"github.com/mitchellh/mapstructure"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
)

// maxJSONLevel the max level of json
const (
	maxJSONLevel                   = 5
	milliPrecision                 = 1000
	KeySyncMode                    = "syncMode"
	CloudMode             SyncMode = "cloud"
	LocalMode             SyncMode = "local"
	KeyNodeProps                   = "nodeprops"
	KeyDevices                     = "devices"
	KeyApps                        = "apps"
	KeySysApps                     = "sysapps"
	KeyAppStats                    = "appstats"
	KeySysAppStats                 = "sysappstats"
	KeyAccelerator                 = "accelerator"
	KeyCluster                     = "cluster"
	KeyOptionalSysApps             = "optionalSysApps"
	KeyNodeMode                    = "nodeMode"
	NVAccelerator                  = "nvidia"
	JetsonAccelerator              = "jetson"
	AscendAccelerator              = "ascend"
	BitmainAccelerator             = "bitmain"
	CambriconAccelerator           = "cambricon"
	KunLunAccelerator              = "kunlun"
	ResourceGPU                    = "gpu"
	ResourceDisk                   = "disk"
	KeyGPUUsedMemory               = "usedMemory"
	KeyGPUTotalMemory              = "totalMemory"
	KeyGPUPercent                  = "percent"
	KeyDiskUsed                    = "diskUsed"
	KeyDiskTotal                   = "diskTotal"
	KeyDiskPercent                 = "diskPercent"
	KeyNetBytesSent                = "netBytesSent"
	KeyNetBytesRecv                = "netBytesRecv"
	KeyNetPacketsSent              = "netPacketsSent"
	KeyNetPacketsRecv              = "netPacketsRecv"
	KeyAppRequestCnt               = "requestCounter"
	KeyAppRequestTotalCnt          = "requestTotal"
	KeyLink                        = "link"
	KeyCoreId                      = "coreId"
	MQTTLink                       = "mqtt"

	BaetylCoreFrequency = "BaetylCoreFrequency"
	BaetylCoreAPIPort   = "BaetylCoreAPIPort"
	BaetylCoreVersion   = "BaetylCoreVersion"
	BaetylAgentPort     = "BaetylAgentPort"

	NodeOffline   = 0
	NodeOnline    = 1
	NodeUninstall = 2
)

var acceleratorMap = map[string]bool{
	NVAccelerator:        true,
	JetsonAccelerator:    true,
	AscendAccelerator:    true,
	BitmainAccelerator:   true,
	CambriconAccelerator: true,
	KunLunAccelerator:    true,
}

type SyncMode string

// ErrJSONLevelExceedsLimit the level of json exceeds the max limit
var ErrJSONLevelExceedsLimit = fmt.Errorf("the level of json exceeds the max limit (%d)", maxJSONLevel)

// Node the spec of node
type Node struct {
	Namespace         string                 `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Name              string                 `json:"name,omitempty" yaml:"name,omitempty" validate:"omitempty,resourceName"`
	Version           string                 `json:"version,omitempty" yaml:"version,omitempty"`
	CreationTimestamp time.Time              `json:"createTime,omitempty" yaml:"createTime,omitempty"`
	Accelerator       string                 `json:"accelerator,omitempty" yaml:"accelerator,omitempty"`
	Mode              SyncMode               `json:"mode,omitempty" yaml:"mode,omitempty"`
	NodeMode          string                 `json:"nodeMode,omitempty" yaml:"nodeMode,omitempty"`
	Cluster           bool                   `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Labels            map[string]string      `json:"labels,omitempty" yaml:"labels,omitempty" validate:"omitempty,validLabels"`
	Annotations       map[string]string      `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Attributes        map[string]interface{} `json:"attr,omitempty" yaml:"attr,omitempty"`
	Report            Report                 `json:"report,omitempty" yaml:"report,omitempty"`
	Desire            Desire                 `json:"desire,omitempty" yaml:"desire,omitempty"`
	SysApps           []string               `json:"sysApps,omitempty" yaml:"sysApps,omitempty"`
	Description       string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Link              string                 `json:"link,omitempty" yaml:"link,omitempty"`
	CoreId            string                 `json:"coreId,omitempty" yaml:"coreId,omitempty"`
}

type NodeView struct {
	Namespace         string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Name              string            `json:"name,omitempty" yaml:"name,omitempty"`
	Version           string            `json:"version,omitempty" yaml:"version,omitempty"`
	CreationTimestamp time.Time         `json:"createTime,omitempty" yaml:"createTime,omitempty"`
	Accelerator       string            `json:"accelerator,omitempty" yaml:"accelerator,omitempty"`
	Labels            map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Report            *ReportView       `json:"report,omitempty" yaml:"report,omitempty"`
	AppMode           string            `json:"appMode,omitempty" yaml:"appMode,omitempty"`
	Desire            Desire            `json:"desire,omitempty" yaml:"desire,omitempty"`
	SysApps           []string          `json:"sysApps,omitempty" yaml:"sysApps,omitempty"`
	Description       string            `json:"description,omitempty" yaml:"description,omitempty"`
	Cluster           bool              `json:"cluster" yaml:"cluster"`
	Ready             int               `json:"ready"`
	Mode              SyncMode          `json:"mode"`
	NodeMode          string            `json:"nodeMode,omitempty" yaml:"nodeMode,omitempty"`
	Link              string            `json:"link,omitempty" yaml:"link,omitempty"`
	CoreId            string            `json:"coreId,omitempty" yaml:"coreId,omitempty"`
}

type ReportView struct {
	Time        *time.Time            `json:"time,omitempty" yaml:"time,omitempty"`
	Apps        []AppInfo             `json:"apps,omitempty" yaml:"apps,omitempty"`
	SysApps     []AppInfo             `json:"sysapps,omitempty" yaml:"sysapps,omitempty"`
	Core        *CoreInfo             `json:"core,omitempty" yaml:"core,omitempty"`
	AppStats    []AppStats            `json:"appstats,omitempty" yaml:"appstats,omitempty"`
	SysAppStats []AppStats            `json:"sysappstats,omitempty" yaml:"sysappstats,omitempty"`
	Node        map[string]*NodeInfo  `json:"node,omitempty" yaml:"node,omitempty"`
	NodeStats   map[string]*NodeStats `json:"nodestats,omitempty" yaml:"nodestats,omitempty"`
	NodeInsNum  map[string]int        `json:"nodeinsnum,omitempty" yaml:"nodeinsnum,omitempty"`
	ModeInfo    string                `yaml:"modeinfo,omitempty" json:"modeinfo,omitempty"`
}

// Report report data
type Report map[string]interface{}

// Desire desire data
type Desire map[string]interface{}

// Delta delta data
type Delta map[string]interface{}

// Merge merge new reported data
func (r Report) Merge(reported Report) error {
	return errors.Trace(merge(r, reported, 1, maxJSONLevel))
}

// Merge merge new reported data
func (d Desire) Merge(desired Desire) error {
	return errors.Trace(merge(d, desired, 1, maxJSONLevel))
}

// Diff diff with reported data, return the delta for desire
func (d Desire) Diff(reported Report) (Desire, error) {
	res, err := diff(d, reported, true)
	return res, errors.Trace(err)
}

// Diff desire diff with report data, return the delta for desire
// and do not clean nil in delta
func (d Desire) DiffWithNil(report Report) (Delta, error) {
	res, err := diff(d, report, false)
	return res, errors.Trace(err)
}

// Patch patch desire with delta, get the new desire
func (d Desire) Patch(delta Delta) (Desire, error) {
	return patch(d, delta)
}

// Patch patch report with delta, get the new report
func (r Report) Patch(delta Delta) (Report, error) {
	return patch(r, delta)
}

func patch(doc, delta map[string]interface{}) (map[string]interface{}, error) {
	docData, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	deltaData, err := json.Marshal(delta)
	if err != nil {
		return nil, err
	}
	patchData, err := jsonpatch.MergePatch(docData, deltaData)
	if err != nil {
		return nil, err
	}
	var newDoc map[string]interface{}
	if err = json.Unmarshal(patchData, &newDoc); err != nil {
		return nil, err
	}
	return newDoc, nil
}

func getDeviceInfos(data map[string]interface{}) []DeviceInfo {
	if data == nil {
		return nil
	}
	devs, ok := data[KeyDevices]
	if !ok || devs == nil {
		return nil
	}
	res, ok := devs.([]DeviceInfo)
	if ok {
		return res
	}
	res = []DeviceInfo{}
	dis, ok := devs.([]interface{})
	if !ok {
		return nil
	}
	for _, di := range dis {
		dim := di.(map[string]interface{})
		if dim == nil {
			return nil
		}
		res = append(res, DeviceInfo{Name: dim["name"].(string), Version: dim["version"].(string)})
	}
	return res
}

func (r Report) DeviceInfos() []DeviceInfo {
	return getDeviceInfos(r)
}

func (r Report) SetDeviceInfos(devs []DeviceInfo) {
	r[KeyDevices] = devs
}

func (d Desire) DeviceInfos() []DeviceInfo {
	return getDeviceInfos(d)
}

func (d Desire) SetDeviceInfos(devs []DeviceInfo) {
	d[KeyDevices] = devs
}

func (r Report) AppInfos(isSys bool) []AppInfo {
	if isSys {
		return getAppInfos(KeySysApps, r)
	} else {
		return getAppInfos(KeyApps, r)
	}
}

func (r Report) SetAppInfos(isSys bool, apps []AppInfo) {
	if isSys {
		r[KeySysApps] = apps
	} else {
		r[KeyApps] = apps
	}
}

func (d Desire) AppInfos(isSys bool) []AppInfo {
	if isSys {
		return getAppInfos(KeySysApps, d)
	} else {
		return getAppInfos(KeyApps, d)
	}
}

func (d Desire) SetAppInfos(isSys bool, apps []AppInfo) {
	if isSys {
		d[KeySysApps] = apps
	} else {
		d[KeyApps] = apps
	}
}

func (r Report) SetAppStats(isSys bool, stats []AppStats) {
	if isSys {
		r[KeySysAppStats] = stats
	} else {
		r[KeyAppStats] = stats
	}
}

func (d Desire) SetAppStats(isSys bool, stats []AppStats) {
	if isSys {
		d[KeySysAppStats] = stats
	} else {
		d[KeyAppStats] = stats
	}
}

func (r Report) AppStats(isSys bool) []AppStats {
	if isSys {
		return getAppStats(KeySysAppStats, r)
	} else {
		return getAppStats(KeyAppStats, r)
	}
}

func (d Desire) AppStats(isSys bool) []AppStats {
	if isSys {
		return getAppStats(KeySysAppStats, d)
	} else {
		return getAppStats(KeyAppStats, d)
	}
}

func (n *Node) View(timeout time.Duration) (*NodeView, error) {
	err := n.compatibleSingleNode()
	if err != nil {
		return nil, errors.Trace(err)
	}
	view := new(NodeView)
	nodeStr, err := json.Marshal(n)
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = json.Unmarshal(nodeStr, view)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if err = view.populateNodeStats(timeout); err != nil {
		return nil, errors.Trace(err)
	}
	if report := view.Report; report != nil {
		if err = report.translateServiceResourceQuantity(); err != nil {
			return nil, errors.Trace(err)
		}
		if view.Ready != NodeOnline {
			err = report.resetNodeAppStats()
			if err != nil {
				return nil, errors.Trace(err)
			}
		}
		report.countInstanceNum()
		view.AppMode = report.calculateAppMode()
	}
	return view, nil
}

func (n *Node) compatibleSingleNode() error {
	edgeNodeName := ""
	nodeInfo, ok := n.Report["node"]
	if ok {
		nodeInfoView := map[string]*NodeInfo{}
		nodeInfoStr, err := json.Marshal(nodeInfo)
		if err != nil {
			return errors.Trace(err)
		}
		err = json.Unmarshal(nodeInfoStr, &nodeInfoView)
		if err != nil {
			log.L().Warn("failed to translate node to cluster node view", log.Any("node", n.Name))
			singleNodeInfo := new(NodeInfo)
			err = json.Unmarshal(nodeInfoStr, singleNodeInfo)
			if err != nil {
				return errors.Trace(err)
			}
			edgeNodeName = singleNodeInfo.Hostname
			singleNodeInfo.Role = "master"
			n.Report["node"] = map[string]*NodeInfo{
				edgeNodeName: singleNodeInfo,
			}

			nodeStats, ok := n.Report["nodestats"]
			if ok {
				nodeStatsStr, err := json.Marshal(nodeStats)
				if err != nil {
					return errors.Trace(err)
				}
				singleNodeStats := new(NodeStats)
				err = json.Unmarshal(nodeStatsStr, singleNodeStats)
				if err != nil {
					return errors.Trace(err)
				}
				n.Report["nodestats"] = map[string]*NodeStats{
					edgeNodeName: singleNodeStats,
				}
			}
		}
	}
	return nil
}

func (view *NodeView) populateNodeStats(timeout time.Duration) (err error) {
	if view.Report == nil {
		view.Ready = NodeUninstall
		return nil
	}

	if stats := view.Report.NodeStats; stats != nil {
		for _, s := range stats {
			if s.Percent == nil {
				s.Percent = map[string]string{}
			}
			if s.Capacity == nil {
				s.Capacity = map[string]string{}
			}
			if s.Usage == nil {
				s.Usage = map[string]string{}
			}
			memory := string(coreV1.ResourceMemory)
			if s.Percent[memory], err = s.processResourcePercent(s, memory, populateMemoryResource); err != nil {
				return errors.Trace(err)
			}

			cpu := string(coreV1.ResourceCPU)
			if s.Percent[cpu], err = s.processResourcePercent(s, cpu, populateCPUResource); err != nil {
				return errors.Trace(err)
			}
			if extension := s.Extension; extension != nil {
				populateGPUStats(s, extension)
				populateDiskNetStats(s, extension)
			}
		}
	}

	if view.Report.Time == nil {
		view.Ready = NodeUninstall
		return nil
	}
	if time.Now().UTC().Before(view.Report.Time.Add(timeout)) {
		view.Ready = NodeOnline
	} else {
		view.Ready = NodeOffline
	}

	return
}

func IsLegalAcceleratorType(accelerator string) bool {
	if _, ok := acceleratorMap[accelerator]; ok {
		return true
	}
	return false
}

func populateGPUStats(s *NodeStats, extension interface{}) {
	stats, ok := extension.(map[string]interface{})
	if !ok {
		return
	}
	if val, ok := stats[KeyGPUUsedMemory]; ok {
		used, _ := val.(float64)
		s.Usage[ResourceGPU] = strconv.FormatFloat(used, 'f', -1, 64)
	}
	if val, ok := stats[KeyGPUTotalMemory]; ok {
		total, _ := val.(float64)
		s.Capacity[ResourceGPU] = strconv.FormatFloat(total, 'f', -1, 64)
	}
	if val, ok := stats[KeyGPUPercent]; ok {
		percent, _ := val.(float64)
		s.Percent[ResourceGPU] = strconv.FormatFloat(percent, 'f', -1, 64)
	}
}

func populateDiskNetStats(s *NodeStats, extension interface{}) {
	stats, _ := extension.(map[string]interface{})
	if val, ok := stats[KeyDiskUsed]; ok {
		used, _ := val.(float64)
		s.Usage[ResourceDisk] = strconv.FormatFloat(used, 'f', -1, 64)
	}
	if val, ok := stats[KeyDiskTotal]; ok {
		total, _ := val.(float64)
		s.Capacity[ResourceDisk] = strconv.FormatFloat(total, 'f', -1, 64)
	}
	if val, ok := stats[KeyDiskPercent]; ok {
		percent, _ := val.(float64)
		s.Percent[ResourceDisk] = strconv.FormatFloat(percent, 'f', -1, 64)
	}
	s.NetIO = map[string]string{}
	if val, ok := stats[KeyNetBytesSent]; ok {
		bytesSent, _ := val.(float64)
		s.NetIO[KeyNetBytesSent] = strconv.FormatFloat(bytesSent, 'f', -1, 64)
	}
	if val, ok := stats[KeyNetBytesRecv]; ok {
		bytesRecv, _ := val.(float64)
		s.NetIO[KeyNetBytesRecv] = strconv.FormatFloat(bytesRecv, 'f', -1, 64)
	}
	if val, ok := stats[KeyNetPacketsRecv]; ok {
		packetsRecv, _ := val.(float64)
		s.NetIO[KeyNetPacketsRecv] = strconv.FormatFloat(packetsRecv, 'f', -1, 64)
	}
	if val, ok := stats[KeyNetPacketsSent]; ok {
		packetsSent, _ := val.(float64)
		s.NetIO[KeyNetPacketsSent] = strconv.FormatFloat(packetsSent, 'f', -1, 64)
	}
}

func (s *NodeStats) processResourcePercent(status *NodeStats, resourceType string,
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
		if usage >= total {
			return "1", nil
		}
		return strconv.FormatFloat(float64(usage)/float64(total), 'f', -1, 64), nil
	}
	return "0", nil
}

func (view *ReportView) calculateAppMode() string {
	for _, node := range view.Node {
		if node.ContainerRuntime == "" && node.MachineID == "" {
			return context.RunModeNative
		}
	}
	return context.RunModeKube
}

func (view *ReportView) countInstanceNum() {
	nums := map[string]int{}
	if view.AppStats != nil {
		for _, stat := range view.AppStats {
			for _, ins := range stat.InstanceStats {
				if _, ok := nums[ins.NodeName]; ok {
					nums[ins.NodeName] += 1
				} else {
					nums[ins.NodeName] = 1
				}
			}
		}
	}
	if view.SysAppStats != nil {
		for _, stat := range view.SysAppStats {
			for _, ins := range stat.InstanceStats {
				if _, ok := nums[ins.NodeName]; ok {
					nums[ins.NodeName] += 1
				} else {
					nums[ins.NodeName] = 1
				}
			}
		}
	}
	view.NodeInsNum = nums
}

func (view *ReportView) translateServiceResourceQuantity() error {
	for idx := range view.SysAppStats {
		instances := view.SysAppStats[idx].InstanceStats
		if instances == nil {
			continue
		}
		for _, v := range instances {
			if err := v.translateResourceQuantity(); err != nil {
				return errors.Trace(err)
			}
		}
	}

	for idx := range view.AppStats {
		services := view.AppStats[idx].InstanceStats
		if services == nil {
			continue
		}
		for _, v := range services {
			if err := v.translateResourceQuantity(); err != nil {
				return errors.Trace(err)
			}
		}
	}

	return nil
}

func (view *ReportView) resetNodeAppStats() error {
	for idx := range view.SysAppStats {
		view.SysAppStats[idx].Status = ""
	}

	for idx := range view.AppStats {
		view.AppStats[idx].Status = ""
	}
	return nil
}

func (s *InstanceStats) translateResourceQuantity() error {
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

func getAppStats(statsType string, data map[string]interface{}) []AppStats {
	if data == nil {
		return nil
	}
	apps, ok := data[statsType]
	if !ok || apps == nil {
		return nil
	}
	if res, ok := apps.([]AppStats); ok {
		return res
	}
	var appstats []AppStats
	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{DecodeHook: timeHookFunc, Result: &appstats})
	if err := decoder.Decode(apps); err == nil {
		return appstats
	}
	return nil
}

func timeHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t != reflect.TypeOf(time.Time{}) {
		return data, nil
	}
	switch f.Kind() {
	case reflect.String:
		return time.Parse(time.RFC3339, data.(string))
	default:
		return data, nil
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

func diff(desired, reported map[string]interface{}, cleanNil bool) (map[string]interface{}, error) {
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
	if cleanNil {
		clean(delta)
	}
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
