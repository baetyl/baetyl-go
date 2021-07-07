package plugin

import "github.com/baetyl/baetyl-go/v2/models"

type Broker interface {
	RegisterQueue (queueName string, limit int) error
	Dequeue (queueName string) (*models.Task, error)
	Enqueue (queueName string, task *models.Task) error
}