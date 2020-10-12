package pubsub

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	topicDown = "test.down"
	topicUp   = "test.up"
)

var (
	syncWG = sync.WaitGroup{}
)

func TestNewPubsubHelper(t *testing.T) {
	pb, err := NewPubsub(1)
	assert.NoError(t, err)
	assert.NotNil(t, pb)

	chDown := pb.Subscribe(topicDown)
	hpDown := NewPubsubHelper(chDown, time.Second*2, &hdDown{pb: pb, t: t})
	hpDown.Start()

	chUp := pb.Subscribe(topicUp)
	hphUp := NewPubsubHelper(chUp, time.Second*2, &hdUp{pb: pb, t: t})
	hphUp.Start()

	pb.Publish(topicDown, "down")
	syncWG.Add(1)
	syncWG.Wait()
	hpDown.Close()
	hphUp.Close()
}

type hdDown struct {
	pb Pubsub
	t  *testing.T
}

func (h *hdDown) OnMessage(msg interface{}) error {
	m, ok := msg.(string)
	assert.True(h.t, ok)
	assert.Equal(h.t, "down", m)
	h.pb.Publish(topicUp, "up")
	return nil
}

func (h *hdDown) OnTimeout() error {
	h.pb.Publish(topicUp, "timeout")
	return nil
}

type hdUp struct {
	pb Pubsub
	t  *testing.T
}

func (h *hdUp) OnMessage(msg interface{}) error {
	m, ok := msg.(string)
	assert.True(h.t, ok)
	assert.Equal(h.t, "up", m)
	syncWG.Done()
	return nil
}

func (h *hdUp) OnTimeout() error {
	return nil
}
