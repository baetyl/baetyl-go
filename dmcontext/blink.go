package dmcontext

import (
	"time"

	"github.com/google/uuid"
)

const (
	MethodPropertyInvoke = "thing.property.invoke"
	MethodPropertyReport = "thing.property.post"
	MethodEventReport    = "thing.event.post"
	MethodPropertyGet    = "thing.property.get"
	DefaultVersion       = "1.0"
)

type BlinkContent struct {
	Blink BlinkData `yaml:"blink,omitempty" json:"blink,omitempty"`
}

type BlinkData struct {
	ReqId      string                 `yaml:"reqId,omitempty" json:"reqId,omitempty"`
	Method     string                 `yaml:"method,omitempty" json:"method,omitempty"`
	Version    string                 `yaml:"version,omitempty" json:"version,omitempty"`
	Timestamp  int64                  `yaml:"timestamp,omitempty" json:"timestamp,omitempty"`
	Properties map[string]interface{} `yaml:"properties,omitempty" json:"properties,omitempty"`
	Events     map[string]interface{} `yaml:"events,omitempty" json:"events,omitempty"`
}

func GenDeltaBlinkData(properties map[string]interface{}) BlinkData {
	return BlinkData{
		ReqId:      uuid.New().String(),
		Method:     MethodPropertyInvoke,
		Version:    DefaultVersion,
		Timestamp:  time.Now().Unix() / 1e6,
		Properties: properties,
	}
}

func GenPropertyReportBlinkData(properties map[string]interface{}) BlinkData {
	return BlinkData{
		ReqId:      uuid.New().String(),
		Method:     MethodPropertyReport,
		Version:    DefaultVersion,
		Timestamp:  time.Now().Unix() / 1e6,
		Properties: properties,
	}
}

func GenEventReportBlinkData(events map[string]interface{}) BlinkData {
	return BlinkData{
		ReqId:      uuid.New().String(),
		Method:     MethodEventReport,
		Version:    DefaultVersion,
		Timestamp:  time.Now().Unix() / 1e6,
		Properties: events,
	}
}
