package dmcontext

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	mqtt2 "github.com/baetyl/baetyl-go/v2/mqtt"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
)

const (
	DefaultDriverConf = "/etc/baetyl/driver.yml"
	DefaultPropsConf  = "/etc/baetyl/props.yml"
	KeyDevice         = "device"
	KeyStatus         = "status"
	OnlineStatus      = "online"
	OfflineStatus     = "offline"
)

var (
	ErrCallbackNotRegister     = errors.New("callback not registered yet")
	ErrInvalidCallback         = errors.New("invalid device callback")
	ErrInvalidMessage          = errors.New("invalid device message")
	ErrInvalidChannel          = errors.New("invalid channel")
	ErrResponseChannelNotExist = errors.New("response channel not exist")
)

type DevicePropConfigs struct {
	Name        string            `yaml:"name,omitempty" json:"name,omitempty"`
	PropConfigs map[string]string `yaml:"propConfigs,omitempty" json:"propConfigs,omitempty"`
}

type DeviceProperties struct {
	Name  string                 `json:"name,omitempty"`
	Props map[string]interface{} `json:"props,omitempty"`
}

type DeviceShadow struct {
	Name   string    `json:"name,omitempty"`
	Report v1.Report `json:"report,omitempty"`
	Desire v1.Desire `json:"desire,omitempty"`
}

type DeviceEvent struct {
	Name  string      `json:"name,omitempty"`
	Event interface{} `json:"event,omitempty"`
}

type DeltaCallback func(*DeviceInfo, *DeviceProperties) error

type EventCallback func(*DeviceInfo, *DeviceEvent) error

type Context interface {
	context.Context
	SystemConfigExt() *SystemConfig
	GetAllDevices() []DeviceInfo
	ReportDeviceProperties(info *DeviceInfo, props *DeviceProperties) error
	GetDeviceProperties(info *DeviceInfo) (*DeviceShadow, error)
	RegisterDeltaCallback(info *DeviceInfo, cb DeltaCallback)
	RegisterEventCallback(info *DeviceInfo, cb EventCallback)
	Online(info *DeviceInfo) error
	Offline(info *DeviceInfo) error
	GetDriverConfig() (string, error)
	GetDevicePropConfigs() (map[string]DevicePropConfigs, error)
	Start()
	io.Closer
}

type DmCtx struct {
	context.Context
	log      *log.Logger
	mqtt     *mqtt2.Client
	tomb     utils.Tomb
	events   sync.Map
	deltas   sync.Map
	response sync.Map
	msgs     map[string]chan *v1.Message
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
	var sc SystemConfig
	if err := c.LoadCustomConfig(sc); err != nil {
		c.log.Error("failed to load system config, to use default config", log.Error(err))
		utils.UnmarshalYAML(nil, sc)
	}
	c.Store(context.KeySysConf, sc)

	var subs []mqtt2.QOSTopic
	for _, dev := range sc.Devices {
		subs = append(subs, dev.Delta, dev.Event, dev.GetResponse)
	}
	mqtt, err := c.Context.NewSystemBrokerClient(subs)
	if err != nil {
		c.log.Warn("fail to create system broker client", log.Any("error", err))
	}
	c.mqtt = mqtt
	c.msgs = make(map[string]chan *v1.Message, 1024)
	if err := c.mqtt.Start(newObserver(c.msgs, c.log)); err != nil {
		c.log.Warn("fail to start mqtt client", log.Any("error", err))
	}
	return c
}

func (c *DmCtx) Start() {
	devices := c.SystemConfigExt().Devices
	for _, dev := range devices {
		c.msgs[dev.Name] = make(chan *v1.Message)
		go c.processing(c.msgs[dev.Name])
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
	device := msg.Metadata[KeyDevice]
	val, ok := c.deltas.Load(device)
	if !ok {
		return errors.Trace(ErrCallbackNotRegister)
	}
	deltaCallback, ok := val.(DeltaCallback)
	if !ok {
		return errors.Trace(ErrInvalidCallback)
	}
	var props *DeviceProperties
	if err := msg.Content.Unmarshal(&props); err != nil {
		return errors.Trace(err)
	}
	if !ok {
		return errors.Trace(ErrInvalidMessage)
	}
	if err := deltaCallback(&DeviceInfo{Name: device}, props); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (c *DmCtx) processEvent(msg *v1.Message) error {
	device := msg.Metadata[KeyDevice]
	val, ok := c.events.Load(device)
	if !ok {
		return errors.Trace(ErrCallbackNotRegister)
	}
	eventCallback, ok := val.(EventCallback)
	if !ok {
		return errors.Trace(ErrInvalidCallback)
	}
	var event *DeviceEvent
	if err := msg.Content.Unmarshal(&event); err != nil {
		return errors.Trace(err)
	}
	if !ok {
		return errors.Trace(ErrInvalidMessage)
	}
	if err := eventCallback(&DeviceInfo{Name: device}, event); err != nil {
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
			case v1.MessageDelta:
				if err := c.processDelta(msg); err != nil {
					c.log.Error("failed to process delta message", log.Error(err))
				} else {
					c.log.Info("process delta message successfully")
				}
			case v1.MessageEvent:
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

func (c *DmCtx) SystemConfigExt() *SystemConfig {
	v, ok := c.Load(context.KeySysConf)
	if !ok {
		return nil
	}
	return v.(*SystemConfig)
}

func (c *DmCtx) GetAllDevices() []DeviceInfo {
	return c.SystemConfigExt().Devices
}

func (c *DmCtx) ReportDeviceProperties(info *DeviceInfo, props *DeviceProperties) error {
	msg := &v1.Message{
		Kind:     v1.MessageReport,
		Metadata: map[string]string{KeyDevice: info.Name},
		Content:  v1.LazyValue{Value: props},
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

func (c *DmCtx) RegisterDeltaCallback(info *DeviceInfo, cb DeltaCallback) {
	_, ok := c.deltas.LoadOrStore(info.Name, cb)
	if ok {
		c.log.Info("delta callback has updated for device", log.Any("device", info.Name))
	} else {
		c.log.Info("delta callback has registered for device", log.Any("device", info.Name))
	}
}

func (c *DmCtx) RegisterEventCallback(info *DeviceInfo, cb EventCallback) {
	_, ok := c.events.LoadOrStore(info.Name, cb)
	if ok {
		c.log.Info("event callback has updated for device", log.Any("device", info.Name))
	} else {
		c.log.Info("event callback has registered for device", log.Any("device", info.Name))
	}
}

func (c *DmCtx) Online(info *DeviceInfo) error {
	props := DeviceProperties{Name: info.Name, Props: map[string]interface{}{KeyStatus: OnlineStatus}}
	msg := &v1.Message{
		Kind:     v1.MessageReport,
		Metadata: map[string]string{KeyDevice: info.Name},
		Content:  v1.LazyValue{Value: props},
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
	props := DeviceProperties{Name: info.Name, Props: map[string]interface{}{KeyStatus: OfflineStatus}}
	msg := &v1.Message{
		Kind:     v1.MessageReport,
		Metadata: map[string]string{KeyDevice: info.Name},
		Content:  v1.LazyValue{Value: props},
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

func (c *DmCtx) GetDriverConfig() (string, error) {
	res, err := ioutil.ReadFile(DefaultDriverConf)
	if err != nil {
		return "", errors.Trace(err)
	}
	return string(res), nil
}

func (c *DmCtx) GetDevicePropConfigs() (map[string]DevicePropConfigs, error) {
	var res map[string]DevicePropConfigs
	if err := c.LoadCustomConfig(&res, DefaultPropsConf); err != nil {
		return nil, errors.Trace(err)
	}
	return res, nil
}
