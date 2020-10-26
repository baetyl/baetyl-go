package pubsub

import (
	"io"
	"strings"
	"sync"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
)

const (
	pubTimeout = time.Millisecond * 10
)

var (
	ErrPubsubTimeout = errors.New("failed to send message to topic because of timeout")
)

type Pubsub interface {
	Publish(topic string, msg interface{}) error
	Subscribe(topic string) (<-chan interface{}, error)
	Unsubscribe(topic string, ch <-chan interface{}) error
	io.Closer
}

type pubsub struct {
	size     int
	channels map[string]map[<-chan interface{}]chan interface{}
	chanLock sync.RWMutex
	log      *log.Logger
}

func NewPubsub(size int) (Pubsub, error) {
	return &pubsub{
		size:     size,
		channels: make(map[string]map[<-chan interface{}]chan interface{}),
		log:      log.With(log.Any("pubsub", "memory")),
	}, nil
}

func (m *pubsub) Publish(topic string, msg interface{}) error {
	var errs []string
	if chs := m.getChannel(topic); chs != nil {
		for _, ch := range chs {
			err := m.publish(ch, msg)
			if err != nil {
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

func (m *pubsub) Subscribe(topic string) (<-chan interface{}, error) {
	m.chanLock.Lock()
	defer m.chanLock.Unlock()

	chs, ok := m.channels[topic]
	if !ok {
		chs = map[<-chan interface{}]chan interface{}{}
		m.channels[topic] = chs
	}
	ch := make(chan interface{}, m.size)
	chs[ch] = ch
	return ch, nil
}

func (m *pubsub) Unsubscribe(topic string, ch <-chan interface{}) error {
	m.chanLock.Lock()
	defer m.chanLock.Unlock()
	if chs, ok := m.channels[topic]; ok {
		if _, exist := chs[ch]; exist {
			delete(chs, ch)
		}
	}
	return nil
}

func (m *pubsub) Close() error {
	m.chanLock.Lock()
	defer m.chanLock.Unlock()
	for topic, chs := range m.channels {
		for k, _ := range chs {
			delete(chs, k)
		}
		delete(m.channels, topic)
	}
	return nil
}

func (m *pubsub) publish(ch chan interface{}, msg interface{}) error {
	timer := time.NewTimer(pubTimeout)
	defer timer.Stop()

	select {
	case ch <- msg:
	case <-timer.C:
		m.log.Warn("publish message timeout")
		return ErrPubsubTimeout
	}
	return nil
}

func (m *pubsub) getChannel(topic string) map[<-chan interface{}]chan interface{} {
	m.chanLock.RLock()
	defer m.chanLock.RUnlock()
	if chs, ok := m.channels[topic]; ok {
		return chs
	}
	return nil
}
