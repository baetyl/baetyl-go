package dmcontext

import (
	mqtt2 "github.com/baetyl/baetyl-go/v2/mqtt"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
)

type DeviceInfo struct {
	Name    string `yaml:"name,omitempty" json:"name,omitempty"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
	Topic   `yaml:",inline" json:",inline"`
}

type Topic struct {
	Delta       mqtt2.QOSTopic `yaml:"delta,omitempty" json:"delta,omitempty"`
	Report      mqtt2.QOSTopic `yaml:"report,omitempty" json:"report,omitempty"`
	Event       mqtt2.QOSTopic `yaml:"event,omitempty" json:"event,omitempty"`
	Get         mqtt2.QOSTopic `yaml:"get,omitempty" json:"get,omitempty"`
	GetResponse mqtt2.QOSTopic `yaml:"getResponse,omitempty" json:"getResponse,omitempty"`
}

type DeviceProperty struct {
	Name    string          `yaml:"name,omitempty" json:"name,omitempty"`
	Type    string          `yaml:"type,omitempty" json:"type,omitempty" validate:"regexp=^(int16|int32|int64|float32|float64|string|bool)?$"`
	Mode    string          `yaml:"mode,omitempty" json:"mode,omitempty" validate:"regexp=^(ro|rw)?$"`
	Visitor PropertyVisitor `yaml:"visitor,omitempty" json:"visitor,omitempty"`
}

type PropertyVisitor struct {
	Modbus *ModbusVisitor `yaml:"modbus,omitempty" json:"modbus,omitempty"`
	Opcua  *OpcuaVisitor  `yaml:"opcua,omitempty" json:"opcua,omitempty"`
	Custom *CustomVisitor `yaml:"custom,omitempty" json:"custom,omitempty"`
}

type ModbusVisitor struct {
	Function     byte    `yaml:"function" json:"function" validate:"min=1,max=4"`
	Address      string  `yaml:"address" json:"address"`
	Quantity     uint16  `yaml:"quantity" json:"quantity"`
	Type         string  `yaml:"type,omitempty" json:"type,omitempty" validate:"regexp=^(int16|int32|int64|float32|float64|string|bool)?$"`
	Scale        float64 `yaml:"scale" json:"scale"`
	SwapByte     bool    `yaml:"swapByte" json:"swapByte"`
	SwapRegister bool    `yaml:"swapRegister" json:"swapRegister"`
}

type OpcuaVisitor struct {
	NodeID string `yaml:"nodeid,omitempty" json:"nodeid,omitempty"`
	Type   string `yaml:"type,omitempty" json:"type,omitempty" validate:"regexp=^(int16|int32|int64|float32|float64|string|bool)?$"`
}

type CustomVisitor string

type Event struct {
	Type    string      `yaml:"type,omitempty" json:"type,omitempty"`
	Payload interface{} `yaml:"payload,omitempty" json:"payload,omitempty"`
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
