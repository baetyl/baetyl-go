package memory

import (
	"sync"
	"time"

	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mq"
)

type defaultMQ struct {
	size    int
	pubsubs sync.Map
	timeout time.Duration
	log     *log.Logger
}

func NewMQ(size int, timeout time.Duration) (mq.MessageQueue, error) {
	return &defaultMQ{
		size:    size,
		pubsubs: sync.Map{},
		timeout: timeout,
		log:     log.With(log.Any("mq", "memorymq")),
	}, nil
}

func (q *defaultMQ) Subscribe(topic string, handler mq.MQHandler) {
	q.loadOrCreatePubsub(topic).Subscribe(handler)
}

func (q *defaultMQ) Unsubscribe(topic string) {
	ps, ok := q.pubsubs.Load(topic)
	if ok {
		q.pubsubs.Delete(topic)
		ps.(*Pubsub).Close()
	}
}

func (q *defaultMQ) Publish(topic string, msg interface{}) error {
	return q.loadOrCreatePubsub(topic).Publish(msg)
}

func (q *defaultMQ) Close() error {
	q.pubsubs.Range(func(key, value interface{}) bool {
		q.Unsubscribe(key.(string))
		return true
	})
	return nil
}

func (q *defaultMQ) loadOrCreatePubsub(topic string) *Pubsub {
	ps, ok := q.pubsubs.Load(topic)
	if ok {
		return ps.(*Pubsub)
	}

	pubsub := NewPubsub(topic, q.size, q.timeout)
	act, loaded := q.pubsubs.LoadOrStore(topic, pubsub)
	if loaded {
		pubsub.Close()
	}
	return act.(*Pubsub)
}
