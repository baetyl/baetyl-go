package dmcontext

import (
	"time"
)

const (
	MappingNone      = "none"
	MappingValue     = "value"
	MappingCalculate = "calculate"
	OpcuaIdTypeI     = "i"
	OpcuaIdTypeS     = "s"
	OpcuaIdTypeG     = "g"
	OpcuaIdTypeB     = "b"
)

type IpcDeviceConfig struct {
	Name          string `yaml:"name" json:"name"`
	StreamAddress string `yaml:"streamAddress" json:"streamAddress"`
	ServiceName   string `yaml:"serviceName" json:"serviceName"`
	ResultTopic   string `yaml:"resultTopic" json:"resultTopic"`
	AgentEnable   bool   `yaml:"agentEnable" json:"agentEnable"`                         // republish rtsp
	System        bool   `yaml:"system" json:"system"`                                   // use baetyl-ipc-cloud or user self defined
	RemoteAddress string `yaml:"remoteAddress,omitempty" json:"remoteAddress,omitempty"` // republish address
	IP            string `yaml:"ip" json:"ip"`                                           // onvif config
	Port          int32  `yaml:"port" json:"port" default:"80"`
	Username      string `yaml:"username" json:"username"`
	Password      string `yaml:"password" json:"password"`
}

type IpcServiceConfig struct {
	Name        string  `yaml:"name,omitempty" json:"name,omitempty"`
	FPS         float64 `yaml:"fps,omitempty" json:"fps,omitempty"`
	ImageFormat string  `yaml:"imageFormat,omitempty" json:"imageFormat,omitempty" default:"jpg"`
	Scale       struct {
		Enable bool `yaml:"enable,omitempty" json:"enable,omitempty"`
		Height int  `yaml:"height,omitempty" json:"height,omitempty"`
		Width  int  `yaml:"width,omitempty" json:"width,omitempty"`
	} `yaml:"scale,omitempty" json:"scale,omitempty"`
	Address string `yaml:"address,omitempty" json:"address,omitempty"`
	Request struct {
		Params map[string]string `yaml:"params,omitempty" json:"params,omitempty"`
	} `yaml:"request,omitempty" json:"request,omitempty"`
	Body struct {
		Content   string                 `yaml:"content,omitempty" json:"content,omitempty"`
		ImageType string                 `yaml:"imageType,omitempty" json:"imageType,omitempty"`
		ImageName string                 `yaml:"imageName,omitempty" json:"imageName,omitempty"`
		Params    map[string]interface{} `yaml:"params,omitempty" json:"params,omitempty"`
	} `yaml:"body,omitempty" json:"body,omitempty"`
	Upload    bool   `yaml:"upload,omitempty" json:"upload,omitempty"`
	Cache     bool   `yaml:"cache,omitempty" json:"cache,omitempty"`
	CachePath string `yaml:"cachePath,omitempty" json:"cachePath,omitempty" default:"var/lib/baetyl/image"`
	CacheTime int    `yaml:"cacheTime,omitempty" json:"cacheTime,omitempty" default:"3" binding:"omitempty,min=3,max=180"` // 图片清理时间，默认3min
}

type ReportProperty struct {
	Time  time.Time   `yaml:"time,omitempty" json:"time,omitempty"`
	Value interface{} `yaml:"value,omitempty" json:"value,omitempty"`
}

type Event struct {
	Type    string      `yaml:"type,omitempty" json:"type,omitempty"`
	Payload interface{} `yaml:"payload,omitempty" json:"payload,omitempty"`
}

type EnumType struct {
	Type   string      `yaml:"type,omitempty" json:"type,omitempty" binding:"enum_type"`
	Values []EnumValue `yaml:"values,omitempty" json:"values,omitempty" binding:"lte=20"`
}

type EnumValue struct {
	Name        string `yaml:"name,omitempty" json:"name,omitempty"`
	Value       string `yaml:"value,omitempty" json:"value,omitempty"`
	DisplayName string `yaml:"displayName,omitempty" json:"displayName,omitempty"`
}

type ArrayType struct {
	Type   string `yaml:"type,omitempty" json:"type,omitempty" binding:"data_type"`
	Min    int    `yaml:"min,omitempty" json:"min,omitempty" binding:"gte=0"`
	Max    int    `yaml:"max,omitempty" json:"max,omitempty" binding:"lte=20"`
	Format string `yaml:"format,omitempty" json:"format,omitempty"` // 当 Type 为 date/time 时使用
}

type ObjectType struct {
	DisplayName string `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	Type        string `yaml:"type,omitempty" json:"type,omitempty" binding:"data_type"`
	Format      string `yaml:"format,omitempty" json:"format,omitempty"` // 当 Type 为 date/time 时使用
}

type PropertyVisitor struct {
	Modbus *ModbusVisitor `yaml:"modbus,omitempty" json:"modbus,omitempty"`
	Opcua  *OpcuaVisitor  `yaml:"opcua,omitempty" json:"opcua,omitempty"`
	Opcda  *OpcdaVisitor  `yaml:"opcda,omitempty" json:"opcda,omitempty"`
	Bacnet *BacnetVisitor `yaml:"bacnet,omitempty" json:"bacnet,omitempty"`
	IEC104 *IEC104Visitor `yaml:"iec104,omitempty" json:"iec104,omitempty"`
	Custom *CustomVisitor `yaml:"custom,omitempty" json:"custom,omitempty"`
}

type IEC104Visitor struct {
	PointNum  uint   `yaml:"pointNum" json:"pointNum"`
	PointType string `yaml:"pointType,omitempty" json:"pointType,omitempty"`
	Type      string `yaml:"type,omitempty" json:"type,omitempty" binding:"oneof=float32 bool"`
}

type ModbusVisitor struct {
	Function     byte    `yaml:"function" json:"function" binding:"min=1,max=4"`
	Address      string  `yaml:"address" json:"address"`
	Quantity     uint16  `yaml:"quantity" json:"quantity"`
	Type         string  `yaml:"type,omitempty" json:"type,omitempty" binding:"data_type"`
	Unit         string  `yaml:"unit,omitempty" json:"unit,omitempty"`
	Scale        float64 `yaml:"scale" json:"scale"`
	SwapByte     bool    `yaml:"swapByte" json:"swapByte"`
	SwapRegister bool    `yaml:"swapRegister" json:"swapRegister"`
}

type OpcuaVisitor struct {
	// Deprecated: Use NsBase, IdBase, OpcuaAccessConfig.NsOffset, OpcuaAccessConfig.IdOffset instead.
	// Change from access template support
	NodeID string `yaml:"nodeid,omitempty" json:"nodeid,omitempty"`
	Type   string `yaml:"type,omitempty" json:"type,omitempty" binding:"data_type"`
	NsBase int    `yaml:"nsBase,omitempty" json:"nsBase,omitempty"`
	IdBase string `yaml:"idBase,omitempty" json:"idBase,omitempty"`
	IdType string `yaml:"idType,omitempty" json:"idType,omitempty" binding:"omitempty,oneof=i s g b"`
}

type OpcdaVisitor struct {
	Datapath string `yaml:"datapath,omitempty" json:"datapath,omitempty"`
	Type     string `yaml:"type,omitempty" json:"type,omitempty" binding:"data_type"`
}

type BacnetVisitor struct {
	Type                 string `yaml:"type,omitempty" json:"type,omitempty" binding:"data_type"`
	BacnetType           uint   `yaml:"bacnetType,omitempty" json:"bacnetType,omitempty"`
	BacnetAddress        uint   `yaml:"bacnetAddress,omitempty" json:"bacnetAddress,omitempty"`
	ApplicationTagNumber byte   `yaml:"applicationTagNumber" json:"applicationTagNumber"`
}

type CustomVisitor string
