package memory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMQ(t *testing.T) {
	mq, err := NewMQ(1, time.Second*100)
	assert.NoError(t, err)
	assert.NotNil(t, mq)

	topic := "test"
	msgSend := "send"

	mq.AddHandler(topic, func(msg interface{}) error {
		assert.Equal(t, msgSend, msg.(string))
		return nil
	}, func() error {
		return nil
	})

	mq.Subscribe(topic)

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

	mq.AddHandler(topic, func(msg interface{}) error {
		assert.Equal(t, msgSend, msg.(string))
		return nil
	}, func() error {
		msgCh <- msg
		return nil
	})

	mq.Subscribe(topic)

	time.Sleep(time.Millisecond * 2)

	err = mq.Publish(topic, msgSend)
	assert.NoError(t, err)

	res := <-msgCh
	assert.Equal(t, msg, res)

	err = mq.Close()
	assert.NoError(t, err)
}
