package v1

type AppDeployment struct {
	App      Application           `yaml:"app,omitempty" json:"appConfig,omitempty"`
	Metadata map[string]MetaVolume `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

type MetaVolume struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
	Meta Meta   `yaml:"meta,omitempty" json:"meta,omitempty"`
}

type Meta struct {
	URL     string `yaml:"url,omitempty" json:"url,omitempty"`
	MD5     string `yaml:"md5,omitempty" json:"md5,omitempty"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
}

type Application struct {
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
	// Name name
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// specifies the service information of the application
	Services map[string]Service `yaml:"services,omitempty" json:"services,omitempty"`
	// specifies the storage volume information of the application
	Volumes map[string]AppVolume `yaml:"volumes,omitempty" json:"volumes,omitempty"`
}

type Service struct {
	// specifies the unique name of the service
	Name string `yaml:"container_name,omitempty" json:"container_name,omitempty"`
	// specifies the hostname of the service
	Hostname string `yaml:"hostname,omitempty" json:"hostname,omitempty"`
	// specifies the image of the service, usually using the Docker image name
	Image string `yaml:"image,omitempty" json:"image,omitempty"`
	// specifies the number of instances started
	Replica int `yaml:"replica,omitempty" json:"replica,omitempty"`
	// specifies the storage volumes that the service needs, map the storage volume to the directory in the container
	Volumes []*ServiceVolume `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	// specifies the port bindings which exposed by the service, only for Docker container mode
	PortMaps []string `yaml:"ports,omitempty" json:"ports,omitempty"`
	// specifies the device bindings which used by the service, only for Docker container mode
	Devices []string `yaml:"devices,omitempty" json:"devices,omitempty"`
	// specified other depended services
	DependsOn []string `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	// specifies the startup arguments of the service program, but does not include `arg[0]`
	Args []string `yaml:"command,omitempty" json:"command,omitempty"`
	// specifies the environment variable of the service program
	Environment map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
	// specifies the restart policy of the instance of the service
	Restart *RestartPolicyInfo `yaml:"restart,omitempty" json:"restart,omitempty"`
	// specifies resource limits for a single instance of the service,  only for Docker container mode
	Resources *Resources `yaml:"resources,omitempty" json:"resources,omitempty"`
	// specifies runtime to use, only for Docker container mode
	Runtime string `yaml:"runtime,omitempty" json:"runtime,omitempty"`
}

type AppVolume struct {
	// specified driver for the storage volume
	Driver string `yaml:"driver,omitempty" json:"driver,omitempty"`
	// specified driver options for the storage volume
	DriverOpts map[string]string `yaml:"driverOpts,omitempty" json:"driverOpts,omitempty"`
	// specified labels for the storage volume
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// ServiceVolume specific volume configuration of service
type ServiceVolume struct {
	// specifies type of volume
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
	// specifies source of volume
	Source string `yaml:"source,omitempty" json:"source,omitempty"`
	// specifies target of volume
	Target string `yaml:"target,omitempty" json:"target,omitempty"`
	// specifies if the volume is read-only
	ReadOnly bool `yaml:"read_only,omitempty" json:"read_only,omitempty"`
}

func (sv *ServiceVolume) MarshalYAML() (interface{}, error) {
	res := sv.Source + ":" + sv.Target
	if sv.ReadOnly {
		res += ":ro"
	}
	return res, nil
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
