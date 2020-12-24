package dmcontext

import (
	"encoding/json"
	"regexp"

	"github.com/256dpi/gomqtt/packet"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
)

const (
	DeviceTopicRe   = "\\$baetyl/device/(.+)/(.+)"
	KindDelta       = "delta"
	kindEvent       = "event"
	KindGetResponse = "getResponse"
)

type observer struct {
	log  *log.Logger
	msgs map[string]chan *v1.Message
}

func NewObserver(msgs map[string]chan *v1.Message, log *log.Logger) mqtt.Observer {
	return &observer{msgs: msgs, log: log}
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
		o.log.Error("parse topic failed", log.Any("topic", pkt.Message.Topic))
		return nil
	}
	var msg *v1.Message
	switch kind {
	case KindGetResponse:
		var shad DeviceShadow
		if err := json.Unmarshal(pkt.Message.Payload, &shad); err != nil {
			return err
		}
		msg = &v1.Message{
			Kind:     v1.MessageResponse,
			Metadata: map[string]string{KeyDevice: device},
			Content:  v1.LazyValue{Value: &shad},
		}
	case KindDelta:
		var props DeviceProperties
		if err := json.Unmarshal(pkt.Message.Payload, &props); err != nil {
			return err
		}
		msg = &v1.Message{
			Kind:     v1.MessageDelta,
			Metadata: map[string]string{KeyDevice: device},
			Content:  v1.LazyValue{Value: &props},
		}
	case kindEvent:
		var event DeviceEvent
		if err := json.Unmarshal(pkt.Message.Payload, &event); err != nil {
			return err
		}
		msg = &v1.Message{
			Kind:     v1.MessageEvent,
			Metadata: map[string]string{KeyDevice: device},
			Content:  v1.LazyValue{Value: &event},
		}
	default:
		o.log.Error("get message from unexpected topic")
	}
	select {
	case o.msgs[device] <- msg:
	default:
		o.log.Error("failed to write delta message")
	}
	return nil
}

func (o *observer) OnPuback(pkt *packet.Puback) error {
	return nil
}

func (o *observer) OnError(err error) {
}
