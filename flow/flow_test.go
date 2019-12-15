package flow

import (
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/mqtt"
	"github.com/stretchr/testify/assert"
)

func TestFlow(t *testing.T) {
	connect := mqtt.NewConnect()
	connack := mqtt.NewConnack()

	subscribe := mqtt.NewSubscribe()
	subscribe.Subscriptions = []mqtt.Subscription{
		{Topic: "test"},
	}
	subscribe.ID = 1

	publish1 := mqtt.NewPublish()
	publish1.ID = 2
	publish1.Message.Topic = "test"
	publish1.Message.QOS = 1

	publish2 := mqtt.NewPublish()
	publish2.ID = 3
	publish2.Message.Topic = "test"
	publish2.Message.QOS = 1

	wait := make(chan struct{})

	server := New().
		Receive(connect).
		Send(connack).
		Run(func() {
			close(wait)
		}).
		Skip(&mqtt.Subscribe{}).
		Receive(publish1, publish2).
		Close()

	client := New().
		Send(connect).
		Receive(connack).
		Run(func() {
			<-wait
		}).
		Send(subscribe).
		Send(publish2, publish1).
		End()

	pipe := NewPipe()

	errCh := server.TestAsync(pipe, 100*time.Millisecond)

	err := client.Test(pipe)
	assert.NoError(t, err)

	err = <-errCh
	assert.NoError(t, err)
}

func TestAlreadyClosedError(t *testing.T) {
	pipe := NewPipe()
	pipe.Close()

	err := pipe.Send(nil)
	assert.Error(t, err)
}
