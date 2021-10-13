package v1

import "time"

const (
	AppTypeContainer = "container"
	AppTypeFunction  = "function"

	ServiceTypeDeployment  = "deployment"
	ServiceTypeStatefulSet = "statefulset"
	ServiceTypeDaemonSet   = "daemonset"
	ServiceTypeJob         = "job"
)

type CronStatusCode int

const (
	CronNotSet   CronStatusCode = 0
	CronWait     CronStatusCode = 1
	CronFinished CronStatusCode = 2
)

// Application application info
type Application struct {
	Name              string            `json:"name,omitempty" yaml:"name,omitempty" validate:"resourceName"`
	Type              string            `json:"type,omitempty" yaml:"type,omitempty" default:"container"`
	Mode              string            `json:"mode,omitempty" yaml:"mode,omitempty" default:"kube"`
	Labels            map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Namespace         string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	CreationTimestamp time.Time         `json:"createTime,omitempty" yaml:"createTime,omitempty"`
	Version           string            `json:"version,omitempty" yaml:"version,omitempty"`
	Selector          string            `json:"selector,omitempty" yaml:"selector,omitempty"`
	NodeSelector      string            `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	Services          []Service         `json:"services,omitempty" yaml:"services,omitempty" validate:"dive"`
	Volumes           []Volume          `json:"volumes,omitempty" yaml:"volumes,omitempty" validate:"dive"`
	Description       string            `json:"description,omitempty" yaml:"description,omitempty"`
	System            bool              `json:"system,omitempty" yaml:"system,omitempty"`
	CronStatus        CronStatusCode    `json:"cronStatus,omitempty" yaml:"cronStatus,omitempty" default:"0"`
	UpdateTime        time.Time         `json:"updateTime,omitempty" yaml:"updateTime,omitempty"`
	CronTime          time.Time         `json:"cronTime,omitempty" yaml:"cronTime,omitempty"`
}

// Service service config1ma1
type Service struct {
	// specifies the unique name of the service
	Name string `json:"name,omitempty" yaml:"name,omitempty" binding:"required" validate:"serviceName"`
	// specifies the hostname of the service
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	// specifies the image of the service, usually using the Docker image name
	Image string `json:"image,omitempty" yaml:"image,omitempty"`
	// specifies the number of instances started
	Replica int `json:"replica,omitempty" yaml:"replica,omitempty" binding:"required" default:"1"`
	// specifies the storage volumes that the service needs, map the storage volume to the directory in the container
	VolumeMounts []VolumeMount `json:"volumeMounts,omitempty" yaml:"volumeMounts,omitempty"`
	// specifies the port bindings which exposed by the service, only for Docker container mode
	Ports []ContainerPort `json:"ports,omitempty" yaml:"ports,omitempty"`
	// specifies the device bindings which used by the service, only for Docker container mode
	Devices []Device `json:"devices,omitempty" yaml:"devices,omitempty"`
	// specifies the startup arguments of the service program, but does not include `arg[0]`
	Args []string `json:"args,omitempty" yaml:"args,omitempty"`
	// specifies the environment variable of the service program
	Env []Environment `json:"env,omitempty" yaml:"env,omitempty"`
	// specifies resource limits for a single instance of the service,  only for Docker container mode
	Resources *Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
	// specifies runtime to use, only for Docker container mode
	Runtime string `json:"runtime,omitempty" yaml:"runtime,omitempty"`
	// labels
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	// specifies the security context of service
	SecurityContext *SecurityContext `json:"security,omitempty" yaml:"security,omitempty"`
	// specifies host network mode of service
	HostNetwork bool `json:"hostNetwork,omitempty" yaml:"hostNetwork,omitempty"`
	// specifies function config of service
	FunctionConfig *ServiceFunctionConfig `json:"functionConfig,omitempty" yaml:"functionConfig,omitempty"`
	// specifies functions of service
	Functions []ServiceFunction `json:"functions,omitempty" yaml:"functions,omitempty"`
	// specifies job config of service
	JobConfig *ServiceJobConfig `json:"jobConfig,omitempty" yaml:"jobConfig,omitempty"`
	// specifies type of service. deployment, daemonset, statefulset
	Type string `json:"type,omitempty" yaml:"type,omitempty" default:"deployment"`
}

type SecurityContext struct {
	Privileged bool `json:"privileged,omitempty" yaml:"privileged,omitempty"`
}

// Environment environment config
type Environment struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

// VolumeDevice device volume config
type Device struct {
	DevicePath  string `json:"devicePath,omitempty" yaml:"devicePath,omitempty"`
	Policy      string `json:"policy,omitempty" yaml:"policy,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// ContainerPort port config in container
type ContainerPort struct {
	HostPort      int32  `json:"hostPort,omitempty" yaml:"hostPort,omitempty"`
	ContainerPort int32  `json:"containerPort,omitempty" yaml:"containerPort,omitempty"`
	Protocol      string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	HostIP        string `json:"hostIP,omitempty" yaml:"hostIP,omitempty"`
}

// Volume volume config
type Volume struct {
	// specified name of the volume
	Name string `json:"name,omitempty" yaml:"name,omitempty" binding:"required" validate:"resourceName"`
	// specified driver for the storage volume
	VolumeSource `json:",inline" yaml:",inline"`
}

// VolumeSource volume source, include empty directory, host path, config and secret
type VolumeSource struct {
	HostPath *HostPathVolumeSource `json:"hostPath,omitempty" yaml:"hostPath,omitempty"`
	Config   *ObjectReference      `json:"config,omitempty" yaml:"config,omitempty"`
	Secret   *ObjectReference      `json:"secret,omitempty" yaml:"secret,omitempty"`
}

// HostPathVolumeSource volume source of host path
type HostPathVolumeSource struct {
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

// ObjectReference object reference to config or secret
type ObjectReference struct {
	Name    string `json:"name,omitempty" yaml:"name,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

// VolumeMount volume mount config
type VolumeMount struct {
	// specifies name of volume
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// specifies mount path of volume
	MountPath string `json:"mountPath,omitempty" yaml:"mountPath,omitempty"`
	// specifies if the volume is read-only
	ReadOnly bool `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	// specifies if the volumeMount is immutable
	Immutable bool `json:"immutable,omitempty" yaml:"immutable,omitempty"`
	// specifies if clean the volume automatically
	AutoClean bool `json:"autoClean,omitempty" yaml:"autoClean,omitempty"`
}

// Retry retry config
type Retry struct {
	Max int `json:"max,omitempty" yaml:"max,omitempty"`
}

// Resources resources config
type Resources struct {
	Limits   map[string]string `json:"limits,omitempty" yaml:"limits,omitempty"`
	Requests map[string]string `json:"requests,omitempty" yaml:"requests,omitempty"`
}

type ServiceFunctionConfig struct {
	Name    string `json:"name,omitempty" yaml:"name,omitempty" validate:"resourceName"`
	Runtime string `json:"runtime,omitempty" yaml:"runtime,omitempty"`
}

type ServiceFunction struct {
	Name    string `json:"name,omitempty" yaml:"name,omitempty"`
	Handler string `json:"handler,omitempty" yaml:"handler,omitempty"`
	CodeDir string `json:"codedir,omitempty" yaml:"codedir,omitempty"`
}

type ServiceJobConfig struct {
	Completions   int    `json:"completions,omitempty" yaml:"completions,omitempty"`
	Parallelism   int    `json:"parallelism,omitempty" yaml:"parallelism,omitempty"`
	BackoffLimit  int    `json:"backoffLimit,omitempty" yaml:"backoffLimit,omitempty"`
	RestartPolicy string `json:"restartPolicy,omitempty" yaml:"restartPolicy,omitempty" default:"Never"`
}
