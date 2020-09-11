package mq

import (
	"io"
)

type MQHandler interface {
	OnMessage(interface{}) error
	OnTimeout() error
}

type MessageQueue interface {
	Subscribe(string, MQHandler)
	Publish(string, interface{}) error
	Unsubscribe(string)
	io.Closer
}
