package trigger

import (
	"reflect"
	"sync"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
)

type eventTrigger struct {
	list sync.Map
}

type EventFunc struct {
	Args  []interface{} //event func register with arg
	Event interface{}   //exec func
}

var triggerCall eventTrigger

func init() {
	triggerCall = eventTrigger{list: sync.Map{}}
}

func Register(triggerName string, eventFunc EventFunc) error {
	fc := reflect.ValueOf(eventFunc.Event)
	if fc.Kind() != reflect.Func {
		return errors.Errorf("eventFuc %s is not func", eventFunc.Event)
	}
	triggerCall.list.Store(triggerName, eventFunc)
	return nil
}

func Exec(triggerName string, params ...interface{}) ([]reflect.Value, error) {
	if execFunc, ok := triggerCall.list.Load(triggerName); ok {
		f, ok := execFunc.(EventFunc)
		if !ok {
			log.L().Error("data is not a eventfuc")
			return nil, errors.Trace(errors.New("data is not a eventfunc"))
		}
		fc := reflect.ValueOf(f.Event)

		paramsNum := fc.Type().NumIn()
		if len(params)+len(f.Args) != paramsNum {
			log.L().Error("event " + triggerName + " params not enough")
			return nil, errors.Trace(errors.Errorf("event %s  params not enough", triggerName))
		}
		in := make([]reflect.Value, paramsNum)
		k := 0
		for _, param := range f.Args {
			in[k] = reflect.ValueOf(param)
			k++
		}
		for _, param := range params {
			in[k] = reflect.ValueOf(param)
			k++
		}
		result := fc.Call(in)
		return result, nil
	}
	log.L().Error("event" + triggerName + " func not exist")
	return nil, errors.Trace(errors.Errorf("event %s func not exit", triggerName))

}

func SyncExec(event string, params ...interface{}) {
	go func() ([]reflect.Value, error) {
		defer func() {
			if r := recover(); r != nil {
				log.L().Error("sync exec trigger error  " + event)
				return
			}
		}()
		return Exec(event, params)
	}()

}
