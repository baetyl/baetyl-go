package context

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/baetyl/baetyl-go/log"
)

// Run service
func Run(handle func(Context) error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "service is stopped with panic: %s\n%s", r, string(debug.Stack()))
		}
	}()
	c, err := newContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s][%s] failed to create context: %s\n", c.sn, c.in, err.Error())
		c.log.Error("failed to create context", log.Error(err))
		return
	}
	c.log.Info("service starting", log.Any("args", os.Args))
	err = handle(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s][%s] service is stopped with error: %s\n", c.sn, c.in, err.Error())
		c.log.Error("service is stopped with error", log.Error(err))
	} else {
		c.log.Info("service stopped")
	}
}
