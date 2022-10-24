package dmcontext

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/baetyl/baetyl-go/v2/errors"
	mqtt2 "github.com/baetyl/baetyl-go/v2/mqtt"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
)

type DeviceView struct {
	Name        string            `json:"name,omitempty"`
	Protocol    string            `json:"protocol,omitempty"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Ready       bool              `json:"ready"`
	DeviceModel string            `json:"deviceModel,omitempty"`
	Alias       string            `json:"alias,omitempty"`
	Status      int               `json:"status"`
	Bind        bool              `json:"bind"`
	NodeName    string            `json:"nodeName,omitempty"`
	Attributes  []DeviceAttribute `json:"attributes,omitempty"`
	Properties  []DeviceProperty  `json:"properties,omitempty"`
	Config      *DeviceConfigView `json:"config,omitempty"`
	CreateTime  time.Time         `json:"createTime,omitempty"`
	UpdateTime  time.Time         `json:"updateTime,omitempty"`
}

type Device struct {
	Name        string            `json:"name,omitempty" binding:"omitempty,res_name"`
	Namespace   string            `json:"namespace,omitempty"`
	Version     string            `json:"version,omitempty"`
	Description string            `json:"description,omitempty"`
	Protocol    string            `json:"protocol,omitempty"`
	Alias       string            `json:"alias,omitempty"`
	Ready       bool              `json:"ready"`
	Active      bool              `json:"active,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	DeviceModel string            `json:"deviceModel,omitempty"`
	Attributes  []DeviceAttribute `json:"attributes,omitempty"`
	Properties  []DeviceProperty  `json:"properties,omitempty"`
	// Deprecated: 移动到 baetyl_node_device 表实现
	NodeName string `json:"nodeName,omitempty"`
	// Deprecated: 移动到 baetyl_node_device 表实现
	DriverName string `json:"driverName,omitempty"`
	// Deprecated: 移动到 baetyl_node_device 表实现
	Config     *DeviceConfig `json:"config,omitempty"`
	Shadow     string        `json:"shadow,omitempty"`
	CreateTime time.Time     `json:"createTime,omitempty"`
	UpdateTime time.Time     `json:"updateTime,omitempty"`
}

func EqualDevice(old, new *Device) bool {
	if old.Name != new.Name || old.Protocol != new.Protocol || old.DeviceModel != new.DeviceModel {
		return false
	}
	if old.Alias != new.Alias || old.Description != new.Description {
		return false
	}
	if len(old.Labels) != len(new.Labels) || !reflect.DeepEqual(old.Labels, new.Labels) {
		return false
	}
	if len(old.Attributes) != len(new.Attributes) || !reflect.DeepEqual(old.Attributes, new.Attributes) {
		return false
	}
	if len(old.Properties) != len(new.Properties) || !reflect.DeepEqual(old.Properties, new.Properties) {
		return false
	}
	if !reflect.DeepEqual(old.Config, new.Config) {
		return false
	}
	if old.NodeName != new.NodeName || old.DriverName != new.DriverName || old.Shadow != new.Shadow || old.Ready != new.Ready || old.Active != new.Active {
		return false
	}
	return true
}

type DeviceAttribute struct {
	Name     string      `json:"name,omitempty"`
	Id       string      `json:"id,omitempty"`
	Type     string      `json:"type,omitempty" binding:"data_type"`
	Unit     string      `json:"unit,omitempty"`
	Required bool        `json:"required"`
	Value    interface{} `json:"value"`
}

type deviceAttribute struct {
	Name     string      `json:"name,omitempty"`
	Id       string      `json:"id,omitempty"`
	Type     string      `json:"type,omitempty" binding:"data_type"`
	Unit     string      `json:"unit,omitempty"`
	Required bool        `json:"required"`
	Value    interface{} `json:"value"`
}

func (da *DeviceAttribute) UnmarshalJSON(b []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	var attr deviceAttribute
	if err := decoder.Decode(&attr); err != nil {
		return err
	}
	da.Name = attr.Name
	da.Id = attr.Id
	da.Unit = attr.Unit
	da.Type = attr.Type
	da.Required = attr.Required
	da.Value = attr.Value
	return nil
}

type DeviceProperty struct {
	Name           string                `yaml:"name,omitempty" json:"name,omitempty"`
	Id             string                `yaml:"id,omitempty" json:"id,omitempty" binding:"nonzero"`
	Type           string                `yaml:"type,omitempty" json:"type,omitempty" binding:"data_type"`
	Mode           string                `yaml:"mode,omitempty" json:"mode,omitempty" binding:"oneof=ro rw"`
	Unit           string                `yaml:"unit,omitempty" json:"unit,omitempty"`
	Visitor        PropertyVisitor       `yaml:"visitor,omitempty" json:"visitor,omitempty"`
	Format         string                `json:"format,omitempty"`                    // 当 Type 为 date/time 时使用
	EnumType       EnumType              `json:"enumType,omitempty" binding:"dive"`   // 当 Type 为 enum 时使用
	ArrayType      ArrayType             `json:"arrayType,omitempty" binding:"dive"`  // 当 Type 为 array 时使用
	ObjectType     map[string]ObjectType `json:"objectType,omitempty" binding:"dive"` // 当 Type 为 object 时使用
	ObjectRequired []string              `json:"objectRequired,omitempty"`            // 当 Type 为 object 时, 记录必填字段
	Current        interface{}           `yaml:"current" json:"current"`
	Expect         interface{}           `yaml:"expect" json:"expect"`
}

type deviceProperty struct {
	Name           string                `yaml:"name,omitempty" json:"name,omitempty"`
	Id             string                `yaml:"id,omitempty" json:"id,omitempty" binding:"nonzero"`
	Type           string                `yaml:"type,omitempty" json:"type,omitempty" binding:"data_type"`
	Mode           string                `yaml:"mode,omitempty" json:"mode,omitempty" binding:"oneof=ro rw"`
	Unit           string                `yaml:"unit,omitempty" json:"unit,omitempty"`
	Visitor        PropertyVisitor       `yaml:"visitor,omitempty" json:"visitor,omitempty"`
	Format         string                `json:"format,omitempty"`                    // 当 Type 为 date/time 时使用
	EnumType       EnumType              `json:"enumType,omitempty" binding:"dive"`   // 当 Type 为 enum 时使用
	ArrayType      ArrayType             `json:"arrayType,omitempty" binding:"dive"`  // 当 Type 为 array 时使用
	ObjectType     map[string]ObjectType `json:"objectType,omitempty" binding:"dive"` // 当 Type 为 object 时使用
	ObjectRequired []string              `json:"objectRequired,omitempty"`            // 当 Type 为 object 时, 记录必填字段
	Current        interface{}           `yaml:"current" json:"current"`
	Expect         interface{}           `yaml:"expect" json:"expect"`
}

func (dp *DeviceProperty) UnmarshalJSON(b []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	var prop deviceProperty
	if err := decoder.Decode(&prop); err != nil {
		return err
	}
	dp.Name = prop.Name
	dp.Id = prop.Id
	dp.Unit = prop.Unit
	dp.Type = prop.Type
	dp.Mode = prop.Mode
	dp.Visitor = prop.Visitor
	dp.Format = prop.Format
	dp.EnumType = prop.EnumType
	dp.ArrayType = prop.ArrayType
	dp.ObjectType = prop.ObjectType
	dp.ObjectRequired = prop.ObjectRequired
	dp.Current = prop.Current
	dp.Expect = prop.Expect
	return nil
}

type ShadowState struct {
	Report v1.Report `json:"report,omitempty"`
	Desire v1.Desire `json:"desire,omitempty"`
}

type ShadowMetadata struct {
	Report v1.Report `json:"report,omitempty"`
	Desire v1.Desire `json:"desire,omitempty"`
}

type DeviceConfig struct {
	Infos *DeviceInfo `yaml:"infos,omitempty" json:"infos,omitempty"`
	// props are not inherited from device model and have not preserved yet,
	// might be updated and saved when device model update in future
	Props  []DeviceProperty    `yaml:"props,omitempty" json:"props,omitempty"`
	Modbus *ModbusConfig       `yaml:"modbus,omitempty" json:"modbus,omitempty"`
	OpcUA  *OpcuaConfig        `yaml:"opcua,omitempty" json:"opcua,omitempty"`
	IEC104 *IEC104Config       `yaml:"iec104,omitempty" json:"iec104,omitempty"`
	Ipc    *IpcDeviceConfig    `yaml:"ipc,omitempty" json:"ipc,omitempty"`
	Custom *CustomDeviceConfig `yaml:"custom,omitempty" json:"custom,omitempty"`
}

type FullDeviceConfig struct {
	Infos *DeviceInfo `yaml:"infos,omitempty" json:"infos,omitempty"`
	Props interface{} `yaml:"props,omitempty" json:"props,omitempty"`
}

type DeviceConfigView struct {
	Modbus *ModbusConfig       `yaml:"modbus,omitempty" json:"modbus,omitempty"`
	OpcUA  *OpcuaConfig        `yaml:"opcua,omitempty" json:"opcua,omitempty"`
	Ipc    *IpcDeviceConfig    `yaml:"ipc,omitempty" json:"ipc,omitempty"`
	IEC104 *IEC104Config       `yaml:"iec104,omitempty" json:"iec104,omitempty"`
	Custom *CustomDeviceConfig `yaml:"custom,omitempty" json:"custom,omitempty"`
}

type ChannelConfig struct {
	ChannelId string         `yaml:"channelId,omitempty" json:"channelId,omitempty"`
	Modbus    *ModbusChannel `yaml:"modbus,omitempty" json:"modbus,omitempty"`
	OpcUA     *OpcuaChannel  `yaml:"opcua,omitempty" json:"opcua,omitempty"`
	IEC104    *IEC104Channel `yaml:"iec104,omitempty" json:"iec104,omitempty"`
}

type ModbusConfig struct {
	ChannelId string `yaml:"channelId,omitempty" json:"channelId,omitempty" binding:"required"`
	SlaveId   byte   `yaml:"slaveId,omitempty" json:"slaveId,omitempty" binding:"required"`
	Interval  int    `yaml:"interval,omitempty" json:"interval,omitempty"` // unit is second
}

type FullDriverConfig struct {
	Devices []DeviceInfo `yaml:"devices" json:"devices"`
	Driver  string       `yaml:"driver" json:"driver"`
}

type OpcuaConfig struct {
	ChannelId string `yaml:"channelId,omitempty" json:"channelId,omitempty" binding:"required"`
	Interval  int    `yaml:"interval" json:"interval,omitempty"`
	NsOffset  int    `yaml:"nsOffset" json:"nsOffset,omitempty"`
	IdOffset  int    `yaml:"idOffset" json:"idOffset,omitempty"`
}

type IEC104Config struct {
	ChannelId string `yaml:"channelId,omitempty" json:"channelId,omitempty" binding:"required"`
	Interval  int    `yaml:"interval" json:"interval,omitempty"`
	AIOffset  int    `yaml:"aiOffset" json:"aiOffset"`
	DIOffset  int    `yaml:"diOffset" json:"diOffset"`
	AOOffset  int    `yaml:"aoOffset" json:"aoOffset"`
	DOOffset  int    `yaml:"doOffset" json:"doOffset"`
	PIOffset  int    `yaml:"piOffset" json:"piOffset"`
}

type ModbusChannel struct {
	Tcp *TcpConfig `yaml:"tcp,omitempty" json:"tcp,omitempty"`
	Rtu *RtuConfig `yaml:"rtu,omitempty" json:"rtu,omitempty"`
}

type OpcuaChannel struct {
	ID          byte              `yaml:"id,omitempty" json:"id,omitempty"`
	Endpoint    string            `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	Timeout     int               `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Security    OpcuaSecurity     `yaml:"security,omitempty" json:"security,omitempty"`
	Auth        *OpcuaAuth        `yaml:"auth,omitempty" json:"auth,omitempty"`
	Certificate *OpcuaCertificate `yaml:"certificate,omitempty" json:"certificate,omitempty"`
}

type IEC104Channel struct {
	Protocol string `yaml:"protocol,omitempty" json:"protocol,omitempty"`
	Address  string `yaml:"address,omitempty" json:"address,omitempty" binding:"required"`
	Port     uint16 `yaml:"port,omitempty" json:"port,omitempty" binding:"required"`
}

type CustomChannel string

type CustomDeviceConfig string

type Duration struct {
	time.Duration
}

func (d Duration) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(d.String())
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

//

type DeviceInfo struct {
	Name    string `yaml:"name,omitempty" json:"name,omitempty"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
	// Deprecated: Use DeviceTopic instead.
	// Change from access template support
	Topic          `yaml:",inline" json:",inline"`
	DeviceModel    string        `yaml:"deviceModel,omitempty" json:"deviceModel,omitempty"`
	AccessTemplate string        `yaml:"accessTemplate,omitempty" json:"accessTemplate,omitempty"`
	DeviceTopic    DeviceTopic   `yaml:"deviceTopic,omitempty" json:"deviceTopic,omitempty"`
	AccessConfig   *AccessConfig `yaml:"accessConfig,omitempty" json:"accessConfig,omitempty"`
}

type DeviceTopic struct {
	Delta           mqtt2.QOSTopic `yaml:"delta,omitempty" json:"delta,omitempty"`
	Report          mqtt2.QOSTopic `yaml:"report,omitempty" json:"report,omitempty"`
	Event           mqtt2.QOSTopic `yaml:"event,omitempty" json:"event,omitempty"`
	Get             mqtt2.QOSTopic `yaml:"get,omitempty" json:"get,omitempty"`
	GetResponse     mqtt2.QOSTopic `yaml:"getResponse,omitempty" json:"getResponse,omitempty"`
	EventReport     mqtt2.QOSTopic `yaml:"eventReport,omitempty" json:"eventReport,omitempty"`
	PropertyGet     mqtt2.QOSTopic `yaml:"propertyGet,omitempty" json:"propertyGet,omitempty"`
	LifecycleReport mqtt2.QOSTopic `yaml:"lifecycleReport,omitempty" json:"lifecycleReport,omitempty"`
}

// Deprecated: Use DeviceTopic instead.
// Change from access template support
type Topic struct {
	Delta       mqtt2.QOSTopic `yaml:"delta,omitempty" json:"delta,omitempty"`
	Report      mqtt2.QOSTopic `yaml:"report,omitempty" json:"report,omitempty"`
	Event       mqtt2.QOSTopic `yaml:"event,omitempty" json:"event,omitempty"`
	Get         mqtt2.QOSTopic `yaml:"get,omitempty" json:"get,omitempty"`
	GetResponse mqtt2.QOSTopic `yaml:"getResponse,omitempty" json:"getResponse,omitempty"`
}

type PropertyGet struct {
	Properties []string `yaml:"properties,omitempty" json:"properties,omitempty"`
}

type DeviceShadow struct {
	Name   string    `yaml:"name,omitempty" json:"name,omitempty"`
	Report v1.Report `yaml:"report,omitempty" json:"report,omitempty"`
	Desire v1.Desire `yaml:"desire,omitempty" json:"desire,omitempty"`
}

type driverConfig struct {
	Devices []DeviceInfo `yaml:"devices,omitempty" json:"devices,omitempty"`
	Driver  string       `yaml:"driver,omitempty" json:"driver,omitempty"`
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
	case TypeBool, TypeString, TypeEnum, TypeArray, TypeTime, TypeDate, TypeObject:
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
