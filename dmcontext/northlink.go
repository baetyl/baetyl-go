package dmcontext

import v1 "github.com/baetyl/baetyl-go/v2/spec/v1"

const (
	Blink = "blink"
)

type Msg interface {
	GenPropertyReportData(properties map[string]any) v1.LazyValue
	GenEventReportData(properties map[string]any) v1.LazyValue
	GenLifecycleReportData(online bool) v1.LazyValue
	GenDeltaBlinkData(properties map[string]interface{}) v1.LazyValue
	GenPropertyGetBlinkData(properties []string) v1.LazyValue
}

func InitMsg(msg string) Msg {
	switch msg {
	case Blink:
		return &MsgBlink{}
	}
	return &MsgBlink{}
}
