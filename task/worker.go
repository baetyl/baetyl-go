package task

import (
	"context"
	"reflect"
	"time"

	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/models"
)

type Worker struct {
	queue        *Queue
	scheduleTime time.Duration
}

// NewWorker New a worker mounted to the specific queue.
// PARAMS:
//   - queueName: specific queue name
//   - scheduleTime: fetch task gap
// RETURNS:
//   Worker: pointer to the new queue
//   error: if has error else nil
func NewWorker(queueName string, scheduleTime int) (*Worker, error) {
	queue, err := GetQueue(queueName)
	if err != nil {
		return nil, err
	}
	return &Worker {
		queue: queue,
		scheduleTime: time.Duration(scheduleTime) * time.Millisecond,
	}, nil
}

// Start start a goroutine and stop when ctx cancel.
func (w *Worker) Start(ctx context.Context) {
	go w.run(ctx)
}

func (w *Worker) run(ctx context.Context)  {
	timer := time.NewTimer(w.scheduleTime)
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			task, err := w.queue.Dequeue()
			if err != nil {
				continue
			}
			handleTask(task)
			timer.Reset(w.scheduleTime)
		}
	}
}

func handleTask(task *models.Task) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if ok {
				log.L().Error("handle a panic", log.Code(err))
			}
		}
	}()
	runHandler(task)
}

func runHandler(task *models.Task)  {
	JobRegister.JobMapLock.RLock()
	defer JobRegister.JobMapLock.RUnlock()

	handlers, ok := JobRegister.JobMap[task.JobName]
	if !ok {
		return
	}
	args := reflectArgs(task.Args)
	for _, handler := range handlers {
		results := handler.Call(args)
		if len(results) == 0 {
			log.L().Error("function has no return error")
			return
		}
		errorResult := results[len(results) - 1]
		if !errorResult.IsNil() {
			log.L().Error("job has an error.", log.Code(errorResult.Interface().(error)))
			return
		}
		args = make([]reflect.Value, len(results) - 1)
		copy(args, results)
	}
}

func reflectArgs(args []models.Arg) []reflect.Value {
	argValues := make([]reflect.Value, len(args))
	for i, arg := range args {
		argValue := reflect.ValueOf(arg.Value)
		argValues[i] = argValue
	}
	return argValues
}
