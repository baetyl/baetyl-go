package task

import (
	"github.com/google/uuid"
)

type taskProducer struct {
	broker  TaskBroker
	backend TaskBackend
}

func NewTaskProducer(broker TaskBroker, backend TaskBackend) TaskProducer {
	return &taskProducer{
		broker,
		backend,
	}
}

func (p *taskProducer) AddTask(name string, args ...interface{}) (*TaskResult, error) {
	id, _ := uuid.NewUUID()
	task := &TaskMessage{
		ID:     id.String(),
		Name:   name,
		Args:   args,
		Kwargs: make(map[string]interface{}),
	}
	encodedMsg, err := task.Encode()
	if err != nil {
		return nil, err
	}
	return &TaskResult{ID: id.String(), backend: p.backend},
		p.broker.SendMessage(&BrokerMessage{
			ID:    id.String(),
			Value: encodedMsg,
		})
}

func (p *taskProducer) AddTaskWithKey(name string, args map[string]interface{}) (*TaskResult, error) {
	id, _ := uuid.NewUUID()
	task := &TaskMessage{
		ID:     id.String(),
		Name:   name,
		Args:   make([]interface{}, 0),
		Kwargs: args,
	}
	encodedMsg, err := task.Encode()
	if err != nil {
		return nil, err
	}
	return &TaskResult{ID: id.String(), backend: p.backend}, p.broker.SendMessage(&BrokerMessage{
		ID:    id.String(),
		Value: encodedMsg,
	})
}
