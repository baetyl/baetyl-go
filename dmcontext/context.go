package dmcontext

import (
	"os"
	"path/filepath"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"gopkg.in/yaml.v2"
)

var _ Context = &DmCtx{}

const (
	DefaultSubDeviceConf      = "sub_devices.yml"
	DefaultDeviceModelConf    = "models.yml"
	DefaultAccessTemplateConf = "access_template.yml"
)

var (
	ErrInvalidPropertyKey     = errors.New("invalid property key")
	ErrDeviceModelNotExist    = errors.New("device model not exist")
	ErrAccessTemplateNotExist = errors.New("access template not exist")
	ErrPropsConfigNotExist    = errors.New("properties config not exist")
	ErrDeviceNotExist         = errors.New("device not exist")
	ErrTypeNotSupported       = errors.New("type not supported")
	ErrInvalidDelta           = errors.New("invalid delta")
)

type Context interface {
	context.Context

	GetDevice(driverName, device string) (*DeviceInfo, error)
	GetDriverNameByDevice(device string) string
	GetAllDevices(driverName string) []DeviceInfo
	GetDeviceModel(driverName string, device *DeviceInfo) ([]DeviceProperty, error)
	GetAccessTemplates(driverName, name string) (*AccessTemplate, error)
	ParsePropertyValues(driverName string, device *DeviceInfo, props map[string]any) (map[string]any, error)
	LoadDriverConfig(path, driverName string) error
}

type DmCtx struct {
	context.Context
	log             *log.Logger
	deviceDriverMap map[string]string
	devices         map[string]map[string]DeviceInfo
	deviceModels    map[string]map[string][]DeviceProperty
	accessTemplates map[string]map[string]AccessTemplate
}

func NewContext(confFile string) Context {
	var c = new(DmCtx)
	c.Context = context.NewContext(confFile)

	var lfs []log.Field
	if c.NodeName() != "" {
		lfs = append(lfs, log.Any("node", c.NodeName()))
	}
	if c.AppName() != "" {
		lfs = append(lfs, log.Any("app", c.AppName()))
	}
	if c.ServiceName() != "" {
		lfs = append(lfs, log.Any("service", c.ServiceName()))
	}
	c.log = log.With(lfs...)
	c.deviceModels = make(map[string]map[string][]DeviceProperty)
	c.accessTemplates = make(map[string]map[string]AccessTemplate)
	c.devices = make(map[string]map[string]DeviceInfo)
	c.deviceDriverMap = make(map[string]string)
	return c
}

func (c *DmCtx) LoadDriverConfig(path, driverName string) error {
	deviceModel := make(map[string][]DeviceProperty)
	if err := unmarshalYAML(filepath.Join(path, DefaultDeviceModelConf), deviceModel); err != nil {
		c.log.Error("failed to load device model", log.Error(err))
		return err
	}
	c.deviceModels[driverName] = deviceModel

	accessTemplate := make(map[string]AccessTemplate)
	if err := unmarshalYAML(filepath.Join(path, DefaultAccessTemplateConf), accessTemplate); err != nil {
		c.log.Error("failed to load access template", log.Error(err))
		return err
	}
	c.accessTemplates[driverName] = accessTemplate

	for name, tpl := range c.accessTemplates[driverName] {
		tpl.Name = name
		c.accessTemplates[driverName][name] = tpl
	}

	var dCfg driverConfig
	if err := unmarshalYAML(filepath.Join(path, DefaultSubDeviceConf), &dCfg); err != nil {
		c.log.Error("failed to load device config", log.Error(err))
		return err
	}

	devices := make(map[string]DeviceInfo)
	for _, dev := range dCfg.Devices {
		devices[dev.Name] = dev
		c.deviceDriverMap[dev.Name] = driverName
	}
	c.devices[driverName] = devices

	return nil
}

func (c *DmCtx) GetAllDevices(driverName string) []DeviceInfo {
	var deviceList []DeviceInfo
	for _, dev := range c.devices[driverName] {
		deviceList = append(deviceList, dev)
	}
	return deviceList
}

func (c *DmCtx) GetDevice(driverName, device string) (*DeviceInfo, error) {
	if deviceInfo, ok := c.devices[driverName][device]; ok {
		return &deviceInfo, nil
	}
	return nil, ErrDeviceNotExist
}

func (c *DmCtx) GetDriverNameByDevice(device string) string {
	if driverName, ok := c.deviceDriverMap[device]; ok {
		return driverName
	}
	return ""
}

func (c *DmCtx) GetDeviceModel(driverName string, device *DeviceInfo) ([]DeviceProperty, error) {
	if devProp, ok := c.deviceModels[driverName][device.DeviceModel]; ok {
		return devProp, nil
	}
	return nil, ErrDeviceModelNotExist
}

func (c *DmCtx) GetAccessTemplates(driverName, name string) (*AccessTemplate, error) {
	if accessTemplate, ok := c.accessTemplates[driverName][name]; ok {
		return &accessTemplate, nil
	}
	return nil, ErrAccessTemplateNotExist
}

func unmarshalYAML(file string, out any) error {
	bs, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bs, out)
}
