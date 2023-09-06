package dmcontext

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
)

type DeviceProperty struct {
	Name           string                `yaml:"name,omitempty" json:"name,omitempty"`
	ID             string                `yaml:"id,omitempty" json:"id,omitempty" binding:"nonzero"`
	Type           string                `yaml:"type,omitempty" json:"type,omitempty" binding:"data_type"`
	Mode           string                `yaml:"mode,omitempty" json:"mode,omitempty" binding:"oneof=ro rw"`
	Unit           string                `yaml:"unit,omitempty" json:"unit,omitempty"`
	Visitor        PropertyVisitor       `yaml:"visitor,omitempty" json:"visitor,omitempty"`
	Format         string                `yaml:"format,omitempty" json:"format,omitempty"`                        // 当 Type 为 date/time 时使用
	EnumType       EnumType              `yaml:"enumType,omitempty" json:"enumType,omitempty" binding:"dive"`     // 当 Type 为 enum 时使用
	ArrayType      ArrayType             `yaml:"arrayType,omitempty" json:"arrayType,omitempty" binding:"dive"`   // 当 Type 为 array 时使用
	ObjectType     map[string]ObjectType `yaml:"objectType,omitempty" json:"objectType,omitempty" binding:"dive"` // 当 Type 为 object 时使用
	ObjectRequired []string              `yaml:"objectRequired,omitempty" json:"objectRequired,omitempty"`        // 当 Type 为 object 时, 记录必填字段
	Current        any                   `yaml:"current" json:"current"`
	Expect         any                   `yaml:"expect" json:"expect"`
}

type deviceProperty struct {
	Name           string                `yaml:"name,omitempty" json:"name,omitempty"`
	Id             string                `yaml:"id,omitempty" json:"id,omitempty" binding:"nonzero"`
	Type           string                `yaml:"type,omitempty" json:"type,omitempty" binding:"data_type"`
	Mode           string                `yaml:"mode,omitempty" json:"mode,omitempty" binding:"oneof=ro rw"`
	Unit           string                `yaml:"unit,omitempty" json:"unit,omitempty"`
	Visitor        PropertyVisitor       `yaml:"visitor,omitempty" json:"visitor,omitempty"`
	Format         string                `yaml:"format,omitempty" json:"format,omitempty"`                        // 当 Type 为 date/time 时使用
	EnumType       EnumType              `yaml:"enumType,omitempty" json:"enumType,omitempty" binding:"dive"`     // 当 Type 为 enum 时使用
	ArrayType      ArrayType             `yaml:"arrayType,omitempty" json:"arrayType,omitempty" binding:"dive"`   // 当 Type 为 array 时使用
	ObjectType     map[string]ObjectType `yaml:"objectType,omitempty" json:"objectType,omitempty" binding:"dive"` // 当 Type 为 object 时使用
	ObjectRequired []string              `yaml:"objectRequired,omitempty" json:"objectRequired,omitempty"`        // 当 Type 为 object 时, 记录必填字段
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
	dp.ID = prop.Id
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

type DeviceInfo struct {
	Name           string        `yaml:"name,omitempty" json:"name,omitempty"`
	Version        string        `yaml:"version,omitempty" json:"version,omitempty"`
	DeviceModel    string        `yaml:"deviceModel,omitempty" json:"deviceModel,omitempty"`
	AccessTemplate string        `yaml:"accessTemplate,omitempty" json:"accessTemplate,omitempty"`
	DeviceTopic    DeviceTopic   `yaml:"deviceTopic,omitempty" json:"deviceTopic,omitempty"`
	AccessConfig   *AccessConfig `yaml:"accessConfig,omitempty" json:"accessConfig,omitempty"`
}

type DeviceTopic struct {
	Delta           mqtt.QOSTopic `yaml:"delta,omitempty" json:"delta,omitempty"`
	Report          mqtt.QOSTopic `yaml:"report,omitempty" json:"report,omitempty"`
	Event           mqtt.QOSTopic `yaml:"event,omitempty" json:"event,omitempty"`
	Get             mqtt.QOSTopic `yaml:"get,omitempty" json:"get,omitempty"`
	GetResponse     mqtt.QOSTopic `yaml:"getResponse,omitempty" json:"getResponse,omitempty"`
	EventReport     mqtt.QOSTopic `yaml:"eventReport,omitempty" json:"eventReport,omitempty"`
	PropertyGet     mqtt.QOSTopic `yaml:"propertyGet,omitempty" json:"propertyGet,omitempty"`
	LifecycleReport mqtt.QOSTopic `yaml:"lifecycleReport,omitempty" json:"lifecycleReport,omitempty"`
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

func (c *DmCtx) ParsePropertyValues(driverName string, device *DeviceInfo, props map[string]any) (map[string]any, error) {
	res := make(map[string]any)
	vals, ok := c.deviceModels[driverName][device.DeviceModel]
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

func parsePropertyValue(tpy string, val any) (any, error) {
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

func ParsePropertyKeys(v any) ([]string, error) {
	properties, ok := v.([]any)
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
