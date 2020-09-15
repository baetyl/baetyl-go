package native

import (
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
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
		return 0, errors.New("ports of service are empty in services mapping file")
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

func (s *ServiceMapping) load() error {
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

func (s *ServiceMapping) save() error {
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

func (s *ServiceMapping) AddServicePorts(serviceName string, ports []int) error {
	s.Lock()
	defer s.Unlock()

	s.Services[serviceName] = ServiceMappingInfo{
		Ports: PortsInfo{
			Items: ports,
		},
	}
	err := s.save()
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (s *ServiceMapping) DeleteServicePorts(serviceName string) {
	s.Lock()
	defer s.Unlock()

	delete(s.Services, serviceName)
}

func (s *ServiceMapping) GetServiceNextPort(serviceName string) (int, error) {
	s.Lock()
	defer s.Unlock()

	serviceInfo, ok := s.Services[serviceName]
	if !ok {
		return 0, errors.New("no such service in services mapping file")
	}

	if len(serviceInfo.Ports.Items) == 0 {
		return 0, errors.New("no ports info in services mapping file")
	}

	port, err := serviceInfo.Ports.Next()
	if err != nil {
		return 0, err
	}
	return port, nil
}

func (s *ServiceMapping) WatchFile(errChan chan<- error, logger *log.Logger) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Trace(err)
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					errChan <- errors.New("error when wait on the events channel")
					return
				}
				logger.Debug("received a file event", log.Any("eventName", event.Name), log.Any("eventOp", event.Op))

				if event.Op&fsnotify.Write != fsnotify.Write {
					continue
				}

				logger.Debug("load services mapping file again", log.Error(err))
				err := s.load()
				if err != nil {
					logger.Warn("load services mapping file failed", log.Error(err))
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					errChan <- errors.New("error when wait on the events channel")
					return
				}
				errChan <- errors.Trace(err)
				return
			}
		}
	}()

	err = watcher.Add(ServiceMappingFile)
	if err != nil {
		return errors.Trace(err)
	}

	err = s.load()
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}
