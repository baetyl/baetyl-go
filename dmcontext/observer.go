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
)

type observer struct {
	log  *log.Logger
	msgs map[string]chan *v1.Message
}

func newObserver(msgs map[string]chan *v1.Message, log *log.Logger) mqtt.Observer {
	return &observer{msgs: msgs, log: log}
}

func ParseTopic(topic string) (string, error) {
	r, err := regexp.Compile(DeviceTopicRe)
	if err != nil {
		return "", err
	}
	res := r.FindStringSubmatch(topic)
	if len(res) != 3 {
		return "", errors.New("illegal topic can not parse")
	}
	return res[1], nil
}

func (o *observer) OnPublish(pkt *packet.Publish) error {
	device, err := ParseTopic(pkt.Message.Topic)
	if err != nil {
		o.log.Error("parse topic failed", log.Any("topic", pkt.Message.Topic))
		return nil
	}
	var msg v1.Message
	if err := json.Unmarshal(pkt.Message.Payload, &msg); err != nil {
		o.log.Error("failed to unmarshal message",
			log.Any("payload", string(pkt.Message.Payload)))
		return nil
	}
	if ch, ok := o.msgs[device]; ok {
		select {
		case ch <- &msg:
		default:
			o.log.Error("failed to write device message", log.Any("msg", msg))
		}
	} else {
		o.log.Error("device channel not exist",
			log.Any("device name", device))
	}
	return nil
}

func (o *observer) OnPuback(pkt *packet.Puback) error {
	return nil
}

func (o *observer) OnError(err error) {
}
