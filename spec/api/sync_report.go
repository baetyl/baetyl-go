package api

import "time"

// ReportRequest report request
type ReportRequest struct {
	Time     time.Time   `json:"time,omitempty"`
	NodeInfo NodeInfo    `json:"node,omitempty"`
	NodeStat NodeStatus  `json:"nodestats,omitempty"`
	AppInfos []AppInfo   `json:"apps,omitempty"`
	AppStats []AppStatus `json:"appstats,omitempty"`
}

// NodeInfo node info
type NodeInfo struct {
	Hostname         string `yaml:"hostname,omitempty" json:"hostname,omitempty"`
	Address          string `yaml:"address,omitempty" json:"address,omitempty"`
	Arch             string `yaml:"arch,omitempty" json:"arch,omitempty"`
	KernelVersion    string `yaml:"kernelVer,omitempty" json:"kernelVer,omitempty"`
	OS               string `yaml:"os,omitempty" json:"os,omitempty"`
	ContainerRuntime string `yaml:"containerRuntime,omitempty" json:"containerRuntime"`
	MachineID        string `yaml:"machineID,omitempty" json:"machineID"`
	BootID           string `yaml:"bootID,omitempty" json:"bootID"`
	SystemUUID       string `yaml:"systemUUID,omitempty" json:"systemUUID"`
	OSImage          string `yaml:"osImage,omitempty" json:"osImage"`
}

// NodeStatus node status
type NodeStatus struct {
	Usage    map[string]*ResourceInfo `yaml:"usage,omitempty" json:"usage,omitempty"`
	Capacity map[string]*ResourceInfo `yaml:"capacity,omitempty" json:"capacity,omitempty"`
}

// AppInfo app info
type AppInfo struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
}

// AppStatus app status
type AppStatus struct {
	AppInfo      `yaml:",inline" json:",inline"`
	Status       string                  `yaml:"status,omitempty" json:"status,omitempty"`
	Cause        string                  `yaml:"cause,omitempty" json:"cause,omitempty"`
	ServiceInfos map[string]*ServiceInfo `yaml:"services,omitempty" json:"services,omitempty"`
	VolumeInfos  map[string]*VolumeInfo  `yaml:"volumes,omitempty" json:"volumes,omitempty"`
}

// ReportResponse report response as desire
type ReportResponse struct {
	AppInfos []AppInfo `json:"apps,omitempty"`
}

// ResourceInfo resource info
type ResourceInfo struct {
	Name        string `yaml:"name,omitempty" json:"name,omitempty"`
	Value       string `yaml:"value,omitempty" json:"value,omitempty"`
	UsedPercent string `yaml:"usedPercent,omitempty" json:"usedPercent,omitempty"`
}

// ServiceInfo service info
type ServiceInfo struct {
	Name       string                   `yaml:"name,omitempty" json:"name,omitempty"`
	Container  Container                `yaml:"container,omitempty" json:"container,omitempty"`
	Usage      map[string]*ResourceInfo `yaml:"usage,omitempty" json:"usage,omitempty"`
	Status     string                   `yaml:"status,omitempty" json:"status,omitempty"`
	Cause      string                   `yaml:"cause,omitempty" json:"cause,omitempty"`
	CreateTime time.Time                `yaml:"createTime,omitempty" json:"createTime,omitempty"`
}

// Container container info
type Container struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	ID   string `yaml:"id,omitempty" json:"id,omitempty"`
}

// VolumeInfo volume info
type VolumeInfo struct {
	Name    string `yaml:"name,omitempty" json:"name,omitempty"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
}
