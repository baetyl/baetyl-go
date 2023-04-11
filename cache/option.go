package cache

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Option represents the optional function.
type Option func(c *Config)

// WithLogger set the custom logger
func WithLogger(l Logger) Option {
	return func(c *Config) {
		if l != nil {
			c.logger = l
		}
	}
}

// Logger define the logger interface
type Logger interface {
	Errorf(string, ...interface{})
}

// Discard the default logger that will discard all logs of gin-cache
type Discard struct {
}

// Errorf will output the log at error level
func (l Discard) Errorf(string, ...interface{}) {
}

// WithCacheStrategyByRequest set up the custom strategy by per request
func WithCacheStrategyByRequest(getGetCacheStrategyByRequest GetCacheStrategyByRequest) Option {
	return func(c *Config) {
		if getGetCacheStrategyByRequest != nil {
			c.getCacheStrategyByRequest = getGetCacheStrategyByRequest
		}
	}
}

// OnHitCacheCallback define the callback when use cache
type OnHitCacheCallback func(c *gin.Context)

var defaultHitCacheCallback = func(c *gin.Context) {}

// WithOnHitCache will be called when cache hit.
func WithOnHitCache(cb OnHitCacheCallback) Option {
	return func(c *Config) {
		if cb != nil {
			c.hitCacheCallback = cb
		}
	}
}

// OnMissCacheCallback define the callback when use cache
type OnMissCacheCallback func(c *gin.Context)

var defaultMissCacheCallback = func(c *gin.Context) {}

// WithOnMissCache will be called when cache miss.
func WithOnMissCache(cb OnMissCacheCallback) Option {
	return func(c *Config) {
		if cb != nil {
			c.missCacheCallback = cb
		}
	}
}

type BeforeReplyWithCacheCallback func(c *gin.Context, cache *ResponseCache)

var defaultBeforeReplyWithCacheCallback = func(c *gin.Context, cache *ResponseCache) {}

// WithBeforeReplyWithCache will be called before replying with cache.
func WithBeforeReplyWithCache(cb BeforeReplyWithCacheCallback) Option {
	return func(c *Config) {
		if cb != nil {
			c.beforeReplyWithCacheCallback = cb
		}
	}
}

// OnShareSingleFlightCallback define the callback when share the singleflight result
type OnShareSingleFlightCallback func(c *gin.Context)

var defaultShareSingleFlightCallback = func(c *gin.Context) {}

// WithOnShareSingleFlight will be called when share the singleflight result
func WithOnShareSingleFlight(cb OnShareSingleFlightCallback) Option {
	return func(c *Config) {
		if cb != nil {
			c.shareSingleFlightCallback = cb
		}
	}
}

// WithSingleFlightForgetTimeout to reduce the impact of long tail requests.
// singleflight.Forget will be called after the timeout has reached for each backend request when timeout is greater than zero.
func WithSingleFlightForgetTimeout(forgetTimeout time.Duration) Option {
	return func(c *Config) {
		if forgetTimeout > 0 {
			c.singleFlightForgetTimeout = forgetTimeout
		}
	}
}

// IgnoreQueryOrder will ignore the queries order in url when generate cache key . This option only takes effect in CacheByRequestURI function
func IgnoreQueryOrder() Option {
	return func(c *Config) {
		c.ignoreQueryOrder = true
	}
}

// WithPrefixKey will prefix the key
func WithPrefixKey(prefix string) Option {
	return func(c *Config) {
		c.prefixKey = prefix
	}
}

func WithoutHeader() Option {
	return func(c *Config) {
		c.withoutHeader = true
	}
}

// WithoutHeaderIgnore Only effective when without is true, including header fields that will not be ignored
// Keys that are not in the array will still be ignored
func WithoutHeaderIgnore(ks []string) Option {
	return func(c *Config) {
		c.withoutHeaderIgnore = ks
	}
}

func KeyWithGinContext(ks []string) Option {
	return func(c *Config) {
		c.keyWithGinContext = ks
	}
}
