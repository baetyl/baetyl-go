package dmcontext

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"strconv"
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
	DefaultAccessConf         = "etc/baetyl/access.yml"
	DefaultPropsConf          = "etc/baetyl/props.yml"
	DefaultDriverConf         = "etc/baetyl/conf.yml"
	DefaultSubDeviceConf      = "etc/baetyl/sub_devices.yml"
	DefaultDeviceModelConf    = "etc/baetyl/models.yml"
	DefaultAccessTemplateConf = "etc/baetyl/access_template.yml"
	KeyDevice                 = "device"
	KeyDeviceProduct          = "deviceProduct"
	KeyAccessTemplate         = "accessTemplate"
	KeyNode                   = "node"
	KeyNodeProduct            = "nodeProduct"
	NodeProduct               = "BIE-Product"
	KeyStatus                 = "status"
	OnlineStatus              = "online"
	OfflineStatus             = "offline"
	TypeReportEvent           = "report"
)

var (
	ErrInvalidMessage          = errors.New("invalid device message")
	ErrInvalidChannel          = errors.New("invalid channel")
	ErrInvalidDelta            = errors.New("invalid delta")
	ErrInvalidPropertyKey      = errors.New("invalid property key")
	ErrResponseChannelNotExist = errors.New("response channel not exist")
	ErrAccessConfigNotExist    = errors.New("access config not exist")
	ErrDeviceModelNotExist     = errors.New("device model not exist")
	ErrAccessTemplateNotExist  = errors.New("access template not exist")
	ErrPropsConfigNotExist     = errors.New("properties config not exist")
	ErrDeviceNotExist          = errors.New("device not exist")
	ErrTypeNotSupported        = errors.New("type not supported")
)

const (
	TypeInt16   = "int16"
	TypeInt32   = "int32"
	TypeInt64   = "int64"
	TypeFloat32 = "float32"
	TypeFloat64 = "float64"
	TypeBool    = "bool"
	TypeString  = "string"
)

type DeltaCallback func(*DeviceInfo, v1.Delta) error
type EventCallback func(*DeviceInfo, *Event) error
type PropertyGetCallback func(*DeviceInfo, []string) error

var oldData = make(map[string]PointClean)

type Context interface {
	context.Context
	GetAllDevices() []DeviceInfo
	ReportDeviceProperties(*DeviceInfo, v1.Report) error
	ReportDeviceEvents(*DeviceInfo, v1.EventReport) error
	GetDeviceProperties(device *DeviceInfo) (*DeviceShadow, error)
	RegisterDeltaCallback(cb DeltaCallback) error
	RegisterEventCallback(cb EventCallback) error
	RegisterPropertyGetCallback(cb PropertyGetCallback) error
	DataClean(device *DeviceInfo, report v1.Report) (convertReport v1.Report, err error)
	Online(device *DeviceInfo) error
	Offline(device *DeviceInfo) error
	GetDriverConfig() string
	GetAccessConfig() map[string]AccessConfig
	GetDeviceAccessConfig(device *DeviceInfo) (*AccessConfig, error)
	GetPropertiesConfig() map[string][]DeviceProperty
	GetDevicePropertiesConfig(device *DeviceInfo) ([]DeviceProperty, error)
	GetDevice(device string) (*DeviceInfo, error)
	GetDeviceModel(device *DeviceInfo) ([]DeviceProperty, error)
	GetAllDeviceModels() map[string][]DeviceProperty
	GetAccessTemplates(device *DeviceInfo) (*AccessTemplate, error)
	GetAllAccessTemplates() map[string]AccessTemplate
	Start()
	io.Closer
}

type DmCtx struct {
	context.Context
	log             *log.Logger
	mqtt            *mqtt2.Client
	tomb            utils.Tomb
	eventCb         EventCallback
	deltaCb         DeltaCallback
	propertyGetCb   PropertyGetCallback
	response        sync.Map
	devices         map[string]DeviceInfo
	msgChs          map[string]chan *v1.Message
	driverConfig    string
	propsConfig     map[string][]DeviceProperty
	accessConfig    map[string]AccessConfig
	deviceModels    map[string][]DeviceProperty
	accessTemplates map[string]AccessTemplate
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

	if err := unmarshalYAML(DefaultDeviceModelConf, &c.deviceModels); err != nil {
		c.log.Error("failed to load device model, to use default config", log.Error(err))
		utils.UnmarshalYAML(nil, &c.deviceModels)
	}

	if err := unmarshalYAML(DefaultAccessTemplateConf, &c.accessTemplates); err != nil {
		c.log.Error("failed to load access template, to use default config", log.Error(err))
		utils.UnmarshalYAML(nil, &c.accessTemplates)
	}
	for name, tpl := range c.accessTemplates {
		tpl.Name = name
		c.accessTemplates[name] = tpl
	}

	var dCfg driverConfig
	if err := unmarshalYAML(DefaultSubDeviceConf, &dCfg); err != nil {
		c.log.Error("failed to load device config, to use default config", log.Error(err))
		utils.UnmarshalYAML(nil, &dCfg)
	}
	c.driverConfig = dCfg.Driver

	devices := make(map[string]DeviceInfo)
	var subs []mqtt2.QOSTopic
	for _, dev := range dCfg.Devices {
		subs = append(subs, dev.DeviceTopic.Delta, dev.DeviceTopic.Event,
			dev.DeviceTopic.GetResponse, dev.DeviceTopic.PropertyGet)
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
	var blinkContent BlinkContent
	if err := msg.Content.ExactUnmarshal(&blinkContent); err != nil {
		return errors.Trace(err)
	}
	delta, ok := blinkContent.Blink.Properties.(map[string]interface{})
	if !ok {
		return errors.Trace(ErrInvalidDelta)
	}
	dev, ok := c.devices[deviceName]
	if !ok {
		return errors.Trace(ErrDeviceNotExist)
	}
	delta, err := c.parsePropertyValues(&dev, delta)
	if err != nil {
		return errors.Trace(err)
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

func (c *DmCtx) processPropertyGet(msg *v1.Message) error {
	deviceName := msg.Metadata[KeyDevice]
	if c.propertyGetCb == nil {
		c.log.Debug("property get callback not set and message will not be process")
		return nil
	}
	dev, ok := c.devices[deviceName]
	if !ok {
		c.log.Warn("property get callback can not find device", log.Any("device", deviceName))
		return nil
	}
	var blinkContent BlinkContent
	if err := msg.Content.Unmarshal(&blinkContent); err != nil {
		return errors.Trace(err)
	}
	properties, err := parsePropertyKeys(blinkContent.Blink.Properties)
	if err != nil {
		return errors.Trace(err)
	}
	if err := c.propertyGetCb(&dev, properties); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (c *DmCtx) processResponse(msg *v1.Message) error {
	deviceName := msg.Metadata[KeyDevice]
	val, ok := c.response.Load(deviceName)
	if !ok {
		return errors.Trace(ErrResponseChannelNotExist)
	}
	ch, ok := val.(chan *DeviceShadow)
	if !ok {
		return errors.Trace(ErrInvalidChannel)
	}
	var shad *DeviceShadow
	if err := msg.Content.ExactUnmarshal(&shad); err != nil {
		return errors.Trace(err)
	}
	if !ok {
		return errors.Trace(ErrInvalidMessage)
	}
	var err error
	dev := c.devices[deviceName]
	shad.Report, err = c.parsePropertyValues(&dev, shad.Report)
	if err != nil {
		return errors.Trace(err)
	}
	shad.Desire, err = c.parsePropertyValues(&dev, shad.Desire)
	if err != nil {
		return errors.Trace(err)
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
			case v1.MessageDevicePropertyGet:
				if err := c.processPropertyGet(msg); err != nil {
					c.log.Error("failed to process property get message", log.Error(err))
				} else {
					c.log.Info("process property get message successfully")
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

func (c *DmCtx) GetDevice(device string) (*DeviceInfo, error) {
	if deviceInfo, ok := c.devices[device]; ok {
		return &deviceInfo, nil
	}
	return nil, ErrDeviceNotExist
}

// 静默窗口（SilentWin）: 静默时间窗口、偏差（Deviation）: 震荡的偏差区间值
func (c *DmCtx) DataClean(device *DeviceInfo, report v1.Report) (convertReport v1.Report, err error) {
	//条件1：当前上报值与前一次上报值的差值超过偏差的区间
	//条件2：当前上报与上一次上报的时间差，超过静默时间窗
	reportData := make(map[string]interface{})
	mappData := make(map[string]PointClean)

	accessTemplates, err := c.GetAccessTemplates(device)
	if err != nil {
		return nil, err
	}
	for _, mapping := range accessTemplates.Mappings {
		mappData[mapping.Attribute] = PointClean{SilentWin: mapping.SilentWin, Deviation: mapping.Deviation}
	}

	for key, value := range report {
		_, ok := oldData[key]
		if !ok {
			reportData[key] = value
			oldData[key] = PointClean{SilentWin: time.Now().Unix()}
		}

		var postValue float64
		switch value.(type) {
		case bool:
			reportData[key] = value
		case string:
			reportData[key] = value
		case float64:
			postValue = value.(float64)
		default:
			return nil, ErrUnsupportedValueType
		}

		clean, isExist := mappData[key]
		if isExist {
			if postValue-oldData[key].Deviation > clean.Deviation || time.Now().Unix()-oldData[key].SilentWin > clean.SilentWin {
				reportData[key] = value
				oldData[key] = PointClean{SilentWin: time.Now().Unix()}
			}
			oldData[key] = PointClean{Deviation: clean.Deviation}
		}
	}

	return reportData, nil
}

func (c *DmCtx) ReportDeviceProperties(info *DeviceInfo, report v1.Report) error {
	msg := &v1.Message{
		Kind:     v1.MessageDeviceReport,
		Metadata: c.genMetadata(info),
		Content:  v1.LazyValue{Value: BlinkContent{Blink: GenPropertyReportBlinkData(report)}},
	}
	pld, err := json.Marshal(msg)
	if err != nil {
		return errors.Trace(err)
	}
	if err := c.mqtt.Publish(mqtt2.QOS(info.DeviceTopic.Report.QOS),
		info.DeviceTopic.Report.Topic, pld, 0, false, false); err != nil {
		return err
	}
	return nil
}

func (c *DmCtx) ReportDeviceEvents(info *DeviceInfo, report v1.EventReport) error {
	msg := &v1.Message{
		Kind:     v1.MessageDeviceEventReport,
		Metadata: c.genMetadata(info),
		Content:  v1.LazyValue{Value: BlinkContent{Blink: GenEventReportBlinkData(report)}},
	}
	pld, err := json.Marshal(msg)
	if err != nil {
		return errors.Trace(err)
	}
	if err := c.mqtt.Publish(mqtt2.QOS(info.DeviceTopic.EventReport.QOS),
		info.DeviceTopic.EventReport.Topic, pld, 0, false, false); err != nil {
		return err
	}
	return nil
}

func (c *DmCtx) GetDeviceProperties(info *DeviceInfo) (*DeviceShadow, error) {
	pld, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}
	if err := c.mqtt.Publish(mqtt2.QOS(info.DeviceTopic.Get.QOS),
		info.DeviceTopic.Get.Topic, pld, 0, false, false); err != nil {
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

func (c *DmCtx) RegisterPropertyGetCallback(cb PropertyGetCallback) error {
	if c.propertyGetCb != nil {
		return errors.New("property get callback already registered")
	}
	c.propertyGetCb = cb
	c.log.Debug("register property get callback successfully")
	return nil
}

func (c *DmCtx) Online(info *DeviceInfo) error {
	msg := &v1.Message{
		Kind:     v1.MessageDeviceLifecycleReport,
		Metadata: c.genMetadata(info),
		Content:  v1.LazyValue{Value: BlinkContent{Blink: GenLifecycleReportBlinkData(true)}},
	}
	pld, err := json.Marshal(msg)
	if err != nil {
		return errors.Trace(err)
	}
	if err := c.mqtt.Publish(mqtt2.QOS(info.DeviceTopic.LifecycleReport.QOS),
		info.DeviceTopic.LifecycleReport.Topic, pld, 0, false, false); err != nil {
		return err
	}
	return nil
}

func (c *DmCtx) Offline(info *DeviceInfo) error {
	msg := &v1.Message{
		Kind:     v1.MessageDeviceLifecycleReport,
		Metadata: c.genMetadata(info),
		Content:  v1.LazyValue{Value: BlinkContent{Blink: GenLifecycleReportBlinkData(false)}},
	}
	pld, err := json.Marshal(msg)
	if err != nil {
		return errors.Trace(err)
	}
	if err := c.mqtt.Publish(mqtt2.QOS(info.DeviceTopic.LifecycleReport.QOS),
		info.DeviceTopic.LifecycleReport.Topic, pld, 0, false, false); err != nil {
		return err
	}
	return nil
}

func (c *DmCtx) GetDriverConfig() string {
	return c.driverConfig
}

// Deprecated: Use GetAllDevices instead.
// Change from access template support
func (c *DmCtx) GetAccessConfig() map[string]AccessConfig {
	return c.accessConfig
}

// Deprecated: Use GetDevice instead.
// Change from access template support
func (c *DmCtx) GetDeviceAccessConfig(device *DeviceInfo) (*AccessConfig, error) {
	if cfg, ok := c.accessConfig[device.Name]; ok {
		return &cfg, nil
	} else {
		return nil, ErrAccessConfigNotExist
	}
}

// Deprecated: Use GetAllDeviceModels instead.
// Change from access template support
func (c *DmCtx) GetPropertiesConfig() map[string][]DeviceProperty {
	return c.propsConfig
}

// Deprecated: Use GetDeviceModel instead.
// Change from access template support
func (c *DmCtx) GetDevicePropertiesConfig(device *DeviceInfo) ([]DeviceProperty, error) {
	if cfg, ok := c.propsConfig[device.Name]; ok {
		return cfg, nil
	} else {
		return nil, ErrPropsConfigNotExist
	}
}

func (c *DmCtx) GetDeviceModel(device *DeviceInfo) ([]DeviceProperty, error) {
	if devProp, ok := c.deviceModels[device.DeviceModel]; ok {
		return devProp, nil
	}
	return nil, ErrDeviceModelNotExist
}

func (c *DmCtx) GetAllDeviceModels() map[string][]DeviceProperty {
	return c.deviceModels
}

func (c *DmCtx) GetAccessTemplates(device *DeviceInfo) (*AccessTemplate, error) {
	if accessTemplate, ok := c.accessTemplates[device.AccessTemplate]; ok {
		return &accessTemplate, nil
	}
	return nil, ErrAccessTemplateNotExist
}

func (c *DmCtx) GetAllAccessTemplates() map[string]AccessTemplate {
	return c.accessTemplates
}

func unmarshalYAML(file string, out interface{}) error {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bs, out)
}

func (c *DmCtx) parsePropertyValues(device *DeviceInfo, props map[string]interface{}) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	vals, ok := c.deviceModels[device.DeviceModel]
	if !ok {
		return nil, errors.Trace(ErrDeviceNotExist)
	}
	cfgs := make(map[string]DeviceProperty)
	for _, val := range vals {
		cfgs[val.Name] = val
	}
	for key, val := range props {
		if cfg, ok := cfgs[key]; ok {
			pVal, err := parsePropertyValue(cfg.Type, val)
			if err != nil {
				return nil, errors.Trace(err)
			}
			res[key] = pVal
		} else {
			return nil, errors.Trace(ErrPropsConfigNotExist)
		}
	}
	return res, nil
}

func parsePropertyValue(tpy string, val interface{}) (interface{}, error) {
	// it is json.Number (string actually) when val is number
	switch tpy {
	case TypeInt16:
		num, _ := val.(json.Number)
		i, err := strconv.ParseInt(num.String(), 10, 16)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return int16(i), nil
	case TypeInt32:
		num, _ := val.(json.Number)
		i, err := strconv.ParseInt(num.String(), 10, 32)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return int32(i), nil
	case TypeInt64:
		num, _ := val.(json.Number)
		i, err := strconv.ParseInt(num.String(), 10, 64)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return i, nil
	case TypeFloat32:
		num, _ := val.(json.Number)
		f, err := strconv.ParseFloat(num.String(), 32)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return float32(f), nil
	case TypeFloat64:
		num, _ := val.(json.Number)
		f, err := strconv.ParseFloat(num.String(), 64)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return f, nil
	case TypeBool, TypeString:
		return val, nil
	default:
		return nil, errors.Trace(ErrTypeNotSupported)
	}
}

func parsePropertyKeys(v interface{}) ([]string, error) {
	properties, ok := v.([]interface{})
	if !ok {
		return nil, ErrInvalidPropertyKey
	}
	var propertyKeys []string
	for _, key := range properties {
		propertyKey, ok := key.(string)
		if !ok {
			return nil, ErrInvalidPropertyKey
		}
		propertyKeys = append(propertyKeys, propertyKey)
	}
	return propertyKeys, nil
}

func (c *DmCtx) genMetadata(info *DeviceInfo) map[string]string {
	return map[string]string{
		KeyDevice:         info.Name,
		KeyDeviceProduct:  info.DeviceModel,
		KeyAccessTemplate: info.AccessTemplate,
		KeyNode:           c.NodeName(),
		KeyNodeProduct:    NodeProduct,
	}
}
