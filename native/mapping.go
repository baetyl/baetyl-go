package native

import (
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/utils"
)

const (
	serviceMappingFile = "var/lib/baetyl/run/services.yml"
)

type ServiceMapping struct {
	Services map[string]ServiceMappingInfo `yaml:"services,omitempty"`
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
	if i.offset == len(i.Items) {
		i.offset = 0
	}
	return port, nil
}

func NewServiceMapping() (*ServiceMapping, error) {
	m := &ServiceMapping{
		Services: make(map[string]ServiceMappingInfo),
	}
	return m, nil
}

func (s *ServiceMapping) Load() error {
	if !utils.FileExists(serviceMappingFile) {
		return errors.Errorf("services mapping file (%s) doesn't exist", serviceMappingFile)
	}
	data, err := ioutil.ReadFile(serviceMappingFile)
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
	data, err := yaml.Marshal(s)
	if err != nil {
		return errors.Trace(err)
	}
	if !utils.PathExists(path.Dir(serviceMappingFile)) {
		err = os.MkdirAll(path.Dir(serviceMappingFile), 0755)
		if err != nil {
			return errors.Trace(err)
		}
	}
	return ioutil.WriteFile(serviceMappingFile, data, 0755)
}
