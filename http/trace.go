package http

import (
	"time"

	"github.com/gin-contrib/cache/persistence"
)

const (
	KeyTraceKey      = "TraceKey"
	ValueTraceKey    = "requestId"
	KeyTraceHeader   = "TraceHeader"
	ValueTraceHeader = "x-bce-request-id" // TODO: change to x-baetyl-request-id when support configuration
)

var cache = persistence.NewInMemoryStore(time.Minute * 10)

func GetTraceKey() string {
	res := ValueTraceKey
	cache.Get(KeyTraceKey, &res)
	return res
}

func SetTraceHeader(v string) {
	cache.Set(KeyTraceHeader, v, -1)
}

func GetTraceHeader() string {
	res := ValueTraceHeader
	cache.Get(KeyTraceHeader, &res)
	return res
}
