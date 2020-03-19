package context

import (
	"flag"
	"os"
	"runtime/debug"

	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
)

// Run service
func Run(handle func(Context) error) {
	utils.Version()

	var h bool
	var c string
	flag.BoolVar(&h, "h", false, "this help")
	flag.StringVar(&c, "c", "etc/baetyl/service.yml", "the configuration file")
	if h {
		flag.Usage()
		return
	}

	ctx := NewContext(c)
	defer func() {
		if r := recover(); r != nil {
			ctx.Log().Error("service is stopped with panic", log.Any("panic", debug.Stack()))
		}
	}()
	ctx.Log().Info("service starting", log.Any("args", os.Args))
	err := handle(ctx)
	if err != nil {
		ctx.Log().Error("service has stopped with error", log.Error(err))
	} else {
		ctx.Log().Info("service has stopped")
	}
}
