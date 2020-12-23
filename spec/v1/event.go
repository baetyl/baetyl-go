package v1

const (
	EventTypeDevice  = "eventTypeDevices"
	EventTypeDelta   = "eventTypeDelta"
	EventTopicDelta  = "eventTopicDelta"
	EventTopicDevice = "eventTopicDevice"
)

type Event struct {
	Type     string            `json:"type,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Payload  interface{}       `json:"payload,omitempty"`
}
