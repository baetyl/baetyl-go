package mq

import (
	"io"
)

type Handler func(interface{}) error
type TimeoutHandler func() error

type MessageQueue interface {
	AddHandler(string, Handler, TimeoutHandler)
	Subscribe(string)
	Publish(string, interface{}) error
	Unsubscribe(string)
	io.Closer
}
