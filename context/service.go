package context

import (
	"os"
	"runtime/debug"

	"github.com/baetyl/baetyl-go/log"
)

// Run service
func Run(handle func(Context) error) {
	c := newContext()
	defer func() {
		if r := recover(); r != nil {
			c.log.Error("service is stopped with panic", log.Any("panic", debug.Stack()))
		}
	}()
	c.log.Info("service starting", log.Any("args", os.Args))
	err := handle(c)
	if err != nil {
		c.log.Error("service has stopped with error", log.Error(err))
	} else {
		c.log.Info("service has stopped")
	}
}
