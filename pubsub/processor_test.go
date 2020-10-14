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

func TestNewTimerProcessor(t *testing.T) {
	pb, err := NewPubsub(1)
	assert.NoError(t, err)
	assert.NotNil(t, pb)

	chDown, err := pb.Subscribe(topicDown)
	assert.NoError(t, err)
	hpDown := NewProcessor(chDown, time.Second*2, &hdDown{pb: pb, t: t})
	hpDown.Start()

	chUp, err := pb.Subscribe(topicUp)
	assert.NoError(t, err)
	hphUp := NewProcessor(chUp, time.Second*2, &hdUp{pb: pb, t: t})
	hphUp.Start()

	err = pb.Publish(topicDown, "down")
	assert.NoError(t, err)
	syncWG.Add(1)
	syncWG.Wait()
	hpDown.Close()
	hphUp.Close()
}

func TestNewProcessor(t *testing.T) {
	pb, err := NewPubsub(1)
	assert.NoError(t, err)
	assert.NotNil(t, pb)

	chDown, err := pb.Subscribe(topicDown)
	assert.NoError(t, err)
	hpDown := NewProcessor(chDown, 0, &hdDown{pb: pb, t: t})
	hpDown.Start()

	chUp, err := pb.Subscribe(topicUp)
	assert.NoError(t, err)
	hphUp := NewProcessor(chUp, 0, &hdUp{pb: pb, t: t})
	hphUp.Start()

	err = pb.Publish(topicDown, "down")
	assert.NoError(t, err)
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
	return h.pb.Publish(topicUp, "up")
}

func (h *hdDown) OnTimeout() error {
	return h.pb.Publish(topicUp, "timeout")
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
