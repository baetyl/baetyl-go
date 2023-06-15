package v1

import (
	"time"

	v1 "k8s.io/api/core/v1"
)

const (
	AppTypeContainer = "container"
	AppTypeFunction  = "function"
	AppTypeAndroid   = "android"

	WorkloadDeployment  = "deployment"
	WorkloadStatefulSet = "statefulset"
	WorkloadDaemonSet   = "daemonset"
	WorkloadJob         = "job"
)

type CronStatusCode int

const (
	CronNotSet   CronStatusCode = 0
	CronWait     CronStatusCode = 1
	CronFinished CronStatusCode = 2
)

// Application application info
type Application struct {
	Name              string            `json:"name,omitempty" yaml:"name,omitempty" binding:"res_name"`
	Type              string            `json:"type,omitempty" yaml:"type,omitempty" default:"container"` // container | function
	Mode              string            `json:"mode,omitempty" yaml:"mode,omitempty" default:"kube"`      // kube | native
	Labels            map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Namespace         string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	CreationTimestamp time.Time         `json:"createTime,omitempty" yaml:"createTime,omitempty"`
	Version           string            `json:"version,omitempty" yaml:"version,omitempty"`
	Selector          string            `json:"selector,omitempty" yaml:"selector,omitempty"`
	NodeSelector      string            `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	Description       string            `json:"description,omitempty" yaml:"description,omitempty"`
	System            bool              `json:"system,omitempty" yaml:"system,omitempty"`
	CronStatus        CronStatusCode    `json:"cronStatus,omitempty" yaml:"cronStatus,omitempty" default:"0"`
	UpdateTime        time.Time         `json:"updateTime,omitempty" yaml:"updateTime,omitempty"`
	CronTime          time.Time         `json:"cronTime,omitempty" yaml:"cronTime,omitempty"`
	Replica           int               `json:"replica" yaml:"replica"`
	Workload          string            `json:"workload,omitempty" yaml:"workload,omitempty"` // deployment | daemonset | statefulset | job
	JobConfig         *AppJobConfig     `json:"jobConfig,omitempty" yaml:"jobConfig,omitempty"`
	Ota               OtaInfo           `json:"ota,omitempty" yaml:"ota,omitempty"`
	AutoScaleCfg      *AutoScaleCfg     `json:"autoScaleCfg,omitempty" yaml:"autoScaleCfg,omitempty"`
	// Reference to k8s PodSpec
	InitServices []Service    `json:"initServices,omitempty" yaml:"initServices,omitempty" binding:"dive"`
	Services     []Service    `json:"services,omitempty" yaml:"services,omitempty" binding:"dive"`
	Volumes      []Volume     `json:"volumes,omitempty" yaml:"volumes,omitempty" binding:"dive"`
	HostNetwork  bool         `json:"hostNetwork,omitempty" yaml:"hostNetwork,omitempty"` // specifies host network mode of service
	DNSPolicy    v1.DNSPolicy `json:"dnsPolicy,omitempty" yaml:"dnsPolicy,omitempty" default:"ClusterFirst"`
}

// Service service config1ma1
type Service struct {
	// specifies the unique name of the service
	Name string `json:"name,omitempty" yaml:"name,omitempty" binding:"required,svc_name"`
	// specifies the hostname of the service
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	// specifies the image of the service, usually using the Docker image name
	Image string `json:"image,omitempty" yaml:"image,omitempty"`
	// specifies the storage volumes that the service needs, map the storage volume to the directory in the container
	VolumeMounts []VolumeMount `json:"volumeMounts,omitempty" yaml:"volumeMounts,omitempty"`
	// specifies the port bindings which exposed by the service, only for Docker container mode
	Ports []ContainerPort `json:"ports,omitempty" yaml:"ports,omitempty"`
	// specifies the device bindings which used by the service, only for Docker container mode
	Devices []Device `json:"devices,omitempty" yaml:"devices,omitempty"`
	// specifies the startup arguments of the service program, but does not include `arg[0]`
	Args []string `json:"args,omitempty" yaml:"args,omitempty"`
	// specifies the commands of the service program
	Command []string `json:"command,omitempty" yaml:"command,omitempty"`
	// specifies the work dir in container
	WorkingDir string `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
	// specifies the environment variable of the service program
	Env []Environment `json:"env,omitempty" yaml:"env,omitempty"`
	// specifies resource limits for a single instance of the service,  only for Docker container mode
	Resources *Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
	// specifies runtime to use, only for Docker container mode
	Runtime string `json:"runtime,omitempty" yaml:"runtime,omitempty"`
	// specifies the security context of service
	SecurityContext *SecurityContext `json:"security,omitempty" yaml:"security,omitempty"`
	// specifies function config of service
	FunctionConfig *ServiceFunctionConfig `json:"functionConfig,omitempty" yaml:"functionConfig,omitempty"`
	// specifies functions of service
	Functions []ServiceFunction `json:"functions,omitempty" yaml:"functions,omitempty"`
	// labels
	// Deprecated: Use Application.Labels instead
	// Change from one workload for each service to one workload for one app, and each service as a container
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	// specifies host network mode of service
	// Deprecated: Use Application.HostNetwork instead
	// Change from one workload for each service to one workload for one app, and each service as a container
	HostNetwork bool `json:"hostNetwork,omitempty" yaml:"hostNetwork,omitempty"`
	// specifies the number of instances started
	// Deprecated: Use Application.Replica instead
	// Change from one workload for each service to one workload for one app, and each service as a container
	Replica int `json:"replica,omitempty" yaml:"replica,omitempty"`
	// specifies job config of service
	// Deprecated: Use Application.JobConfig instead.
	// Change from one workload for each service to one workload for one app, and each service as a container
	JobConfig *ServiceJobConfig `json:"jobConfig,omitempty" yaml:"jobConfig,omitempty"`
	// specifies type of service. deployment, daemonset, statefulset
	// Deprecated: Use Application.Workload instead.
	// Change from one workload for each service to one workload for one app, and each service as a container
	Type string `json:"type,omitempty" yaml:"type,omitempty" default:"deployment"`
	//Probe describes a health check to be performed against a container to
	//determine whether it is alive or ready to receive traffic.
	LivenessProbe *Probe `json:"livenessProbe,omitempty" yaml:"livenessProbe,omitempty"`
	// Image pull policy.
	// One of Always, Never, IfNotPresent.
	ImagePullPolicy PullPolicy `json:"imagePullPolicy,omitempty" yaml:"imagePullPolicy" `
}

type PullPolicy string

type Probe struct {
	// The action taken to determine the health of a container
	ProbeHandler `json:",inline"`
	// Number of seconds after the container has started before liveness probes are initiated.
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty"`
	// Number of seconds after which the probe times out.
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty"`
	// How often (in seconds) to perform the probe.
	// Default to 10 seconds. Minimum value is 1.
	PeriodSeconds int32 `json:"periodSeconds,omitempty"`
	// Minimum consecutive successes for the probe to be considered successful after having failed.
	// Defaults to 1. Must be 1 for liveness and startup. Minimum value is 1.
	SuccessThreshold int32 `json:"successThreshold,omitempty"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	// Defaults to 3. Minimum value is 1.
	FailureThreshold int32 `json:"failureThreshold,omitempty"`
}

type ProbeHandler struct {
	// HTTPGet specifies the http request to perform.
	HTTPGet *HTTPGetAction `json:"httpGet,omitempty"`
	// Exec specifies the action to take.
	Exec *ExecAction `json:"exec,omitempty"`
	// TCPSocket specifies an action involving a TCP port.
	TCPSocket *TCPSocketAction `json:"tcpSocket,omitempty"`
}

type TCPSocketAction struct {
	// Number or name of the port to access on the container.
	// Number must be in the range 1 to 65535.
	// Name must be an IANA_SVC_NAME.
	Port IntOrString `json:"port,omitempty"`
	// Host name to connect to, defaults to the pod IP.
	Host string `json:"host,omitempty"`
}

type ExecAction struct {
	// Command is the command line to execute inside the container, the working directory for the
	// command  is root ('/') in the container's filesystem. The command is simply exec'd, it is
	// not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use
	// a shell, you need to explicitly call out to that shell.
	// Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
	Command []string `json:"command,omitempty"`
}

type HTTPGetAction struct {
	// Path to access on the HTTP server.
	Path string `json:"path,omitempty"`
	// Name or number of the port to access on the container.
	// Number must be in the range 1 to 65535.
	// Name must be an IANA_SVC_NAME.
	Port IntOrString `json:"port,omitempty"`
	// Host name to connect to, defaults to the pod IP. You probably want to set
	Host string `json:"host,omitempty"`
	// Scheme to use for connecting to the host.
	Scheme URIScheme `json:"scheme,omitempty" default:"HTTP"`
	// Custom headers to set in the request. HTTP allows repeated headers.
	HTTPHeaders []HTTPHeader `json:"httpHeaders,omitempty"`
}

// URIScheme identifies the scheme used for connection to a host for Get actions
type URIScheme string

const (
	// URISchemeHTTP means that the scheme used will be http://
	URISchemeHTTP URIScheme = "HTTP"
	// URISchemeHTTPS means that the scheme used will be https://
	URISchemeHTTPS URIScheme = "HTTPS"
)

type HTTPHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type IntOrString struct {
	Type   Type   `json:"type"`
	IntVal int32  `json:"intVal"`
	StrVal string `json:"strVal"`
}

type Type int64

const (
	Int Type = iota
	String
)

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
	// specifies type of ingress methods for a service.
	// Default ClusterIP, can choose NodePort.
	ServiceType string `json:"serviceType,omitempty" yaml:"serviceType,omitempty" default:"ClusterIP"`
	NodePort    int32  `json:"nodePort,omitempty" yaml:"nodePort,omitempty"`
}

// Volume volume config
type Volume struct {
	// specified name of the volume
	Name string `json:"name,omitempty" yaml:"name,omitempty" binding:"required,res_name"`
	// specified driver for the storage volume
	VolumeSource `json:",inline" yaml:",inline"`
}

// VolumeSource volume source, include empty directory, host path, config and secret
type VolumeSource struct {
	HostPath *HostPathVolumeSource `json:"hostPath,omitempty" yaml:"hostPath,omitempty"`
	Config   *ObjectReference      `json:"config,omitempty" yaml:"config,omitempty"`
	Secret   *ObjectReference      `json:"secret,omitempty" yaml:"secret,omitempty"`
	EmptyDir *EmptyDirVolumeSource `json:"emptyDir,omitempty" yaml:"emptyDir,omitempty"`
}

type EmptyDirVolumeSource struct {
	Medium    string `json:"medium,omitempty" yaml:"medium,omitempty"`
	SizeLimit string `json:"sizeLimit,omitempty" yaml:"sizeLimit,omitempty"`
}

// HostPathVolumeSource volume source of host path
type HostPathVolumeSource struct {
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
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
	// specifies sub mount path
	SubPath string `json:"subPath,omitempty" yaml:"subPath,omitempty"`
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
	Name    string `json:"name,omitempty" yaml:"name,omitempty" binding:"res_name"`
	Runtime string `json:"runtime,omitempty" yaml:"runtime,omitempty"`
}

type ServiceFunction struct {
	Name    string `json:"name,omitempty" yaml:"name,omitempty"`
	Handler string `json:"handler,omitempty" yaml:"handler,omitempty"`
	CodeDir string `json:"codedir,omitempty" yaml:"codedir,omitempty"`
}

type AppJobConfig struct {
	Completions   int    `json:"completions" yaml:"completions"`
	Parallelism   int    `json:"parallelism,omitempty" yaml:"parallelism,omitempty"`
	BackoffLimit  int    `json:"backoffLimit,omitempty" yaml:"backoffLimit,omitempty"`
	RestartPolicy string `json:"restartPolicy,omitempty" yaml:"restartPolicy,omitempty" default:"Never"`
}

type AutoScaleCfg struct {
	MinReplicas int          `json:"minReplicas,omitempty" yaml:"minReplicas,omitempty"`
	MaxReplicas int          `json:"maxReplicas,omitempty" yaml:"maxReplicas,omitempty"`
	Metrics     []MetricSpec `json:"metrics,omitempty" yaml:"metrics,omitempty"`
}

type MetricSpec struct {
	Type     string          `json:"type,omitempty" yaml:"type,omitempty" default:"Resource"`
	Resource *ResourceMetric `json:"resource,omitempty" yaml:"resource,omitempty"`
}

type ResourceMetric struct {
	// name is the name of the resource, cpu / memory
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// type represents whether the metric type is Utilization, Value, or AverageValue
	TargetType         string `json:"targetType,omitempty" yaml:"targetType,omitempty"`
	Value              string `json:"value,omitempty" yaml:"value,omitempty"`
	AverageValue       string `json:"averageValue,omitempty" yaml:"averageValue,omitempty"`
	AverageUtilization int    `json:"averageUtilization,omitempty" yaml:"averageUtilization,omitempty"`
}

// Deprecated: Use AppJobConfig instead.
// Change from one workload for each service to one workload for one app, and each service as a container
type ServiceJobConfig struct {
	Completions   int    `json:"completions,omitempty" yaml:"completions,omitempty"`
	Parallelism   int    `json:"parallelism,omitempty" yaml:"parallelism,omitempty"`
	BackoffLimit  int    `json:"backoffLimit,omitempty" yaml:"backoffLimit,omitempty"`
	RestartPolicy string `json:"restartPolicy,omitempty" yaml:"restartPolicy,omitempty" default:"Never"`
}
