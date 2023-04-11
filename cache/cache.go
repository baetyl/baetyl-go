package cache

import (
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/singleflight"

	"github.com/baetyl/baetyl-go/v2/cache/persist"
)

// Strategy the cache strategy
type Strategy struct {
	CacheKey string

	// CacheStore if nil, use default cache store instead
	CacheStore persist.CacheStore

	// CacheDuration
	CacheDuration time.Duration
}

// GetCacheStrategyByRequest User can this function to design custom cache strategy by request.
// The second return value bool means whether this request should be cached.
// The first return value Strategy determine the special strategy by this request.
type GetCacheStrategyByRequest func(c *gin.Context) (Strategy, bool)

func _cache(defaultCacheStore persist.CacheStore, defaultExpire time.Duration, cfg *Config, isMiddleware bool, handle gin.HandlerFunc) gin.HandlerFunc {
	if cfg.getCacheStrategyByRequest == nil {
		panic("cache strategy is nil")
	}

	sfGroup := singleflight.Group{}

	return func(c *gin.Context) {
		cacheStrategy, shouldCache := cfg.getCacheStrategyByRequest(c)
		if !shouldCache {
			if isMiddleware {
				c.Next()
			} else {
				handle(c)
			}
			return
		}

		cacheKey := cacheStrategy.CacheKey

		if cfg.prefixKey != "" {
			cacheKey = cfg.prefixKey + cacheKey
		}

		if cfg.keyWithGinContext != nil && len(cfg.keyWithGinContext) > 0 {
			for _, k := range cfg.keyWithGinContext {
				cacheKey = c.GetString(k) + cacheKey
			}
		}

		// merge cfg
		cacheStore := defaultCacheStore
		if cacheStrategy.CacheStore != nil {
			cacheStore = cacheStrategy.CacheStore
		}

		cacheDuration := defaultExpire
		if cacheStrategy.CacheDuration > 0 {
			cacheDuration = cacheStrategy.CacheDuration
		}

		// read cache first
		{
			respCache := &ResponseCache{}
			err := cacheStore.Get(cacheKey, &respCache)
			if err == nil {
				replyWithCache(c, cfg, respCache)
				cfg.hitCacheCallback(c)
				return
			}

			if err != persist.ErrCacheMiss {
				cfg.logger.Errorf("get cache error: %s, cache key: %s", err, cacheKey)
			}
			cfg.missCacheCallback(c)
		}

		// cache miss, then call the backend

		// use responseCacheWriter in order to record the response
		cacheWriter := &responseCacheWriter{
			ResponseWriter: c.Writer,
		}
		c.Writer = cacheWriter

		inFlight := false
		rawRespCache, _, _ := sfGroup.Do(cacheKey, func() (interface{}, error) {
			if cfg.singleFlightForgetTimeout > 0 {
				forgetTimer := time.AfterFunc(cfg.singleFlightForgetTimeout, func() {
					sfGroup.Forget(cacheKey)
				})
				defer forgetTimer.Stop()
			}

			if isMiddleware {
				c.Next()
			} else {
				handle(c)
			}

			inFlight = true

			respCache := &ResponseCache{}
			respCache.fillWithCacheWriter(cacheWriter, cfg.withoutHeader, cfg.withoutHeaderIgnore)

			// only cache 2xx response
			if !c.IsAborted() && cacheWriter.Status() < 300 && cacheWriter.Status() >= 200 {
				if err := cacheStore.Set(cacheKey, respCache, cacheDuration); err != nil {
					cfg.logger.Errorf("set cache key error: %s, cache key: %s", err, cacheKey)
				}
			}

			return respCache, nil
		})

		if !inFlight {
			replyWithCache(c, cfg, rawRespCache.(*ResponseCache))
			cfg.shareSingleFlightCallback(c)
		}
	}
}
