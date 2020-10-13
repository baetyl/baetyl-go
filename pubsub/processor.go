package pubsub

import (
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/utils"
)

var (
	ErrProcessorTimeout = errors.New("failed to send message because of timeout")
)

type Processor interface {
	Start()
	Close()
}

type processor struct {
	channel chan interface{}
	timeout time.Duration
	handler Handler
	tomb    utils.Tomb
	log     *log.Logger
}

func NewProcessor(ch chan interface{}, timeout time.Duration, handler Handler) Processor {
	return &processor{
		channel: ch,
		timeout: timeout,
		handler: handler,
		tomb:    utils.Tomb{},
		log:     log.L().With(log.Any("pubsub", "processor")),
	}
}

func (p *processor) Start() {
	p.tomb.Go(p.processing)
}

func (p *processor) Close() {
	p.tomb.Kill(nil)
	p.tomb.Wait()
}

func (p *processor) processing() error {
	timer := time.NewTimer(p.timeout)
	defer timer.Stop()
	for {
		select {
		case msg := <-p.channel:
			if p.handler != nil {
				if err := p.handler.OnMessage(msg); err != nil {
					p.log.Error("failed to handle message", log.Error(err))
				}
			}
			timer.Reset(p.timeout)
		case <-timer.C:
			p.log.Warn("pubsub timeout")
			if p.handler != nil {
				if err := p.handler.OnTimeout(); err != nil {
					p.log.Error("failed to handle message because of timeout", log.Error(err))
				}
			}
			p.tomb.Kill(ErrProcessorTimeout)
		case <-p.tomb.Dying():
			return nil
		}
	}
}
