package v1

// MessageKind message kind
type MessageKind string

const (
    // MessageReport report message kind
    MessageReport MessageKind = "report"
    // MessageDesire desire message kind
    MessageDesire MessageKind = "desire"
)

// Message general structure for http and ws sync
type Message struct {
    Kind     MessageKind       `yaml:"kind" json:"kind"`
    Metadata map[string]string `yaml:"meta" json:"meta"`
    Content  interface{}       `yaml:"content" json:"content"`
}
