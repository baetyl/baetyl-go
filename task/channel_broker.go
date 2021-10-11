package task

import (
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
)

var (
	SendMsgTimeout = errors.New("failed to send message")
	GetMsgTimeout  = errors.New("failed to get message")
)

type channelBroker struct {
	broker chan *BrokerMessage

}

func NewChannelBroker(cache int) TaskBroker {
	return &channelBroker{
		broker: make(chan *BrokerMessage, cache),
	}
}

func (b *channelBroker) SendMessage(msg *BrokerMessage) error {
	select {
	case b.broker <- msg:
		return nil
	case <- time.After(time.Millisecond):
		return SendMsgTimeout
	}
}

func (b *channelBroker) GetMessage() (*BrokerMessage, error) {
	select {
	case msg := <- b.broker:
		return msg, nil
	case <- time.After(time.Millisecond):
		return nil, GetMsgTimeout
	}
}

func (b *channelBroker) Close() error {
	close(b.broker)
	return nil
}
