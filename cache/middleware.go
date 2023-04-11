package cache

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/baetyl/baetyl-go/v2/cache/persist"
)

// MCache user must pass getCacheKey to describe the way to generate cache key
func MCache(defaultCacheStore persist.CacheStore, defaultExpire time.Duration, opts ...Option) gin.HandlerFunc {
	cfg := newConfigByOpts(opts...)
	return mCache(defaultCacheStore, defaultExpire, cfg)
}

func mCache(defaultCacheStore persist.CacheStore, defaultExpire time.Duration, cfg *Config) gin.HandlerFunc {
	return _cache(defaultCacheStore, defaultExpire, cfg, true, nil)
}

// MCacheByRequestURI a shortcut function for caching response by uri
func MCacheByRequestURI(defaultCacheStore persist.CacheStore, defaultExpire time.Duration, opts ...Option) gin.HandlerFunc {
	cfg := newConfigByOpts(opts...)
	cfg.setRequestURI()
	return mCache(defaultCacheStore, defaultExpire, cfg)
}

// MCacheByRequestPath a shortcut function for caching response by url path, means will discard the query params
func MCacheByRequestPath(defaultCacheStore persist.CacheStore, defaultExpire time.Duration, opts ...Option) gin.HandlerFunc {
	opts = append(opts, WithCacheStrategyByRequest(func(c *gin.Context) (Strategy, bool) {
		return Strategy{
			CacheKey: c.Request.URL.Path,
		}, true
	}))

	return MCache(defaultCacheStore, defaultExpire, opts...)
}
