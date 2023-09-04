package dmcontext

import (
	"time"

	"github.com/google/uuid"

	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
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

type MsgBlink struct {
}

type ContentBlink struct {
	Blink DataBlink `yaml:"blink,omitempty" json:"blink,omitempty"`
}

type DataBlink struct {
	ReqID      string         `yaml:"reqId,omitempty" json:"reqId,omitempty"`
	Method     string         `yaml:"method,omitempty" json:"method,omitempty"`
	Version    string         `yaml:"version,omitempty" json:"version,omitempty"`
	Timestamp  int64          `yaml:"timestamp,omitempty" json:"timestamp,omitempty"`
	Properties any            `yaml:"properties,omitempty" json:"properties,omitempty"`
	Events     map[string]any `yaml:"events,omitempty" json:"events,omitempty"`
	Params     map[string]any `yaml:"params,omitempty" json:"params,omitempty"`
}

func (b *MsgBlink) GenDeltaBlinkData(properties map[string]interface{}) v1.LazyValue {
	return v1.LazyValue{
		Value: ContentBlink{
			Blink: DataBlink{
				ReqID:      uuid.New().String(),
				Method:     MethodPropertyInvoke,
				Version:    DefaultVersion,
				Timestamp:  getCurrentTimestamp(),
				Properties: properties,
			},
		},
	}
}

func (b *MsgBlink) GenPropertyReportData(properties map[string]any) v1.LazyValue {
	return v1.LazyValue{
		Value: ContentBlink{
			Blink: DataBlink{
				ReqID:      uuid.New().String(),
				Method:     MethodPropertyReport,
				Version:    DefaultVersion,
				Timestamp:  getCurrentTimestamp(),
				Properties: properties,
			},
		},
	}
}

func (b *MsgBlink) GenEventReportData(events map[string]any) v1.LazyValue {
	return v1.LazyValue{
		Value: ContentBlink{
			Blink: DataBlink{
				ReqID:     uuid.New().String(),
				Method:    MethodEventReport,
				Version:   DefaultVersion,
				Timestamp: getCurrentTimestamp(),
				Events:    events,
			},
		},
	}
}

func (b *MsgBlink) GenPropertyGetBlinkData(properties []string) v1.LazyValue {
	return v1.LazyValue{
		Value: ContentBlink{
			Blink: DataBlink{
				ReqID:      uuid.New().String(),
				Method:     MethodPropertyGet,
				Version:    DefaultVersion,
				Timestamp:  getCurrentTimestamp(),
				Properties: properties,
			},
		},
	}
}

func (b *MsgBlink) GenLifecycleReportData(online bool) v1.LazyValue {
	return v1.LazyValue{
		Value: ContentBlink{
			Blink: DataBlink{
				ReqID:     uuid.New().String(),
				Method:    MethodLifecyclePost,
				Version:   DefaultVersion,
				Timestamp: getCurrentTimestamp(),
				Params:    map[string]any{KeyOnlineState: online},
			},
		},
	}
}

func getCurrentTimestamp() int64 {
	return time.Now().UnixNano() / 1e6
}
