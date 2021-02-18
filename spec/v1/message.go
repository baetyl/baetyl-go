package v1

// MessageKind message kind
type MessageKind string

const (
	// MessageReport report message kind
	MessageReport MessageKind = "report"
	// MessageDesire desire message kind
	MessageDesire MessageKind = "desire"
	// MessageKeep keep alive message kind
	MessageKeep MessageKind = "keepalive"
	// MessageCMD command message kind
	MessageCMD MessageKind = "cmd"
	// MessageData data message kind
	MessageData MessageKind = "data"
	// MessageError error message kind
	MessageError MessageKind = "error"
	// Message response message kind
	MessageResponse MessageKind = "response"
	// MessageDelta delta message kind
	MessageDelta MessageKind = "delta"
	// MessageEvent event message = "event"
	MessageEvent MessageKind = "event"
	// MessageNodeProps node props message kind
	MessageNodeProps MessageKind = "nodeProps"
	// MessageDevices devices message kind
	MessageDevices MessageKind = "devices"
	// MessageDeviceEvent device event message kind
	MessageDeviceEvent MessageKind = "deviceEvent"
	// MessageReport device report message kind
	MessageDeviceReport MessageKind = "deviceReport"
	// MessageDesire device desire message kind
	MessageDeviceDesire MessageKind = "deviceDesire"
	// MessageDesire device delta message kind
	MessageDeviceDelta MessageKind = "deviceDelta"

	// MessageCommandConnect start remote debug command
	MessageCommandConnect = "connect"
	// MessageCommandDisconnect stop remote debug command
	MessageCommandDisconnect = "disconnect"
	// MessageCommandNodeLabel label the edge cluster nodes
	MessageCommandNodeLabel = "nodeLabel"
)

// Message general structure for http and ws sync
type Message struct {
	Kind     MessageKind       `yaml:"kind" json:"kind"`
	Metadata map[string]string `yaml:"meta" json:"meta"`
	Content  LazyValue         `yaml:"content" json:"content"`
}
