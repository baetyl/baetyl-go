package cache

import (
	"bytes"
	"encoding/gob"
	"net/http"

	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(&ResponseCache{})
}

// ResponseCache record the http response cache
type ResponseCache struct {
	Status int
	Header http.Header
	Data   []byte
}

func (c *ResponseCache) fillWithCacheWriter(cacheWriter *responseCacheWriter, withoutHeader bool, withoutHeaderIgnore []string) {
	c.Status = cacheWriter.Status()
	c.Data = cacheWriter.body.Bytes()
	if !withoutHeader {
		c.Header = cacheWriter.Header().Clone()
	} else {
		if c.Header == nil {
			c.Header = http.Header{}
		}
		for _, k := range withoutHeaderIgnore {
			c.Header.Set(k, cacheWriter.Header().Get(k))
		}
	}
}

// responseCacheWriter
type responseCacheWriter struct {
	gin.ResponseWriter

	body bytes.Buffer
}

func (w *responseCacheWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseCacheWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func replyWithCache(c *gin.Context, cfg *Config, respCache *ResponseCache) {
	cfg.beforeReplyWithCacheCallback(c, respCache)

	c.Writer.WriteHeader(respCache.Status)

	if !cfg.withoutHeader {
		for key, values := range respCache.Header {
			for _, val := range values {
				c.Writer.Header().Set(key, val)
			}
		}
	} else {
		for _, key := range cfg.withoutHeaderIgnore {
			c.Writer.Header().Set(key, respCache.Header.Get(key))
		}
	}

	if _, err := c.Writer.Write(respCache.Data); err != nil {
		cfg.logger.Errorf("write response error: %s", err)
	}

	// abort handler chain and return directly
	c.Abort()
}
