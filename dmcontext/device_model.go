package dmcontext

import (
	"bytes"
	"encoding/json"
	"reflect"
	"time"

	"github.com/jinzhu/copier"
)

type DeviceModelView struct {
	Name        string                 `json:"name,omitempty" binding:"omitempty,res_name"`
	Description string                 `json:"description,omitempty"`
	Protocol    string                 `json:"protocol,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Attributes  []DeviceModelAttribute `json:"attributes,omitempty"`
	Properties  []DeviceModelProperty  `json:"properties,omitempty"`
	CreateTime  time.Time              `json:"createTime,omitempty"`
	UpdateTime  time.Time              `json:"updateTime,omitempty"`
}

type DeviceModel struct {
	Name          string                 `json:"name,omitempty" binding:"omitempty,dev_model"`
	Namespace     string                 `json:"namespace,omitempty"`
	ProductSecret string                 `json:"productSecret,omitempty"`
	Version       string                 `json:"version,omitempty"`
	Description   string                 `json:"description,omitempty"`
	Protocol      string                 `json:"protocol,omitempty"`
	Labels        map[string]string      `json:"labels,omitempty"`
	Attributes    []DeviceModelAttribute `json:"attributes,omitempty" binding:"dive"`
	Properties    []DeviceModelProperty  `json:"properties,omitempty" binding:"dive"`
	CreateTime    time.Time              `json:"createTime,omitempty"`
	UpdateTime    time.Time              `json:"updateTime,omitempty"`
	Type          byte                   `json:"type"`
}

func (dm *DeviceModel) ToDeviceModelView() (*DeviceModelView, error) {
	res := new(DeviceModelView)
	err := copier.Copy(res, dm)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func EqualDeviceModel(old, new *DeviceModel) bool {
	if old.Name != new.Name || old.Protocol != new.Protocol || old.Description != new.Description {
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
	return true
}

type DeviceModelAttribute struct {
	Name         string      `json:"name,omitempty"`
	Id           string      `json:"id,omitempty"`
	Type         string      `json:"type,omitempty" binding:"data_type"`
	Unit         string      `json:"unit,omitempty"`
	DefaultValue interface{} `json:"defaultValue,omitempty"`
	Required     bool        `json:"required"`
}

type DeviceModelProperty struct {
	Name           string                `json:"name,omitempty"`
	Id             string                `json:"id,omitempty" binding:"required"`
	Type           string                `json:"type,omitempty" binding:"data_plus_type"`
	Mode           string                `json:"mode,omitempty" binding:"oneof=ro rw"`
	Unit           string                `json:"unit,omitempty"`
	Format         string                `json:"format,omitempty"`                    // 当 Type 为 date/time 时使用
	EnumType       EnumType              `json:"enumType,omitempty" binding:"dive"`   // 当 Type 为 enum 时使用
	ArrayType      ArrayType             `json:"arrayType,omitempty" binding:"dive"`  // 当 Type 为 array 时使用
	ObjectType     map[string]ObjectType `json:"objectType,omitempty" binding:"dive"` // 当 Type 为 object 时使用
	ObjectRequired []string              `json:"objectRequired,omitempty"`            // 当 Type 为 object 时, 记录必填字段
	Visitor        PropertyVisitor       `json:"visitor,omitempty" binding:"dive"`
}

type DeviceModelPropertyYaml struct {
	Name           string                `json:"name,omitempty" yaml:"name,omitempty"`
	Type           string                `json:"type,omitempty" yaml:"type,omitempty"`
	Mode           string                `json:"mode,omitempty" yaml:"mode,omitempty"`
	Format         string                `json:"format,omitempty" yaml:"format,omitempty"`                        // 当 Type 为 date/time 时使用
	EnumType       EnumType              `json:"enumType,omitempty" yaml:"enumType,omitempty" binding:"dive"`     // 当 Type 为 enum 时使用
	ArrayType      ArrayType             `json:"arrayType,omitempty" yaml:"arrayType,omitempty" binding:"dive"`   // 当 Type 为 array 时使用
	ObjectType     map[string]ObjectType `json:"objectType,omitempty" yaml:"objectType,omitempty" binding:"dive"` // 当 Type 为 object 时使用
	ObjectRequired []string              `json:"objectRequired,omitempty" yaml:"objectRequired,omitempty"`        // 当 Type 为 object 时, 记录必填字段
}

type deviceModelAttribute struct {
	Name         string      `json:"name,omitempty"`
	Id           string      `json:"id,omitempty"`
	Type         string      `json:"type,omitempty" binding:"data_type"`
	Unit         string      `json:"unit,omitempty"`
	DefaultValue interface{} `json:"defaultValue,omitempty"`
	Required     bool        `json:"required"`
}

func (da *DeviceModelAttribute) UnmarshalJSON(b []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	var attr deviceModelAttribute
	if err := decoder.Decode(&attr); err != nil {
		return err
	}
	da.Name = attr.Name
	da.Id = attr.Id
	da.Type = attr.Type
	da.Unit = attr.Unit
	da.DefaultValue = attr.DefaultValue
	da.Required = attr.Required
	return nil
}
