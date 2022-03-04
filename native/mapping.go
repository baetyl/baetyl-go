package native

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/utils"
)

const (
	ServiceMappingFile = "run/services.yml"
)

type ServiceMapping struct {
	services map[string]*serviceMappingInfo
	tomb     utils.Tomb
	error    error
	sync.RWMutex
}

type serviceMappingInfo struct {
	PortInfo portsInfo `yaml:",inline"`
}

type portsInfo struct {
	Ports  []int `yaml:"ports,omitempty"`
	offset int
}

func (i *portsInfo) Next() (int, error) {
	if len(i.Ports) == 0 {
		return 0, errors.New("ports of service are empty in services mapping file")
	}
	port := i.Ports[i.offset]
	i.offset++
	i.offset = i.offset % len(i.Ports)
	return port, nil
}

func NewServiceMapping() (*ServiceMapping, error) {
	m := &ServiceMapping{
		services: make(map[string]*serviceMappingInfo),
	}
	hostPath, err := context.HostPathLib()
	if err != nil {
		return nil, err
	}
	if !utils.FileExists(filepath.Join(hostPath, ServiceMappingFile)) {
		err := m.save()
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (s *ServiceMapping) load() error {
	hostPath, err := context.HostPathLib()
	if err != nil {
		return err
	}
	mappingFile := filepath.Join(hostPath, ServiceMappingFile)
	if !utils.FileExists(mappingFile) {
		return errors.Errorf("services mapping file (%s) doesn't exist", mappingFile)
	}
	data, err := ioutil.ReadFile(mappingFile)
	if err != nil {
		return errors.Trace(err)
	}
	err = yaml.Unmarshal(data, &s.services)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (s *ServiceMapping) save() error {
	hostPath, err := context.HostPathLib()
	if err != nil {
		return err
	}
	mappingFile := filepath.Join(hostPath, ServiceMappingFile)
	data, err := yaml.Marshal(&s.services)
	if err != nil {
		return errors.Trace(err)
	}
	if !utils.PathExists(filepath.Dir(mappingFile)) {
		err = os.MkdirAll(filepath.Dir(mappingFile), 0755)
		if err != nil {
			return errors.Trace(err)
		}
	}
	return ioutil.WriteFile(mappingFile, data, 0755)
}

func (s *ServiceMapping) SetServicePorts(serviceName string, ports []int) error {
	s.Lock()
	defer s.Unlock()

	s.services[serviceName] = &serviceMappingInfo{
		PortInfo: portsInfo{
			Ports: ports,
		},
	}
	err := s.save()
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (s *ServiceMapping) DeleteServicePorts(serviceName string) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.services[serviceName]; !ok {
		return nil
	}

	delete(s.services, serviceName)
	err := s.save()
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (s *ServiceMapping) GetServiceNextPort(serviceName string) (int, error) {
	s.Lock()
	defer s.Unlock()

	if s.error != nil {
		return 0, s.error
	}

	serviceInfo, ok := s.services[serviceName]
	if !ok {
		return 0, errors.New("no such service in services mapping file")
	}

	if len(serviceInfo.PortInfo.Ports) == 0 {
		return 0, errors.New("no ports info in services mapping file")
	}

	port, err := s.services[serviceName].PortInfo.Next()
	if err != nil {
		return 0, err
	}
	return port, nil
}

func (s *ServiceMapping) WatchFile(logger *log.Logger) error {
	hostPath, err := context.HostPathLib()
	if err != nil {
		return err
	}
	mappingFile := filepath.Join(hostPath, ServiceMappingFile)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Trace(err)
	}

	s.tomb.Go(func() error {
		defer func() {
			watcher.Close()
			logger.Info("stop to watch services mapping file", log.Any("file", mappingFile))
		}()
		logger.Info("start to watch services mapping file", log.Any("file", mappingFile))

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return nil
				}
				logger.Debug("received a file event", log.Any("eventName", event.Name), log.Any("eventOp", event.Op))

				if event.Op&fsnotify.Write != fsnotify.Write {
					continue
				}

				logger.Debug("load services mapping file again")
				s.Lock()
				err := s.load()
				if err != nil {
					s.error = err
					logger.Warn("load services mapping file failed", log.Error(err))
				}
				s.Unlock()
			case err, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
				// TODO: check return or continue under this case
				logger.Warn(err.Error())
				s.Lock()
				s.error = err
				s.Unlock()
				return nil
			case <-s.tomb.Dying():
				return nil
			}
		}
	})

	err = watcher.Add(mappingFile)
	if err != nil {
		return errors.Trace(err)
	}

	s.Lock()
	err = s.load()
	s.Unlock()
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (s *ServiceMapping) Close() {
	s.tomb.Kill(nil)
	s.tomb.Wait()
}
