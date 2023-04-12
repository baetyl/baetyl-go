package cache

import (
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Config contains all options
type Config struct {
	logger Logger

	getCacheStrategyByRequest GetCacheStrategyByRequest

	hitCacheCallback  OnHitCacheCallback
	missCacheCallback OnMissCacheCallback

	beforeReplyWithCacheCallback BeforeReplyWithCacheCallback

	singleFlightForgetTimeout time.Duration
	shareSingleFlightCallback OnShareSingleFlightCallback

	ignoreQueryOrder bool

	withoutHeader bool
	// Only effective when without is true, including header fields that will not be ignored
	// Keys that are not in the array will still be ignored
	withoutHeaderIgnore []string

	prefixKey         string
	keyWithGinContext []string
}

func newConfigByOpts(opts ...Option) *Config {
	cfg := &Config{
		logger:                       Discard{},
		hitCacheCallback:             defaultHitCacheCallback,
		missCacheCallback:            defaultMissCacheCallback,
		beforeReplyWithCacheCallback: defaultBeforeReplyWithCacheCallback,
		shareSingleFlightCallback:    defaultShareSingleFlightCallback,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

func (cfg *Config) setRequestURI() {
	var cacheStrategy GetCacheStrategyByRequest
	if cfg.ignoreQueryOrder {
		cacheStrategy = func(c *gin.Context) (Strategy, bool) {
			newUri, err := getRequestUriIgnoreQueryOrder(c.Request.RequestURI)
			if err != nil {
				cfg.logger.Errorf("getRequestUriIgnoreQueryOrder error: %s", err)
				newUri = c.Request.RequestURI
			}

			return Strategy{
				CacheKey: newUri,
			}, true
		}

	} else {
		cacheStrategy = func(c *gin.Context) (Strategy, bool) {
			return Strategy{
				CacheKey: c.Request.RequestURI,
			}, true
		}
	}
	cfg.getCacheStrategyByRequest = cacheStrategy
}

func getRequestUriIgnoreQueryOrder(requestURI string) (string, error) {
	parsedUrl, err := url.ParseRequestURI(requestURI)
	if err != nil {
		return "", err
	}

	values := parsedUrl.Query()

	if len(values) == 0 {
		return requestURI, nil
	}

	queryKeys := make([]string, 0, len(values))
	for queryKey := range values {
		queryKeys = append(queryKeys, queryKey)
	}
	sort.Strings(queryKeys)

	queryVals := make([]string, 0, len(values))
	for _, queryKey := range queryKeys {
		sort.Strings(values[queryKey])
		for _, val := range values[queryKey] {
			queryVals = append(queryVals, queryKey+"="+val)
		}
	}

	return parsedUrl.Path + "?" + strings.Join(queryVals, "&"), nil
}
