package dmcontext

import "time"

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
	Precision  int     `yaml:"precision" json:"precision" default:"2"`
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
	Id          byte          `yaml:"id,omitempty" json:"id,omitempty"`
	Interval    time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout     time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" default:"10s"`
	IdleTimeout time.Duration `yaml:"idletimeout,omitempty" json:"idletimeout,omitempty" default:"1m"`
	Tcp         *TcpConfig    `yaml:"tcp,omitempty" json:"tcp,omitempty"`
	Rtu         *RtuConfig    `yaml:"rtu,omitempty" json:"rtu,omitempty"`
}

type IEC104AccessConfig struct {
	Id       byte          `yaml:"id,omitempty" json:"id,omitempty"`
	Interval time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
	Endpoint string        `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	AIOffset uint16        `yaml:"aiOffset,omitempty" json:"aiOffset,omitempty"`
	DIOffset uint16        `yaml:"diOffset,omitempty" json:"diOffset,omitempty"`
	AOOffset uint16        `yaml:"aoOffset,omitempty" json:"aoOffset,omitempty"`
	DOOffset uint16        `yaml:"doOffset,omitempty" json:"doOffset,omitempty"`
}

type TcpConfig struct {
	Address string `yaml:"address,omitempty" json:"address,omitempty" binding:"required"`
	Port    uint16 `yaml:"port,omitempty" json:"port,omitempty" binding:"required"`
}

type RtuConfig struct {
	Port     string `yaml:"port,omitempty" json:"port,omitempty" binding:"required"`
	BaudRate int    `yaml:"baudrate,omitempty" json:"baudrate,omitempty" default:"19200"`
	Parity   string `yaml:"parity,omitempty" json:"parity,omitempty" default:"E" binding:"oneof=E N O"`
	DataBit  int    `yaml:"databit,omitempty" json:"databit,omitempty" default:"8" binding:"min=5,max=8"`
	StopBit  int    `yaml:"stopbit,omitempty" json:"stopbit,omitempty" default:"1" binding:"min=1,max=2"`
}

type OpcuaAccessConfig struct {
	Id          byte              `yaml:"id,omitempty" json:"id,omitempty"`
	Endpoint    string            `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	Interval    time.Duration     `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout     time.Duration     `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Security    OpcuaSecurity     `yaml:"security,omitempty" json:"security,omitempty"`
	Auth        *OpcuaAuth        `yaml:"auth,omitempty" json:"auth,omitempty"`
	Certificate *OpcuaCertificate `yaml:"certificate,omitempty" json:"certificate,omitempty"`
	NsOffset    int               `yaml:"nsOffset,omitempty" json:"nsOffset,omitempty"`
	IdOffset    int               `yaml:"idOffset,omitempty" json:"idOffset,omitempty"`
}

type OpcdaAccessConfig struct {
	Server   string        `yaml:"server" json:"server"`
	Host     string        `yaml:"host" json:"host"`
	Group    string        `yaml:"group" json:"group"`
	Interval time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
}

type BacnetAccessConfig struct {
	Id            byte          `yaml:"id,omitempty" json:"id,omitempty"`
	Interval      time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
	DeviceId      uint32        `yaml:"deviceId,omitempty" json:"deviceId,omitempty"`
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

func (a *AccessConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var acc accessConfig
	if err := unmarshal(&acc); err == nil {
		a.Modbus = acc.Modbus
		a.Opcua = acc.Opcua
		a.Opcda = acc.Opcda
		a.Bacnet = acc.Bacnet
		a.IEC104 = acc.IEC104
		a.Custom = acc.Custom
		// for backward compatibility
		if a.Modbus == nil && a.Opcua == nil && a.Custom == nil && a.IEC104 == nil && a.Opcda == nil && a.Bacnet == nil {
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
