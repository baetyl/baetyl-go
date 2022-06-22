package dmcontext

import (
	"time"

	mqtt2 "github.com/baetyl/baetyl-go/v2/mqtt"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
)

type DeviceInfo struct {
	Name           string        `yaml:"name,omitempty" json:"name,omitempty"`
	Version        string        `yaml:"version,omitempty" json:"version,omitempty"`
	DeviceModel    string        `yaml:"deviceModel,omitempty" json:"deviceModel,omitempty"`
	AccessTemplate string        `yaml:"accessTemplate,omitempty" json:"accessTemplate,omitempty"`
	Topics         Topic         `yaml:"topics,omitempty" json:"topics,omitempty"`
	AccessConfig   *AccessConfig `yaml:"accessConfig,omitempty" json:"accessConfig,omitempty"`
}

type Topic struct {
	Delta       mqtt2.QOSTopic `yaml:"delta,omitempty" json:"delta,omitempty"`
	Report      mqtt2.QOSTopic `yaml:"report,omitempty" json:"report,omitempty"`
	Event       mqtt2.QOSTopic `yaml:"event,omitempty" json:"event,omitempty"`
	Get         mqtt2.QOSTopic `yaml:"get,omitempty" json:"get,omitempty"`
	GetResponse mqtt2.QOSTopic `yaml:"getResponse,omitempty" json:"getResponse,omitempty"`
}

func (a *AccessConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var acc accessConfig
	if err := unmarshal(&acc); err == nil {
		a.Modbus = acc.Modbus
		a.Opcua = acc.Opcua
		a.Custom = acc.Custom
		// for backward compatibility
		if a.Modbus == nil && a.Opcua == nil && a.Custom == nil {
			var modbus ModbusAccessConfig
			if err = unmarshal(&modbus); err == nil {
				a.Modbus = &modbus
				return nil
			}
		}
		return nil
	}
	// for backward compatibility
	var custom CustomAccessConfig
	if err := unmarshal(&custom); err != nil {
		return err
	}
	a.Custom = &custom
	return nil
}

type AccessConfig struct {
	Modbus *ModbusAccessConfig `yaml:"modbus,omitempty" json:"modbus,omitempty"`
	Opcua  *OpcuaAccessConfig  `yaml:"opcua,omitempty" json:"opcua,omitempty"`
	Custom *CustomAccessConfig `yaml:"custom,omitempty" json:"custom,omitempty"`
}

type accessConfig struct {
	Modbus *ModbusAccessConfig `yaml:"modbus,omitempty" json:"modbus,omitempty"`
	Opcua  *OpcuaAccessConfig  `yaml:"opcua,omitempty" json:"opcua,omitempty"`
	Custom *CustomAccessConfig `yaml:"custom,omitempty" json:"custom,omitempty"`
}

type ModbusAccessConfig struct {
	Id          byte          `yaml:"id,omitempty" json:"id,omitempty"`
	Interval    time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout     time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" default:"10s"`
	IdleTimeout time.Duration `yaml:"idletimeout,omitempty" json:"idletimeout,omitempty" default:"1m"`
	Tcp         *TcpConfig    `yaml:"tcp,omitempty" json:"tcp,omitempty"`
	Rtu         *RtuConfig    `yaml:"rtu,omitempty" json:"rtu,omitempty"`
}

type TcpConfig struct {
	Address string `yaml:"address,omitempty" json:"address,omitempty" validate:"required"`
	Port    uint16 `yaml:"port,omitempty" json:"port,omitempty" validate:"required"`
}

type RtuConfig struct {
	Port     string `yaml:"port,omitempty" json:"port,omitempty" validate:"required"`
	BaudRate int    `yaml:"baudrate,omitempty" json:"baudrate,omitempty" default:"19200"`
	Parity   string `yaml:"parity,omitempty" json:"parity,omitempty" default:"E" validate:"regexp=^(E|N|O)?$"`
	DataBit  int    `yaml:"databit,omitempty" json:"databit,omitempty" default:"8" validate:"min=5, max=8"`
	StopBit  int    `yaml:"stopbit,omitempty" json:"stopbit,omitempty" default:"1" validate:"min=1, max=2"`
}

type OpcuaAccessConfig struct {
	Id          byte              `yaml:"id,omitempty" json:"id,omitempty"`
	Endpoint    string            `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	Interval    time.Duration     `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout     time.Duration     `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Security    OpcuaSecurity     `yaml:"security,omitempty" json:"security,omitempty"`
	Auth        *OpcuaAuth        `yaml:"auth,omitempty" json:"auth,omitempty"`
	Certificate *OpcuaCertificate `yaml:"certificate,omitempty" json:"certificate,omitempty"`
}

type OpcuaSecurity struct {
	Policy string `yaml:"policy,omitempty" json:"policy,omitempty"`
	Mode   string `yaml:"mode,omitempty" json:"mode,omitempty"`
}

type OpcuaAuth struct {
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
}

type OpcuaCertificate struct {
	Cert string `yaml:"cert,omitempty" json:"cert,omitempty"`
	Key  string `yaml:"key,omitempty" json:"key,omitempty"`
}

type CustomAccessConfig string

type DeviceProperty struct {
	Name    string          `yaml:"name,omitempty" json:"name,omitempty"`
	Id      string          `yaml:"id,omitempty" json:"id,omitempty"`
	Type    string          `yaml:"type,omitempty" json:"type,omitempty" validate:"regexp=^(int16|int32|int64|float32|float64|string|bool)?$"`
	Mode    string          `yaml:"mode,omitempty" json:"mode,omitempty" validate:"regexp=^(ro|rw)?$"`
	Unit    string          `yaml:"unit,omitempty" json:"unit,omitempty"`
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

type AccessTemplate struct {
	Name       string           `yaml:"name,omitempty" json:"name,omitempty"`
	Version    string           `yaml:"version,omitempty" json:"version,omitempty"`
	Properties []DeviceProperty `yaml:"properties,omitempty" json:"properties,omitempty"`
	Mappings   []ModelMapping   `yaml:"mappings,omitempty" json:"mappings,omitempty"`
}

type ModelMapping struct {
	Attribute  string `yaml:"attribute,omitempty" json:"attribute,omitempty"`
	Type       string `yaml:"type,omitempty" json:"type,omitempty"`
	Expression string `yaml:"expression,omitempty" json:"expression,omitempty"`
}

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
