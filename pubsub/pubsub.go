package pubsub

import (
	"io"
	"sync"
	"time"

	"github.com/baetyl/baetyl-go/v2/log"
)

const (
	pubTimeout = time.Millisecond * 10
)

type Pubsub interface {
	Publish(topic string, msg interface{})
	Subscribe(topic string) chan interface{}
	Unsubscribe(topic string, ch chan interface{})
	io.Closer
}

type pubsub struct {
	size     int
	channels map[string]map[chan interface{}]struct{}
	chanLock sync.Mutex
	log      *log.Logger
}

func NewPubsub(size int) (Pubsub, error) {
	return &pubsub{
		size:     size,
		channels: make(map[string]map[chan interface{}]struct{}),
		log:      log.With(log.Any("pubsub", "memory")),
	}, nil
}

func (m *pubsub) Publish(topic string, msg interface{}) {
	m.chanLock.Lock()
	chs, ok := m.channels[topic]
	if !ok {
		chs = map[chan interface{}]struct{}{}
		m.channels[topic] = chs
	}
	m.chanLock.Unlock()

	for ch, _ := range chs {
		m.publish(ch, msg)
	}
}

func (m *pubsub) Subscribe(topic string) chan interface{} {
	m.chanLock.Lock()
	defer m.chanLock.Unlock()

	chs, ok := m.channels[topic]
	if !ok {
		chs = map[chan interface{}]struct{}{}
		m.channels[topic] = chs
	}
	ch := make(chan interface{}, m.size)
	chs[ch] = struct{}{}
	return ch
}

func (m *pubsub) Unsubscribe(topic string, ch chan interface{}) {
	m.chanLock.Lock()
	defer m.chanLock.Unlock()
	if chs, ok := m.channels[topic]; ok {
		if _, exist := chs[ch]; exist {
			delete(chs, ch)
		}
	}
}

func (m *pubsub) Close() error {
	m.chanLock.Lock()
	defer m.chanLock.Unlock()
	for topic, chs := range m.channels {
		for k, _ := range chs {
			delete(chs, k)
			close(k)
		}
		delete(m.channels, topic)
	}
	return nil
}

func (m *pubsub) publish(ch chan interface{}, msg interface{}) {
	timer := time.NewTimer(pubTimeout)
	select {
	case ch <- msg:
		if !timer.Stop() {
			<-timer.C
		}
	case <-timer.C:
		m.log.Warn("publish message timeout")
	}
}
