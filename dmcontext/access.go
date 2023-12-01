package dmcontext

import (
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
)

var (
	ErrUnknownPropertyID     = errors.New("unknown property id")
	ErrConfigIDNotExist      = errors.New("config id not exist")
	ErrPropertyValueNotExist = errors.New("prop value not exist")
)

type CustomAccessConfig string

type AccessTemplate struct {
	Name       string           `yaml:"name,omitempty" json:"name,omitempty"`
	Version    string           `yaml:"version,omitempty" json:"version,omitempty"`
	Properties []DeviceProperty `yaml:"properties,omitempty" json:"properties,omitempty"`
	Mappings   []ModelMapping   `yaml:"mappings,omitempty" json:"mappings,omitempty"`
}

type ModelMapping struct {
	Attribute  string  `yaml:"attribute,omitempty" json:"attribute,omitempty"`
	Type       string  `yaml:"type,omitempty" json:"type,omitempty" default:"none"`
	Expression string  `yaml:"expression,omitempty" json:"expression,omitempty"`
	Precision  int     `yaml:"precision" json:"precision"`
	Deviation  float64 `yaml:"deviation" json:"deviation"`
	SilentWin  int     `yaml:"silentWin" json:"silentWin"`
}

type AccessConfig struct {
	Modbus *ModbusAccessConfig `yaml:"modbus,omitempty" json:"modbus,omitempty"`
	Opcua  *OpcuaAccessConfig  `yaml:"opcua,omitempty" json:"opcua,omitempty"`
	Opcda  *OpcdaAccessConfig  `yaml:"opcda,omitempty" json:"opcda,omitempty"`
	Bacnet *BacnetAccessConfig `yaml:"bacnet,omitempty" json:"bacnet,omitempty"`
	IEC104 *IEC104AccessConfig `yaml:"iec104,omitempty" json:"iec104,omitempty"`
	Custom *CustomAccessConfig `yaml:"custom,omitempty" json:"custom,omitempty"`
}

type accessConfig struct {
	Modbus *ModbusAccessConfig `yaml:"modbus,omitempty" json:"modbus,omitempty"`
	Opcua  *OpcuaAccessConfig  `yaml:"opcua,omitempty" json:"opcua,omitempty"`
	Opcda  *OpcdaAccessConfig  `yaml:"opcda,omitempty" json:"opcda,omitempty"`
	Bacnet *BacnetAccessConfig `yaml:"bacnet,omitempty" json:"bacnet,omitempty"`
	IEC104 *IEC104AccessConfig `yaml:"iec104,omitempty" json:"iec104,omitempty"`
	Custom *CustomAccessConfig `yaml:"custom,omitempty" json:"custom,omitempty"`
}

type ModbusAccessConfig struct {
	ID          byte          `yaml:"id,omitempty" json:"id,omitempty"`
	Interval    time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout     time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" default:"10s"`
	IdleTimeout time.Duration `yaml:"idletimeout,omitempty" json:"idletimeout,omitempty" default:"1m"`
	TCP         *TCPConfig    `yaml:"tcp,omitempty" json:"tcp,omitempty"`
	RTU         *RTUConfig    `yaml:"rtu,omitempty" json:"rtu,omitempty"`
}

type IEC104AccessConfig struct {
	ID       byte          `yaml:"id,omitempty" json:"id,omitempty"`
	Interval time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
	Endpoint string        `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	AIOffset uint16        `yaml:"aiOffset,omitempty" json:"aiOffset,omitempty"`
	DIOffset uint16        `yaml:"diOffset,omitempty" json:"diOffset,omitempty"`
	AOOffset uint16        `yaml:"aoOffset,omitempty" json:"aoOffset,omitempty"`
	DOOffset uint16        `yaml:"doOffset,omitempty" json:"doOffset,omitempty"`
}

type TCPConfig struct {
	Address string `yaml:"address,omitempty" json:"address,omitempty" binding:"required"`
	Port    uint16 `yaml:"port,omitempty" json:"port,omitempty" binding:"required"`
}

type RTUConfig struct {
	Port     string `yaml:"port,omitempty" json:"port,omitempty" binding:"required"`
	BaudRate int    `yaml:"baudrate,omitempty" json:"baudrate,omitempty" default:"19200"`
	Parity   string `yaml:"parity,omitempty" json:"parity,omitempty" default:"E" binding:"oneof=E N O"`
	DataBit  int    `yaml:"databit,omitempty" json:"databit,omitempty" default:"8" binding:"min=5,max=8"`
	StopBit  int    `yaml:"stopbit,omitempty" json:"stopbit,omitempty" default:"1" binding:"min=1,max=2"`
}

type OpcuaAccessConfig struct {
	ID          byte              `yaml:"id,omitempty" json:"id,omitempty"`
	Endpoint    string            `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	Subscribe   bool              `yaml:"subscribe,omitempty" json:"subscribe,omitempty"`
	Interval    time.Duration     `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout     time.Duration     `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Security    OpcuaSecurity     `yaml:"security,omitempty" json:"security,omitempty"`
	Auth        *OpcuaAuth        `yaml:"auth,omitempty" json:"auth,omitempty"`
	Certificate *OpcuaCertificate `yaml:"certificate,omitempty" json:"certificate,omitempty"`
	NsOffset    int               `yaml:"nsOffset,omitempty" json:"nsOffset,omitempty"`
	IDOffset    int               `yaml:"idOffset,omitempty" json:"idOffset,omitempty"`
}

type OpcdaAccessConfig struct {
	Host      string `yaml:"host,omitempty" json:"host,omitempty"`
	ClsID     string `yaml:"clsid,omitempty" json:"clsid,omitempty"`
	ProgramID string `yaml:"programid,omitempty" json:"programid,omitempty"`
	UserName  string `yaml:"username,omitempty" json:"username,omitempty"`
	Password  string `yaml:"password,omitempty" json:"password,omitempty"`
	Interval  int    `yaml:"interval,omitempty" json:"interval,omitempty"`
}

type BacnetAccessConfig struct {
	ID            byte          `yaml:"id,omitempty" json:"id,omitempty"`
	Interval      time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
	DeviceID      uint32        `yaml:"deviceId,omitempty" json:"deviceId,omitempty"`
	AddressOffset uint          `yaml:"addressOffset,omitempty" json:"addressOffset,omitempty"`
	Address       string        `yaml:"address,omitempty" json:"address,omitempty"`
	Port          int           `yaml:"port,omitempty" json:"port,omitempty"`
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

func (a *AccessConfig) UnmarshalYAML(unmarshal func(any) error) error {
	var acc accessConfig
	if err := unmarshal(&acc); err == nil {
		a.Modbus = acc.Modbus
		a.Opcua = acc.Opcua
		a.Opcda = acc.Opcda
		a.Bacnet = acc.Bacnet
		a.IEC104 = acc.IEC104
		a.Custom = acc.Custom
		return nil
	}
	return nil
}

func GetMappingName(id string, template *AccessTemplate) (string, error) {
	var name string

	for _, deviceProperty := range template.Properties {
		if id == deviceProperty.ID {
			name = deviceProperty.Name
			break
		}
	}
	if name == "" {
		return "", ErrUnknownPropertyID
	}
	return name, nil
}

func GetConfigIDByModelName(name string, template *AccessTemplate) (string, error) {
	for _, modelMapping := range template.Mappings {
		if modelMapping.Attribute == name {
			ids, err := ParseExpression(modelMapping.Expression)
			if err != nil {
				return "", err
			}
			if len(ids) > 0 {
				return ids[0][1:], nil
			}
		}
	}
	return "", ErrConfigIDNotExist
}

func GetPropValueByModelName(name string, val any, template *AccessTemplate) (any, error) {
	for _, modelMapping := range template.Mappings {
		if modelMapping.Attribute == name {
			if modelMapping.Type == MappingValue {
				return val, nil
			}
			value, err := ParseValueToFloat64(val)
			if err != nil {
				return nil, err
			}
			propVal, err := SolveExpression(modelMapping.Expression, value)
			if err != nil {
				return nil, err
			}
			return propVal, nil
		}
	}
	return nil, ErrPropertyValueNotExist
}
