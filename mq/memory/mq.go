package memory

import (
	"sync"
	"time"

	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mq"
)

type defaultMQ struct {
	size      int
	handlers  sync.Map
	tHandlers sync.Map
	pubsubs   sync.Map
	timeout   time.Duration
	log       *log.Logger
}

func NewMQ(size int, timeout time.Duration) (mq.MessageQueue, error) {
	return &defaultMQ{
		size:      size,
		handlers:  sync.Map{},
		tHandlers: sync.Map{},
		pubsubs:   sync.Map{},
		timeout:   timeout,
		log:       log.With(log.Any("mq", "memorymq")),
	}, nil
}

func (q *defaultMQ) AddHandler(topic string, handler mq.Handler, timeoutHandler mq.TimeoutHandler) {
	q.handlers.Store(topic, handler)
	q.tHandlers.Store(topic, timeoutHandler)
}

func (q *defaultMQ) Subscribe(topic string) {
	q.loadOrCreatePubsub(topic).Subscribe()
}

func (q *defaultMQ) Unsubscribe(topic string) {
	q.handlers.Delete(topic)
	ps, ok := q.pubsubs.Load(topic)
	if ok {
		ps.(*Pubsub).Close()
	}
	q.pubsubs.Delete(topic)
}

func (q *defaultMQ) Publish(topic string, msg interface{}) error {
	return q.loadOrCreatePubsub(topic).Public(msg)
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

	var handler mq.Handler
	h, ok := q.handlers.Load(topic)
	if !ok {
		q.log.Warn("no handler is added to the subscribed topic", log.Any("topic", topic))
		handler = nil
	} else {
		handler = h.(mq.Handler)
	}

	var timeoutHandler mq.TimeoutHandler
	th, ok := q.tHandlers.Load(topic)
	if !ok {
		q.log.Warn("no timeout handler is added to the subscribed topic", log.Any("topic", topic))
		timeoutHandler = nil
	} else {
		timeoutHandler = th.(mq.TimeoutHandler)
	}

	pubsub := NewPubsub(topic, q.size, q.timeout, handler, timeoutHandler)
	act, loaded := q.pubsubs.LoadOrStore(topic, pubsub)
	if loaded {
		pubsub.Close()
	}
	return act.(*Pubsub)
}
