package v1

import "time"

// Application application info
type Application struct {
	Name              string            `json:"name,omitempty" validate:"resourceName,nonBaetyl"`
	Labels            map[string]string `json:"labels,omitempty"`
	Namespace         string            `json:"namespace,omitempty"`
	CreationTimestamp time.Time         `json:"createTime,omitempty"`
	Version           string            `json:"version,omitempty"`
	Selector          string            `json:"selector,omitempty"`
	Services          []Service         `json:"services,omitempty"`
	Volumes           []Volume          `json:"volumes,omitempty"`
	Description       string            `json:"description,omitempty"`
}

// Service service config1ma1
type Service struct {
	// specifies the unique name of the service
	Name string `json:"name,omitempty" binding:"required"`
	// specifies the hostname of the service
	Hostname string `json:"hostname,omitempty"`
	// specifies the image of the service, usually using the Docker image name
	Image string `json:"image,omitempty" binding:"required"`
	// specifies the number of instances started
	Replica int `json:"replica,omitempty" binding:"required" default:"1"`
	// specifies the storage volumes that the service needs, map the storage volume to the directory in the container
	VolumeMounts []VolumeMount `json:"volumeMounts,omitempty"`
	// specifies the port bindings which exposed by the service, only for Docker container mode
	Ports []ContainerPort `json:"ports,omitempty"`
	// specifies the device bindings which used by the service, only for Docker container mode
	Devices []Device `json:"devices,omitempty"`
	// specifies the startup arguments of the service program, but does not include `arg[0]`
	Args []string `json:"args,omitempty"`
	// specifies the environment variable of the service program
	Env []Environment `json:"env,omitempty"`
	// specifies resource limits for a single instance of the service,  only for Docker container mode
	Resources *Resources `json:"resources,omitempty"`
	// specifies runtime to use, only for Docker container mode
	Runtime string `json:"runtime,omitempty"`
	// labels
	Labels map[string]string `json:"labels,omitempty"`
	// specifies the security context of service
	SecurityContext *SecurityContext `json:"security,omitempty"`
	// specifies host network mode of service
	HostNetwork bool `json:"hostNetwork,omitempty"`
}

type SecurityContext struct {
	Privileged bool `json:"privileged,omitempty"`
}

// Environment environment config
type Environment struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// VolumeDevice device volume config
type Device struct {
	DevicePath  string `json:"devicePath,omitempty"`
	Policy      string `json:"policy,omitempty"`
	Description string `json:"description,omitempty"`
}

// ContainerPort port config in container
type ContainerPort struct {
	HostPort      int32  `json:"hostPort,omitempty"`
	ContainerPort int32  `json:"containerPort,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
	HostIP        string `json:"hostIP,omitempty"`
}

// Volume volume config
type Volume struct {
	// specified name of the volume
	Name string `json:"name,omitempty" binding:"required"`
	// specified driver for the storage volume
	VolumeSource `json:",inline"`
}

// VolumeSource volume source, include empty directory, host path, config and secret
type VolumeSource struct {
	HostPath *HostPathVolumeSource `json:"hostPath,omitempty"`
	Config   *ObjectReference      `json:"config,omitempty"`
	Secret   *ObjectReference      `json:"secret,omitempty"`
}

// HostPathVolumeSource volume source of host path
type HostPathVolumeSource struct {
	Path string `json:"path,omitempty"`
}

// ObjectReference object reference to config or secret
type ObjectReference struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// VolumeMount volume mount config
type VolumeMount struct {
	// specifies name of volume
	Name string `json:"name,omitempty"`
	// specifies mount path of volume
	MountPath string `json:"mountPath,omitempty"`
	// specifies if the volume is read-only
	ReadOnly bool `json:"readOnly,omitempty"`
}

// Retry retry config
type Retry struct {
	Max int `json:"max,omitempty"`
}

// Resources resources config
type Resources struct {
	Limits   map[string]string `json:"limits,omitempty"`
	Requests map[string]string `json:"requests,omitempty"`
}
