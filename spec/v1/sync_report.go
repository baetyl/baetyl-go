package v1

import "time"

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
	CPU    CPUStats    `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	Memory MemoryStats `yaml:"memory,omitempty" json:"memory,omitempty"`
}

type MemoryStats struct {
	Used        int64   `yaml:"used,omitempty" json:"used,omitempty"`
	Total       int64   `yaml:"total,omitempty" json:"total,omitempty"`
	UsedPercent float64 `yaml:"usedPercent,omitempty" json:"usedPercent,omitempty"`
}

type CPUStats struct {
	Used        float64 `yaml:"used,omitempty" json:"used,omitempty"`
	Total       float64 `yaml:"total,omitempty" json:"total,omitempty"`
	UsedPercent float64 `yaml:"usedPercent,omitempty" json:"usedPercent,omitempty"`
}

// AppInfo app info
type AppInfo struct {
	Name    string `yaml:"name,omitempty" json:"name,omitempty"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
}

// AppStatus app status
type AppStatus struct {
	AppInfo      `yaml:",inline" json:",inline"`
	Status       string                  `yaml:"status,omitempty" json:"status,omitempty"`
	Cause        string                  `yaml:"cause,omitempty" json:"cause,omitempty"`
	ServiceInfos map[string]*ServiceInfo `yaml:"services,omitempty" json:"services,omitempty"`
}

type CoreInfo struct {
	GoVersion   string `yaml:"goVersion,omitempty" json:"goVersion,omitempty"`
	BinVersion  string `yaml:"binVersion,omitempty" json:"binVersion,omitempty"`
	GitRevision string `yaml:"gitRevision,omitempty" json:"gitRevision,omitempty"`
}

// ServiceInfo service info
type ServiceInfo struct {
	Name       string      `yaml:"name,omitempty" json:"name,omitempty"`
	Container  Container   `yaml:"container,omitempty" json:"container,omitempty"`
	CPU        CPUStats    `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	Memory     MemoryStats `yaml:"memory,omitempty" json:"memory,omitempty"`
	Status     string      `yaml:"status,omitempty" json:"status,omitempty"`
	Cause      string      `yaml:"cause,omitempty" json:"cause,omitempty"`
	CreateTime time.Time   `yaml:"createTime,omitempty" json:"createTime,omitempty"`
}

// Container container info
type Container struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	ID   string `yaml:"id,omitempty" json:"id,omitempty"`
}
