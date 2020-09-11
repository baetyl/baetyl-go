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
	channel chan interface{}
	timeout time.Duration
	tomb    utils.Tomb
	log     *log.Logger
}

func NewPubsub(topic string, size int, timeout time.Duration) *Pubsub {
	return &Pubsub{
		topic:   topic,
		channel: make(chan interface{}, size),
		timeout: timeout,
		log:     log.With(log.Any("memorymq", "pubsub")),
	}
}

func (p *Pubsub) Publish(msg interface{}) error {
	select {
	case p.channel <- msg:
		return nil
	case <-p.tomb.Dying():
		return ErrPubsubClosed
	}
}

func (p *Pubsub) Subscribe(handler mq.MQHandler) {
	p.tomb.Go(func() error {
		timer := time.NewTimer(p.timeout)
		defer timer.Stop()
		for {
			select {
			case msg := <-p.channel:
				if handler != nil {
					if err := handler.OnMessage(msg); err != nil {
						p.log.Error("failed to handle msg", log.Any("topic", p.topic), log.Error(err))
					}
				}
				timer.Reset(p.timeout)
			case <-timer.C:
				p.log.Warn("message queue timeout", log.Any("topic", p.topic))
				if handler != nil {
					if err := handler.OnTimeout(); err != nil {
						p.log.Error("failed to execute timeout handler", log.Any("topic", p.topic), log.Error(err))
					}
				}
				p.tomb.Kill(nil)
			case <-p.tomb.Dying():
				return nil
			}
		}
	})
}

func (p *Pubsub) Close() {
	p.tomb.Kill(nil)
	p.tomb.Wait()
}
