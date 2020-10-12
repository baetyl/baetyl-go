package pubsub

import (
	"time"

	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/utils"
)

type Handler interface {
	OnMessage(interface{}) error
	OnTimeout() error
}

type PubsubHelper interface {
	Start()
	Close()
}

type helper struct {
	channel chan interface{}
	timeout time.Duration
	handler Handler
	tomb    utils.Tomb
	log     *log.Logger
}

func NewPubsubHelper(ch chan interface{}, timeout time.Duration, handler Handler) PubsubHelper {
	return &helper{
		channel: ch,
		timeout: timeout,
		handler: handler,
		tomb:    utils.Tomb{},
		log:     log.L().With(log.Any("pubsub", "helper")),
	}
}

func (h *helper) Start() {
	h.tomb.Go(h.processing)
}

func (h *helper) Close() {
	h.tomb.Kill(nil)
	h.tomb.Wait()
}

func (h *helper) processing() error {
	timer := time.NewTimer(h.timeout)
	defer timer.Stop()
	for {
		select {
		case msg := <-h.channel:
			if h.handler != nil {
				if err := h.handler.OnMessage(msg); err != nil {
					h.log.Error("failed to handle msg", log.Error(err))
				}
			}
			timer.Reset(h.timeout)
		case <-timer.C:
			h.log.Warn("pubsub queue timeout")
			if h.handler != nil {
				if err := h.handler.OnTimeout(); err != nil {
					h.log.Error("failed to execute timeout helper", log.Error(err))
				}
			}
			h.tomb.Kill(nil)
		case <-h.tomb.Dying():
			return nil
		}
	}
}
