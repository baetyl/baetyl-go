package task

import (
	"bytes"
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl-go/v2/models"
	_ "github.com/baetyl/baetyl-go/v2/plugin/broker"
)

const (
	queueName = "queueTest"
	job1Name  = "jobTest1"
	job2Name  = "jobTest2"
	limit     = 10
)

type testStruct struct {
	strList []string
	strMap  map[string]string
}

func TestWorker(t *testing.T)  {
	var err error
	handler1 := func(a, b int) (int, error) {
		return a + b, nil
	}
	handler2 := func(c int) (string, error) {
		result := strconv.Itoa(c)
		return result, nil
	}

	handler3 := func(d *testStruct) error {
		var buffer bytes.Buffer
		for _, str := range d.strList {
			buffer.WriteString(str)
		}
		for v, str := range d.strMap {
			buffer.WriteString(v)
			buffer.WriteString(str)
		}
		return nil
	}

	handler4 := func() error {
		return nil
	}

	task1 := &models.Task{
		Name: "task",
		Namespace: "task",
		JobName: job1Name,
		Args: []models.Arg{{Type: "int", Value: 1}, {Type: "int", Value: 2}},
		Async: true,
	}

	argStruct := &testStruct{
		strList: []string{"test1"},
		strMap: map[string]string{},
	}
	argStruct.strMap["key"] = "value"
	task2 := &models.Task{
		Name: "task2",
		Namespace: "task",
		JobName: job2Name,
		Args: []models.Arg{{Type: "string", Value: argStruct}},
		Async: true,
	}

	err = JobRegister.Register(job1Name, handler1, handler2)
	assert.NoError(t, err)

	err = JobRegister.Register(job2Name, handler3, handler4)
	assert.NoError(t, err)

	task1, err = JobRegister.NewTask(task1)
	assert.NoError(t, err)
	task2, err = JobRegister.NewTask(task2)
	assert.NoError(t, err)

	queue, err := NewQueue(queueName, limit)
	assert.NoError(t, err)

	err = queue.Enqueue(task1)
	assert.NoError(t, err)

	worker, err := NewWorker(queueName, 10)
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	worker.Start(ctx)

	err = queue.Enqueue(task2)
	assert.NoError(t, err)

	time.Sleep(time.Duration(200) * time.Millisecond)
	cancel()
}
