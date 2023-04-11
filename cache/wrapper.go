package cache

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/baetyl/baetyl-go/v2/cache/persist"
)

// WCache user must pass getCacheKey to describe the way to generate cache key
func WCache(defaultCacheStore persist.CacheStore, defaultExpire time.Duration, handle gin.HandlerFunc, opts ...Option) gin.HandlerFunc {
	cfg := newConfigByOpts(opts...)
	return wCache(defaultCacheStore, defaultExpire, cfg, handle)
}

func wCache(defaultCacheStore persist.CacheStore, defaultExpire time.Duration, cfg *Config, handle gin.HandlerFunc) gin.HandlerFunc {
	return _cache(defaultCacheStore, defaultExpire, cfg, false, handle)
}

// WCacheByRequestURI a shortcut function for caching response by uri
func WCacheByRequestURI(defaultCacheStore persist.CacheStore, defaultExpire time.Duration, handle gin.HandlerFunc, opts ...Option) gin.HandlerFunc {
	cfg := newConfigByOpts(opts...)
	cfg.setRequestURI()
	return wCache(defaultCacheStore, defaultExpire, cfg, handle)
}

// WCacheByRequestPath a shortcut function for caching response by url path, means will discard the query params
func WCacheByRequestPath(defaultCacheStore persist.CacheStore, defaultExpire time.Duration, handle gin.HandlerFunc, opts ...Option) gin.HandlerFunc {
	opts = append(opts, WithCacheStrategyByRequest(func(c *gin.Context) (Strategy, bool) {
		return Strategy{
			CacheKey: c.Request.URL.Path,
		}, true
	}))

	return WCache(defaultCacheStore, defaultExpire, handle, opts...)
}
