package dmcontext

import (
	"encoding/json"
	"regexp"

	"github.com/256dpi/gomqtt/packet"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mqtt"
)

const (
	DeviceTopicRe   = "\\$baetyl/device/(.+)/(.+)"
	KindDelta       = "delta"
	kindEvent       = "event"
	KindGetResponse = "getResponse"
)

type observer struct {
	Ctx DmCtx
}

func NewObserver() mqtt.Observer {
	return &observer{}
}

func parseTopic(topic string) (string, string, error) {
	r, err := regexp.Compile(DeviceTopicRe)
	if err != nil {
		return "", "", err
	}
	res := r.FindStringSubmatch(topic)
	if len(res) != 3 {
		return "", "", errors.New("illegal topic can not parse")
	}
	return res[1], res[2], nil
}

func (o *observer) OnPublish(pkt *packet.Publish) error {
	device, kind, err := parseTopic(pkt.Message.Topic)
	if err != nil {
		o.Ctx.log.Error("parse topic failed", log.Any("topic", pkt.Message.Topic))
		return nil
	}
	var msg DeviceMessage
	switch kind {
	case KindGetResponse:
		var shad DeviceShadow
		if err := json.Unmarshal(pkt.Message.Payload, &shad); err != nil {
			return err
		}
		msg = DeviceMessage{
			Type:       ResponseMessage,
			DeviceInfo: &DeviceInfo{Name: device},
			Payload:    &shad,
		}
	case KindDelta:
		var deltaMsg DeviceProperties
		if err := json.Unmarshal(pkt.Message.Payload, &deltaMsg); err != nil {
			return err
		}
		msg = DeviceMessage{
			Type:       DeltaMessage,
			DeviceInfo: &DeviceInfo{Name: device},
			Payload:    deltaMsg,
		}
	case kindEvent:
		var eventMsg DeviceEvent
		if err := json.Unmarshal(pkt.Message.Payload, &eventMsg); err != nil {
			return err
		}
		msg = DeviceMessage{
			Type:       EventMessage,
			DeviceInfo: &DeviceInfo{Name: device},
			Payload:    eventMsg,
		}
	default:
		o.Ctx.log.Error("get message from unexpected topic")
	}
	select {
	case o.Ctx.msgs[device] <- &msg:
	default:
		o.Ctx.log.Error("failed to write delta message")
	}
	return nil
}

func (o *observer) OnPuback(pkt *packet.Puback) error {
	return nil
}

func (o *observer) OnError(err error) {
}
