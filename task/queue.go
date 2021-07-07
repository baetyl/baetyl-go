package task

import (
	"sync"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/models"
	"github.com/baetyl/baetyl-go/v2/plugin"
)

type Queue struct {
	Name   string
	Limit  int
	queue  plugin.Broker
}

var queueMap = sync.Map{}

var (
	ErrQueueExist    = errors.New("failed to new queue, queue exists")
	ErrQueueNotExist = errors.New("failed to get queue, queue not exists")
)

// NewQueue New a FIFO queue and add it to the queue map.
// PARAMS:
//   - name: specific queue name
//   - limit: max length of this queue
// RETURNS:
//   Queue: pointer to the new queue
//   error: if has error else nil
func NewQueue(name string, limit int) (*Queue, error) {
	if _, ok := queueMap.Load(name); ok {
		return nil, ErrQueueExist
	}
	q, err := plugin.GetPlugin("defaultqueue")
	if err != nil {
		return nil, err
	}
	err = q.(plugin.Broker).RegisterQueue(name, limit)
	if err != nil {
		return nil, err
	}
	queue := &Queue{Name: name, Limit: limit, queue: q.(plugin.Broker)}
	queueMap.Store(name, queue)
	return queue, nil
}

// GetQueue Get queue from the queue map.
// PARAMS:
//   - name: specific queue name
// RETURNS:
//   Queue: pointer to the new queue
//   error: if has error else nil
func GetQueue(name string) (*Queue, error) {
	if v, ok := queueMap.Load(name); ok {
		return v.(*Queue), nil
	}
	return nil, ErrQueueNotExist
}

// Enqueue Add task to the queue.
// PARAMS:
//   - task: task to be added.
// RETURNS:
//   error: if has error else nil
func (q *Queue) Enqueue(task *models.Task) error {
	return q.queue.Enqueue(q.Name, task)
}

// Enqueue Get one task from the queue.
// RETURNS:
//   task: pointer to the task.
//   error: if has error else nil
func (q *Queue) Dequeue() (*models.Task, error) {
	return q.queue.Dequeue(q.Name)
}
