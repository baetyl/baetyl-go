package task

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl-go/v2/errors"
)

var (
	ErrArgs   = errors.New("failed to parse args, due to unknown args")
	ErrFormat = errors.New("failed to parse args, due to err format")
)

func TestTaskBasic(t *testing.T) {
	broker := NewChannelBroker(10)
	backend := NewMapBackend()
	producer := NewTaskProducer(broker, backend)
	worker := NewTaskWorker(broker, backend)

	addTask := "Add"
	worker.Register(addTask, Add)

	addWithKey := "AddKey"
	worker.Register(addWithKey, &addInt{})

	worker.StartWorker(context.Background())
	defer worker.StopWorker()

	asyncResult1, err := producer.AddTask(addTask, 1, 2)
	assert.NoError(t, err)

	asyncResult2, err := producer.AddTaskWithKey(addWithKey, map[string]interface{}{
		"a": 1,
		"b": "2",
	})
	assert.NoError(t, err)

	result1, err := asyncResult1.Get(time.Second)
	assert.NoError(t, err)
	assert.Equal(t, result1.Status, TaskSuccess)
	assert.Equal(t, result1.Result, int64(3))

	result2, err := asyncResult2.Get(time.Second)
	assert.NoError(t, err)
	assert.Equal(t, result2.Status, TaskSuccess)
	assert.Equal(t, result2.Result, 3)
}

func Add(a, b int) (int, error) {
	return a + b, nil
}

type addInt struct {
	a int
	b string
}

func (a *addInt) ParseKwargs(kwargs map[string]interface{}) error {
	argA, ok := kwargs["a"]
	if !ok {
		return ErrArgs
	}
	argAFloat, ok := argA.(float64)
	if !ok {
		return ErrFormat
	}
	a.a = int(argAFloat)
	argB, ok := kwargs["b"]
	if !ok {
		return ErrArgs
	}
	a.b, ok = argB.(string)
	if !ok {
		return ErrFormat
	}
	return nil
}

func (a *addInt) RunTask() (interface{}, error) {
	b, err := strconv.Atoi(a.b)
	if err != nil {
		return nil, err
	}
	return a.a + b, nil
}

func DoNothing() {}

func ValueBool(a bool) (bool, error) {
	return a, nil
}

func ValueFloat(a float32) (float32, error) {
	return a, nil
}

func ValueString(a string) (string, error) {
	return a, nil
}

func ValueMap(a string) (map[string]string, error) {
	return map[string]string{"result": a}, nil
}

func TestTaskMultiValues(t *testing.T) {
	broker := NewChannelBroker(10)
	backend := NewMapBackend()
	producer := NewTaskProducer(broker, backend)
	worker := NewTaskWorker(broker, backend)

	blankTask := "blank"
	worker.Register(blankTask, DoNothing)
	boolTask := "bool"
	worker.Register(boolTask, ValueBool)
	floatTask := "float"
	worker.Register(floatTask, ValueFloat)
	stringTask := "string"
	worker.Register(stringTask, ValueString)
	mapTask := "map"
	worker.Register(mapTask, ValueMap)

	worker.StartWorker(context.Background())
	defer worker.StopWorker()

	asyncBlank, err := producer.AddTask(blankTask)
	assert.NoError(t, err)
	asyncBool, err := producer.AddTask(boolTask, true)
	assert.NoError(t, err)
	asyncFloat, err := producer.AddTask(floatTask, float32(1))
	assert.NoError(t, err)
	asyncString, err := producer.AddTask(stringTask, "test")
	assert.NoError(t, err)
	asyncMap, err := producer.AddTask(mapTask, "test")
	assert.NoError(t, err)

	resultBool, err := asyncBool.Get(time.Second)
	assert.NoError(t, err)
	assert.Equal(t, resultBool.Result, true)
	resultFloat, err := asyncFloat.Get(time.Second)
	assert.NoError(t, err)
	assert.Equal(t, resultFloat.Result, float64(1))
	resultString, err := asyncString.Get(time.Second)
	assert.NoError(t, err)
	assert.Equal(t, resultString.Result, "test")
	resultMap, err := asyncMap.Get(time.Second)
	assert.NoError(t, err)
	assert.Equal(t, resultMap.Result, map[string]string{"result": "test"})

	_, err = asyncBlank.Get(time.Millisecond)
	assert.NotNil(t, err)
}
