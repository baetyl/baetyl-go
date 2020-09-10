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
)

// Message general structure for http and ws sync
type Message struct {
	Kind     MessageKind       `yaml:"kind" json:"kind"`
	Metadata map[string]string `yaml:"meta" json:"meta"`
	Content  VariableValue     `yaml:"content" json:"content"`
}
