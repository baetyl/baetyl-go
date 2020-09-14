package native

import (
	"io/ioutil"
	"os"
	"path"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/utils"
)

const (
	ServiceMappingFile = "var/lib/baetyl/run/services.yml"
)

type ServiceMapping struct {
	Services map[string]ServiceMappingInfo `yaml:"services,omitempty"`
	sync.RWMutex
}

type ServiceMappingInfo struct {
	Ports PortsInfo `yaml:"ports,omitempty"`
}

type PortsInfo struct {
	Items  []int `yaml:"items,omitempty"`
	offset int
}

func (i *PortsInfo) Next() (int, error) {
	if len(i.Items) == 0 {
		return 0, errors.New("ports of service are empty in ports mapping file")
	}
	port := i.Items[i.offset]
	i.offset++
	i.offset = i.offset % len(i.Items)
	return port, nil
}

func NewServiceMapping() *ServiceMapping {
	return &ServiceMapping{
		Services: make(map[string]ServiceMappingInfo),
	}
}

func (s *ServiceMapping) Load() error {
	s.Lock()
	defer s.Unlock()

	if !utils.FileExists(ServiceMappingFile) {
		return errors.Errorf("services mapping file (%s) doesn't exist", ServiceMappingFile)
	}
	data, err := ioutil.ReadFile(ServiceMappingFile)
	if err != nil {
		return errors.Trace(err)
	}
	err = yaml.Unmarshal(data, s)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (s *ServiceMapping) Save() error {
	s.Lock()
	defer s.Unlock()

	data, err := yaml.Marshal(s)
	if err != nil {
		return errors.Trace(err)
	}
	if !utils.PathExists(path.Dir(ServiceMappingFile)) {
		err = os.MkdirAll(path.Dir(ServiceMappingFile), 0755)
		if err != nil {
			return errors.Trace(err)
		}
	}
	return ioutil.WriteFile(ServiceMappingFile, data, 0755)
}
