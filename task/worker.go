package task

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
)

const (
	RatePeriod = 100 * time.Millisecond
)

var (
	ErrInvalidArgs = errors.New("failed to exec task, due to invalid args")
)

type taskWorker struct {
	broker          TaskBroker
	backend         TaskBackend
	registeredTasks map[string]interface{}
	cancel          context.CancelFunc
	rateLimitPeriod time.Duration
	lock            sync.RWMutex
	log             *log.Logger
}

func NewTaskWorker(broker TaskBroker, backend TaskBackend) TaskWorker {
	return &taskWorker{
		broker:          broker,
		backend:         backend,
		registeredTasks: map[string]interface{}{},
		rateLimitPeriod: RatePeriod,
		log:             log.L().With(log.Any("task", "worker")),
	}
}

func (w *taskWorker) StartWorker(ctx context.Context) {
	var workerCtx context.Context
	workerCtx, w.cancel = context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(w.rateLimitPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				taskMsg, err := w.broker.GetMessage()
				if err != nil || taskMsg == nil {
					continue
				}
				decodedMsg, err := taskMsg.Decode()
				if err != nil {
					w.log.Error("failed to decode message ", log.Error(err))
					continue
				}
				resultMsg, err := w.runTask(decodedMsg)
				if err != nil {
					w.log.Error("failed to run task ", log.Error(err))
					continue
				}
				if resultMsg.Result != nil {
					err = w.backend.SetResult(taskMsg.ID, resultMsg)
					if err != nil {
						w.log.Error("failed to set result ", log.Error(err))
					}
				}
			}
		}
	}()
}

func (w *taskWorker) StopWorker() {
	w.cancel()
}

func (w *taskWorker) Register(name string, task interface{}) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.registeredTasks[name] = task
}

func (w *taskWorker) getTask(name string) interface{} {
	w.lock.RLock()
	defer w.lock.RUnlock()
	task, ok := w.registeredTasks[name]
	if !ok {
		return nil
	}
	return task
}

func (w *taskWorker) runTask(msg *TaskMessage) (*ResultMessage, error) {
	if msg.Expires != nil && msg.Expires.UTC().Before(time.Now().UTC()) {
		return nil, fmt.Errorf("task %s is expired on %s", msg.ID, msg.Expires)
	}
	if msg.Args == nil {
		return nil, fmt.Errorf("task %s is malformed - args cannot be nil", msg.ID)
	}
	task := w.getTask(msg.Name)
	if task == nil {
		return nil, fmt.Errorf("task %s is not registered", msg.Name)
	}
	taskInterface, ok := task.(AsyncTask)
	// If realize paresKwargs or RunTask function
	if ok {
		if err := taskInterface.ParseKwargs(msg.Kwargs); err != nil {
			return nil, err
		}
		val, err := taskInterface.RunTask()
		result := &ResultMessage{
			ID:        msg.ID,
			Status:    TaskSuccess,
			Traceback: "",
			Result:    val,
		}
		if err != nil {
			result.Status = TaskFail
			result.Traceback = err.Error()
		}
		return result, nil
	}

	taskFunc := reflect.ValueOf(task)
	return runTaskFunc(&taskFunc, msg)
}

func runTaskFunc(taskFunc *reflect.Value, msg *TaskMessage) (*ResultMessage, error) {
	numArgs := taskFunc.Type().NumIn()
	msgNumArgs := len(msg.Args)
	if numArgs != msgNumArgs {
		return nil, ErrInvalidArgs
	}
	params := make([]reflect.Value, msgNumArgs)
	for i, arg := range msg.Args {
		origType := taskFunc.Type().In(i).Kind()
		msgType := reflect.TypeOf(arg).Kind()
		// special case - convert float64 to int if applicable
		// this is due to json limitation where all numbers are converted to float64
		if origType == reflect.Int && msgType == reflect.Float64 {
			arg = int(arg.(float64))
		}
		if origType == reflect.Float32 && msgType == reflect.Float64 {
			arg = float32(arg.(float64))
		}
		params[i] = reflect.ValueOf(arg)
	}

	res := taskFunc.Call(params)
	result := &ResultMessage{
		ID:        msg.ID,
		Status:    TaskSuccess,
		Traceback: "",
	}
	if len(res) == 0 {
		return result, nil
	}

	result.Result = GetRealValue(&res[0])
	errorResult := res[1]
	if !errorResult.IsNil() {
		result.Status = TaskFail
		result.Traceback = errorResult.Interface().(error).Error()
	}

	return result, nil
}
