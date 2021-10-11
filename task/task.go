package task

import (
	"context"
	"io"
	"reflect"
)

type TaskProducer interface {
	AddTask(name string, args ...interface{}) (*TaskResult, error)
	AddTaskWithKey(name string, args map[string]interface{}) (*TaskResult, error)
}

type TaskBroker interface {
	SendMessage(*BrokerMessage) error
	GetMessage() (*BrokerMessage, error)
	io.Closer
}

type TaskWorker interface {
	StartWorker(ctx context.Context)
	StopWorker()
	Register(name string, task interface{})
}

type TaskBackend interface {
	GetResult(taskId string) (*ResultMessage, error)
	SetResult(taskID string, result *ResultMessage) error
}

type AsyncTask interface {
	// ParseKwargs - define a method to parse kwargs
	ParseKwargs(map[string]interface{}) error

	// RunTask - define a method for execution
	RunTask() (interface{}, error)
}

func GetRealValue(val *reflect.Value) interface{} {
	if val == nil {
		return nil
	}
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int()
	case reflect.String:
		return val.String()
	case reflect.Bool:
		return val.Bool()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint()
	case reflect.Float32, reflect.Float64:
		return val.Float()
	case reflect.Slice, reflect.Map:
		return val.Interface()
	default:
		return nil
	}
}
