package persist

import (
	"reflect"
	"time"

	"github.com/robfig/go-cache"
)

// InMemoryStore represents the cache with memory persistence
type InMemoryStore struct {
	cache.Cache
}

// NewInMemoryStore returns a InMemoryStore
func NewInMemoryStore(defaultExpiration time.Duration) *InMemoryStore {
	return &InMemoryStore{*cache.New(defaultExpiration, time.Minute)}
}

// Get (see CacheStore interface)
func (c *InMemoryStore) Get(key string, value interface{}) error {
	val, found := c.Cache.Get(key)
	if !found {
		return ErrCacheMiss
	}

	v := reflect.ValueOf(value)
	if v.Type().Kind() == reflect.Ptr && v.Elem().CanSet() {
		v.Elem().Set(reflect.ValueOf(val))
		return nil
	}
	return ErrNotStored
}

// Set (see CacheStore interface)
func (c *InMemoryStore) Set(key string, value interface{}, expires time.Duration) error {
	// NOTE: go-cache understands the values of DEFAULT and FOREVER
	c.Cache.Set(key, value, expires)
	return nil
}

// Delete (see CacheStore interface)
func (c *InMemoryStore) Delete(key string) error {
	if found := c.Cache.Delete(key); !found {
		return ErrCacheMiss
	}
	return nil
}
