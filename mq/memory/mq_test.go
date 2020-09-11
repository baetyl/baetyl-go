package memory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type handler struct {
	h  func(interface{}) error
	th func() error
}

func (h *handler) Handler(msg interface{}) error {
	return h.h(msg)
}

func (h *handler) Timeout() error {
	return h.th()
}

func TestNewMQ(t *testing.T) {
	mq, err := NewMQ(1, time.Second*100)
	assert.NoError(t, err)
	assert.NotNil(t, mq)

	topic := "test"
	msgSend := "send"

	handler := &handler{
		h: func(msg interface{}) error {
			assert.Equal(t, msgSend, msg.(string))
			return nil
		},
		th: func() error {
			return nil
		},
	}

	mq.Subscribe(topic, handler)

	err = mq.Publish(topic, msgSend)
	assert.NoError(t, err)

	mq.Unsubscribe(topic)

	err = mq.Close()
	assert.NoError(t, err)
}

func TestTimeout(t *testing.T) {
	mq, err := NewMQ(1, time.Millisecond)
	assert.NoError(t, err)
	assert.NotNil(t, mq)

	topic := "test"
	msgSend := "send"

	msg := "test"
	msgCh := make(chan string, 1)

	handler := &handler{
		h: func(msg interface{}) error {
			assert.Equal(t, msgSend, msg.(string))
			return nil
		},
		th: func() error {
			msgCh <- msg
			return nil
		},
	}

	mq.Subscribe(topic, handler)

	time.Sleep(time.Millisecond * 2)

	err = mq.Publish(topic, msgSend)
	assert.NoError(t, err)

	res := <-msgCh
	assert.Equal(t, msg, res)

	err = mq.Close()
	assert.NoError(t, err)
}
