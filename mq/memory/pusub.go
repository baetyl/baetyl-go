package memory

import (
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mq"
	"github.com/baetyl/baetyl-go/v2/utils"
)

var (
	ErrPubsubClosed = errors.New("failed to publish message")
)

type Pubsub struct {
	topic   string
	size    int
	channel chan interface{}
	handler mq.MQHandler
	timeout time.Duration
	timer   *time.Timer
	tomb    utils.Tomb
	log     *log.Logger
}

func NewPubsub(topic string, size int, timeout time.Duration) *Pubsub {
	return &Pubsub{
		topic:   topic,
		size:    size,
		channel: make(chan interface{}, size),
		timeout: timeout,
		timer:   time.NewTimer(timeout),
		log:     log.With(log.Any("memorymq", "pubsub")),
	}
}

func (p *Pubsub) Publish(msg interface{}) error {
	select {
	case p.channel <- msg:
	case <-p.tomb.Dying():
		return ErrPubsubClosed
	}
	p.timer.Reset(p.timeout)
	return nil
}

func (p *Pubsub) Subscribe(handler mq.MQHandler) {
	p.handler = handler
	p.tomb.Go(p.receiving)
}

func (p *Pubsub) Close() {
	p.tomb.Kill(nil)
	p.tomb.Wait()
}

func (p *Pubsub) receiving() error {
	for {
		select {
		case msg := <-p.channel:
			if p.handler != nil {
				if err := p.handler.Handler(msg); err != nil {
					p.log.Error("failed to handle msg", log.Any("topic", p.topic), log.Error(err))
				}
			}
		case <-p.timer.C:
			p.log.Warn("message queue timeout", log.Any("topic", p.topic))
			if p.handler != nil {
				if err := p.handler.Timeout(); err != nil {
					p.log.Error("failed to execute timeout handler", log.Any("topic", p.topic), log.Error(err))
				}
			}
			return nil
		case <-p.tomb.Dying():
			return nil
		}
	}
}
