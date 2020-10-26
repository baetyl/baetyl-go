package pubsub

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	expMsg = "exp"
)

var (
	wg = sync.WaitGroup{}
)

func TestPubsub(t *testing.T) {
	pb, err := NewPubsub(1)
	assert.NoError(t, err)
	assert.NotNil(t, pb)

	topicTest1 := "1"
	topicTest2 := "2"

	ch1, err := pb.Subscribe(topicTest1)
	assert.NoError(t, err)
	go Reading(t, ch1)
	wg.Add(1)
	pb.Publish(topicTest1, expMsg)

	ch2, err := pb.Subscribe(topicTest1)
	assert.NoError(t, err)
	go Reading(t, ch2)
	wg.Add(2)
	pb.Publish(topicTest1, expMsg)

	ch3, err := pb.Subscribe(topicTest2)
	assert.NoError(t, err)
	go Reading(t, ch3)
	wg.Add(1)
	pb.Publish(topicTest2, expMsg)

	wg.Wait()

	err = pb.Close()
	assert.NoError(t, err)
}

func Reading(t *testing.T, ch <-chan interface{}) {
	for {
		msg := <-ch
		switch msg.(type) {
		case string:
			assert.Equal(t, expMsg, msg)
			wg.Done()
		default:
			return
		}
	}
}
