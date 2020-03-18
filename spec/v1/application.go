package v1

import (
	"time"
	)

type Application struct {
	Name              string                      `json:"name,omitempty"`
	Labels            map[string]string           `json:"labels,omitempty"`
	Namespace         string                      `json:"namespace,omitempty"`
	CreationTimestamp time.Time                   `json:"creationTimestamp,omitempty"`
	Version           string                      `json:"version,omitempty"`
	Selector          string                      `json:"selector"`
	Services          []Service                   `json:"services,omitempty"`
	Volumes           []Volume                    `json:"volumes,omitempty"`
	Registries        map[string]*ObjectReference `json:"registries,omitempty"`
	Description       string                      `json:"description,omitempty"`
}

type Service struct {
	// specifies the unique name of the service
	Name string `json:"name,omitempty" binding:"required"`
	// specifies the hostname of the service
	Hostname string `json:"hostname,omitempty"`
	// specifies the image of the service, usually using the Docker image name
	Image string `json:"image,omitempty" binding:"required"`
	// specifies the number of instances started
	Replica int `json:"replica,omitempty" binding:"required"`
	// specifies the storage volumes that the service needs, map the storage volume to the directory in the container
	VolumeMounts []VolumeMount `json:"volumeMounts,omitempty"`
	// specifies the port bindings which exposed by the service, only for Docker container mode
	Ports []ContainerPort `json:"ports,omitempty"`
	// specifies the device bindings which used by the service, only for Docker container mode
	VolumeDevices []VolumeDevice `json:"devices,omitempty"`
	// specifies the startup arguments of the service program, but does not include `arg[0]`
	Args []string `json:"args,omitempty"`
	// specifies the environment variable of the service program
	Env []Environment `json:"env,omitempty"`
	// specifies the restart policy of the instance of the service
	Restart *RestartPolicyInfo `json:"restart,omitempty"`
	// specifies resource limits for a single instance of the service,  only for Docker container mode
	Resources *Resources `json:"resources,omitempty"`
	// specifies runtime to use, only for Docker container mode
	Runtime string `json:"runtime,omitempty"`
	// labels
	Labels map[string]string `json:"labels,omitempty"`
}

type Environment struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type VolumeDevice struct {
	DevicePath  string `json:"devicePath,omitempty"`
	Policy      string `json:"policy,omitempty"`
	Description string `json:"description,omitempty"`
}

// ContainerPort port map configuration
type ContainerPort struct {
	HostPort      int32  `json:"hostPort,omitempty"`
	ContainerPort int32  `json:"containerPort,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
	HostIP        string `json:"hostIP,omitempty"`
}

// Volume volume configuration of compose
type Volume struct {
	// specified name of the volume
	Name string `json:"name,omitempty" binding:"required"`
	// specified driver for the storage volume
	VolumeSource `json:",inline"`
}

// VolumeSource volume source, include empty directory, host path, config
type VolumeSource struct {
	HostPath *HostPathVolumeSource `json:"hostPath,omitempty"`
	Config   *ObjectReference      `json:"config,omitempty"`
	Secret   *ObjectReference      `json:"secret,omitempty"`
}

type HostPathVolumeSource struct {
	Path string `json:"path,omitempty"`
}

type ObjectReference struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// ServiceVolume specific volume configuration of service
type VolumeMount struct {
	// specifies name of volume
	Name string `json:"name,omitempty"`
	// specifies mount path of volume
	MountPath string `json:"mountPath,omitempty"`
	// specifies if the volume is read-only
	ReadOnly bool `json:"readOnly,omitempty"`
}

// RestartPolicyInfo holds the policy of a module
type RestartPolicyInfo struct {
	Retry   *Retry       `json:"retry,omitempty"`
	Policy  string       `json:"policy,omitempty"`
	Backoff *BackoffInfo `json:"backoff,omitempty"`
}

type Retry struct {
	Max int `json:"max,omitempty"`
}

// BackoffInfo holds backoff value
type BackoffInfo struct {
	Min    string  `json:"min,omitempty" validate:"duration"`
	Max    string  `json:"max,omitempty" validate:"duration"`
	Factor float64 `json:"factor,omitempty"`
}

// Resources resources config
type Resources struct {
	CPU    *CPU    `json:"cpu,omitempty"`
	Pids   *Pids   `json:"pids,omitempty"`
	Memory *Memory `json:"memory,omitempty"`
}

// CPU cpu config
type CPU struct {
	Cpus    float64 `json:"cpus,omitempty"`
	SetCPUs string  `json:"setcpus,omitempty" validate:"setcpus"`
}

// Pids pids config
type Pids struct {
	Limit int64 `json:"limit,omitempty"`
}

// Memory memory config
type Memory struct {
	Limit string `json:"limit,omitempty" validate:"mem"`
	Swap  string `json:"swap,omitempty" validate:"mem"`
}
