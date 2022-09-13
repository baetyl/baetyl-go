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
	MethodLifecyclePost  = "thing.lifecycle.post"
	DefaultVersion       = "1.0"
	KeyOnlineState       = "online_state"
)

type BlinkContent struct {
	Blink BlinkData `yaml:"blink,omitempty" json:"blink,omitempty"`
}

type BlinkData struct {
	ReqId      string                 `yaml:"reqId,omitempty" json:"reqId,omitempty"`
	Method     string                 `yaml:"method,omitempty" json:"method,omitempty"`
	Version    string                 `yaml:"version,omitempty" json:"version,omitempty"`
	Timestamp  int64                  `yaml:"timestamp,omitempty" json:"timestamp,omitempty"`
	Properties interface{}            `yaml:"properties,omitempty" json:"properties,omitempty"`
	Events     map[string]interface{} `yaml:"events,omitempty" json:"events,omitempty"`
	Params     map[string]interface{} `yaml:"params,omitempty" json:"params,omitempty"`
}

func GenDeltaBlinkData(properties map[string]interface{}) BlinkData {
	return BlinkData{
		ReqId:      uuid.New().String(),
		Method:     MethodPropertyInvoke,
		Version:    DefaultVersion,
		Timestamp:  getCurrentTimestamp(),
		Properties: properties,
	}
}

func GenPropertyReportBlinkData(properties map[string]interface{}) BlinkData {
	return BlinkData{
		ReqId:      uuid.New().String(),
		Method:     MethodPropertyReport,
		Version:    DefaultVersion,
		Timestamp:  getCurrentTimestamp(),
		Properties: properties,
	}
}

func GenEventReportBlinkData(events map[string]interface{}) BlinkData {
	return BlinkData{
		ReqId:     uuid.New().String(),
		Method:    MethodEventReport,
		Version:   DefaultVersion,
		Timestamp: getCurrentTimestamp(),
		Events:    events,
	}
}

func GenPropertyGetBlinkData(properties []string) BlinkData {
	return BlinkData{
		ReqId:      uuid.New().String(),
		Method:     MethodPropertyGet,
		Version:    DefaultVersion,
		Timestamp:  getCurrentTimestamp(),
		Properties: properties,
	}
}

func GenLifecycleReportBlinkData(online bool) BlinkData {
	return BlinkData{
		ReqId:     uuid.New().String(),
		Method:    MethodLifecyclePost,
		Version:   DefaultVersion,
		Timestamp: getCurrentTimestamp(),
		Params:    map[string]interface{}{KeyOnlineState: online},
	}
}

func getCurrentTimestamp() int64 {
	return time.Now().UnixNano() / 1e6
}
