package v1

import "time"

type Status string

const (
	Pending Status = "Pending"
	Failed  Status = "Failed"
	Running Status = "Running"
	Unknown Status = "Unknown"
)

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

// NodeStats node statistics
type NodeStats struct {
	Usage    map[string]string `yaml:"usage,omitempty" json:"usage,omitempty"`
	Capacity map[string]string `yaml:"capacity,omitempty" json:"capacity,omitempty"`
	Percent  map[string]string `yaml:"percent,omitempty" json:"percent,omitempty"`
}

// AppInfo app info
type AppInfo struct {
	Name    string `yaml:"name,omitempty" json:"name,omitempty"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
}

// AppStats app statistics
type AppStats struct {
	AppInfo       `yaml:",inline" json:",inline"`
	Status        Status                   `yaml:"status,omitempty" json:"status,omitempty"`
	InstanceStats map[string]InstanceStats `yaml:"instances,omitempty" json:"instances,omitempty"`
}

type CoreInfo struct {
	GoVersion   string `yaml:"goVersion,omitempty" json:"goVersion,omitempty"`
	BinVersion  string `yaml:"binVersion,omitempty" json:"binVersion,omitempty"`
	GitRevision string `yaml:"gitRevision,omitempty" json:"gitRevision,omitempty"`
}

// InstanceStats instance stats
type InstanceStats struct {
	Name        string            `yaml:"name,omitempty" json:"name,omitempty"`
	ServiceName string            `yaml:"serviceName,omitempty" json:"serviceName"`
	Usage       map[string]string `yaml:"usage,omitempty" json:"usage,omitempty"`
	Status      Status            `yaml:"status,omitempty" json:"status,omitempty"`
	Cause       string            `yaml:"cause,omitempty" json:"cause,omitempty"`
	CreateTime  time.Time         `yaml:"createTime,omitempty" json:"createTime,omitempty"`
}
