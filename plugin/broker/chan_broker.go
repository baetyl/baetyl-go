package broker

import (
	"sync"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/models"
	"github.com/baetyl/baetyl-go/v2/plugin"
)

var (
	ErrQueueExist = errors.New("failed to register queue, queue exists")
	ErrQueueNotExist = errors.New("failed to get queue, queue not exists")
)

func init() {
	plugin.RegisterFactory("defaultqueue", NewBroker)
}

type ChanBroker struct {
	QueueMap map[string]chan *models.Task
	BrokerLock sync.RWMutex
}

func NewBroker() (plugin.Plugin, error) {
	return &ChanBroker{QueueMap: map[string]chan *models.Task{}}, nil
}

// RegisterQueue Register a new queue with specific name.
func (c *ChanBroker) RegisterQueue(queueName string, limit int) error {
	c.BrokerLock.Lock()
	defer c.BrokerLock.Unlock()

	if _, ok := c.QueueMap[queueName]; ok {
		return ErrQueueExist
	}

	c.QueueMap[queueName] = make(chan *models.Task, limit)
	return nil
}

func (c *ChanBroker) Dequeue(queueName string) (*models.Task, error) {
	c.BrokerLock.RLock()
	defer c.BrokerLock.RUnlock()

	value, ok := c.QueueMap[queueName]
	if !ok {
		return nil, ErrQueueNotExist
	}

	return <- value, nil
}

func (c *ChanBroker) Enqueue(queueName string, task *models.Task) error {
	c.BrokerLock.RLock()
	defer c.BrokerLock.RUnlock()

	value, ok := c.QueueMap[queueName]
	if !ok {
		return ErrQueueNotExist
	}

	value <- task
	return nil
}

// Close close all channels.
func (c *ChanBroker) Close() error {
	c.BrokerLock.Lock()
	defer c.BrokerLock.Unlock()

	for _, v := range c.QueueMap {
		close(v)
	}
	return nil
}
