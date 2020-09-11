package mq

import (
	"io"
)

type MQHandler interface {
	Handler(interface{}) error
	Timeout() error
}

type MessageQueue interface {
	Subscribe(string, MQHandler)
	Publish(string, interface{}) error
	Unsubscribe(string)
	io.Closer
}
