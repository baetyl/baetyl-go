package dmcontext

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	mqtt2 "github.com/baetyl/baetyl-go/v2/mqtt"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
)

const (
	DefaultAccessConf = "etc/baetyl/access.yml"
	DefaultPropsConf  = "etc/baetyl/props.yml"
	DefaultDriverConf = "etc/baetyl/conf.yml"
	KeyDevice         = "device"
	KeyStatus         = "status"
	OnlineStatus      = "online"
	OfflineStatus     = "offline"
	TypeReportEvent   = "report"
)

var (
	ErrInvalidMessage          = errors.New("invalid device message")
	ErrInvalidChannel          = errors.New("invalid channel")
	ErrResponseChannelNotExist = errors.New("response channel not exist")
	ErrAccessConfigNotExist    = errors.New("access config not exist")
	ErrPropsConfigNotExist     = errors.New("properties config not exist")
)

type DeltaCallback func(*DeviceInfo, v1.Delta) error
type EventCallback func(*DeviceInfo, *Event) error

type Context interface {
	context.Context
	GetAllDevices() []DeviceInfo
	ReportDeviceProperties(*DeviceInfo, v1.Report) error
	GetDeviceProperties(device *DeviceInfo) (*DeviceShadow, error)
	RegisterDeltaCallback(cb DeltaCallback) error
	RegisterEventCallback(cb EventCallback) error
	Online(device *DeviceInfo) error
	Offline(device *DeviceInfo) error
	GetDriverConfig() string
	GetAccessConfig() map[string]string
	GetDeviceAccessConfig(device *DeviceInfo) (string, error)
	GetPropertiesConfig() map[string][]DeviceProperty
	GetDevicePropertiesConfig(device *DeviceInfo) ([]DeviceProperty, error)
	Start()
	io.Closer
}

type DmCtx struct {
	context.Context
	log          *log.Logger
	mqtt         *mqtt2.Client
	tomb         utils.Tomb
	eventCb      EventCallback
	deltaCb      DeltaCallback
	response     sync.Map
	devices      map[string]DeviceInfo
	msgChs       map[string]chan *v1.Message
	driverConfig string
	propsConfig  map[string][]DeviceProperty
	accessConfig map[string]string
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

	if err := unmarshalYAML(DefaultAccessConf, &c.accessConfig); err != nil {
		c.log.Error("failed to load access config, to use default config", log.Error(err))
		utils.UnmarshalYAML(nil, &c.accessConfig)
	}

	if err := unmarshalYAML(DefaultPropsConf, &c.propsConfig); err != nil {
		c.log.Error("failed to load props config, to use default config", log.Error(err))
		utils.UnmarshalYAML(nil, &c.propsConfig)
	}

	var dCfg driverConfig
	if err := unmarshalYAML(DefaultDriverConf, &dCfg); err != nil {
		c.log.Error("failed to load driver config, to use default config", log.Error(err))
		utils.UnmarshalYAML(nil, &dCfg)
	}
	c.driverConfig = dCfg.Driver

	devices := make(map[string]DeviceInfo)
	var subs []mqtt2.QOSTopic
	for _, dev := range dCfg.Devices {
		subs = append(subs, dev.Delta, dev.Event, dev.GetResponse)
		devices[dev.Name] = dev
	}
	c.devices = devices
	mqtt, err := c.Context.NewSystemBrokerClient(subs)
	if err != nil {
		c.log.Warn("fail to create system broker client", log.Any("error", err))
	}
	c.mqtt = mqtt
	c.msgChs = make(map[string]chan *v1.Message)
	if err := c.mqtt.Start(newObserver(c.msgChs, c.log)); err != nil {
		c.log.Warn("fail to start mqtt client", log.Any("error", err))
	}
	return c
}

func (c *DmCtx) Start() {
	for name, dev := range c.devices {
		c.msgChs[name] = make(chan *v1.Message, 1024)
		go c.processing(c.msgChs[dev.Name])
	}
}

func (c *DmCtx) Close() error {
	if c.mqtt != nil {
		c.mqtt.Close()
	}
	c.tomb.Kill(nil)
	return c.tomb.Wait()
}

func (c *DmCtx) processDelta(msg *v1.Message) error {
	deviceName := msg.Metadata[KeyDevice]
	if c.deltaCb == nil {
		c.log.Debug("delta callback not set and message will not be process")
		return nil
	}
	var delta v1.Delta
	if err := msg.Content.Unmarshal(&delta); err != nil {
		return errors.Trace(err)
	}
	dev, ok := c.devices[deviceName]
	if !ok {
		c.log.Warn("delta callback can not find device", log.Any("device", deviceName))
		return nil
	}
	if err := c.deltaCb(&dev, delta); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (c *DmCtx) processEvent(msg *v1.Message) error {
	deviceName := msg.Metadata[KeyDevice]
	if c.eventCb == nil {
		c.log.Debug("event callback not set and message will not be process")
		return nil
	}
	var event Event
	if err := msg.Content.Unmarshal(&event); err != nil {
		return errors.Trace(err)
	}
	dev, ok := c.devices[deviceName]
	if !ok {
		c.log.Warn("event callback can not find device", log.Any("device", deviceName))
		return nil
	}
	if err := c.eventCb(&dev, &event); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (c *DmCtx) processResponse(msg *v1.Message) error {
	device := msg.Metadata[KeyDevice]
	val, ok := c.response.Load(device)
	if !ok {
		return errors.Trace(ErrResponseChannelNotExist)
	}
	ch, ok := val.(chan *DeviceShadow)
	if !ok {
		return errors.Trace(ErrInvalidChannel)
	}
	var shad *DeviceShadow
	if err := msg.Content.Unmarshal(&shad); err != nil {
		return errors.Trace(err)
	}
	if !ok {
		return errors.Trace(ErrInvalidMessage)
	}
	select {
	case ch <- shad:
	default:
		c.log.Error("failed to write response message")
	}
	return nil
}

func (c *DmCtx) processing(ch chan *v1.Message) {
	for {
		select {
		case <-c.tomb.Dying():
			return
		case msg := <-ch:
			switch msg.Kind {
			case v1.MessageDeviceDelta:
				if err := c.processDelta(msg); err != nil {
					c.log.Error("failed to process delta message", log.Error(err))
				} else {
					c.log.Info("process delta message successfully")
				}
			case v1.MessageDeviceEvent:
				if err := c.processEvent(msg); err != nil {
					c.log.Error("failed to process event message", log.Error(err))
				} else {
					c.log.Info("process event message successfully")
				}
			case v1.MessageResponse:
				if err := c.processResponse(msg); err != nil {
					c.log.Error("failed to process response message", log.Error(err))
				} else {
					c.log.Info("process response message successfully")
				}
			default:
				c.log.Error("device message type not supported yet")
			}
		}
	}
}

func (c *DmCtx) GetAllDevices() []DeviceInfo {
	var deviceList []DeviceInfo
	for _, dev := range c.devices {
		deviceList = append(deviceList, dev)
	}
	return deviceList
}

func (c *DmCtx) ReportDeviceProperties(info *DeviceInfo, report v1.Report) error {
	msg := &v1.Message{
		Kind:     v1.MessageDeviceReport,
		Metadata: map[string]string{KeyDevice: info.Name},
		Content:  v1.LazyValue{Value: report},
	}
	pld, err := json.Marshal(msg)
	if err != nil {
		return errors.Trace(err)
	}
	if err := c.mqtt.Publish(mqtt2.QOS(info.Report.QOS),
		info.Report.Topic, pld, 0, false, false); err != nil {
		return err
	}
	return nil
}

func (c *DmCtx) GetDeviceProperties(info *DeviceInfo) (*DeviceShadow, error) {
	pld, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}
	if err := c.mqtt.Publish(mqtt2.QOS(info.Get.QOS),
		info.Get.Topic, pld, 0, false, false); err != nil {
		return nil, err
	}
	timer := time.NewTimer(time.Second)
	ch := make(chan *DeviceShadow)
	_, ok := c.response.LoadOrStore(info.Name, ch)
	if ok {
		// another routine may not finish getting device properties
		return nil, errors.Errorf("waiting for getting properties")
	}
	defer func() {
		timer.Stop()
		c.response.Delete(info.Name)
	}()
	for {
		select {
		case <-timer.C:
			c.log.Error("get device properties timeout", log.Any("device", info.Name))
			return nil, errors.Errorf("get device: %s properties timeout", info.Name)
		case props := <-ch:
			return props, nil
		}
	}
}

func (c *DmCtx) RegisterDeltaCallback(cb DeltaCallback) error {
	if c.deltaCb != nil {
		return errors.New("delta callback already registered")
	}
	c.deltaCb = cb
	c.log.Debug("register delta callback successfully")
	return nil
}

func (c *DmCtx) RegisterEventCallback(cb EventCallback) error {
	if c.eventCb != nil {
		return errors.New("event callback already registered")
	}
	c.eventCb = cb
	c.log.Debug("register event callback successfully")
	return nil
}

func (c *DmCtx) Online(info *DeviceInfo) error {
	r := v1.Report{KeyStatus: OnlineStatus}
	msg := &v1.Message{
		Kind:     v1.MessageDeviceReport,
		Metadata: map[string]string{KeyDevice: info.Name},
		Content:  v1.LazyValue{Value: r},
	}
	pld, err := json.Marshal(msg)
	if err != nil {
		return errors.Trace(err)
	}
	if err := c.mqtt.Publish(mqtt2.QOS(info.Report.QOS),
		info.Report.Topic, pld, 0, false, false); err != nil {
		return err
	}
	return nil
}

func (c *DmCtx) Offline(info *DeviceInfo) error {
	r := v1.Report{KeyStatus: OfflineStatus}
	msg := &v1.Message{
		Kind:     v1.MessageDeviceReport,
		Metadata: map[string]string{KeyDevice: info.Name},
		Content:  v1.LazyValue{Value: r},
	}
	pld, err := json.Marshal(msg)
	if err != nil {
		return errors.Trace(err)
	}
	if err := c.mqtt.Publish(mqtt2.QOS(info.Report.QOS),
		info.Report.Topic, pld, 0, false, false); err != nil {
		return err
	}
	return nil
}

func (c *DmCtx) GetDriverConfig() string {
	return c.driverConfig
}
func (c *DmCtx) GetAccessConfig() map[string]string {
	return c.accessConfig
}

func (c *DmCtx) GetDeviceAccessConfig(device *DeviceInfo) (string, error) {
	if cfg, ok := c.accessConfig[device.Name]; ok {
		return cfg, nil
	} else {
		return "", ErrAccessConfigNotExist
	}
}

func (c *DmCtx) GetPropertiesConfig() map[string][]DeviceProperty {
	return c.propsConfig
}

func (c *DmCtx) GetDevicePropertiesConfig(device *DeviceInfo) ([]DeviceProperty, error) {
	if cfg, ok := c.propsConfig[device.Name]; ok {
		return cfg, nil
	} else {
		return nil, ErrPropsConfigNotExist
	}
}

func unmarshalYAML(file string, out interface{}) error {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bs, out)
}
