package storage

import (
	"sync"
	"time"
)

type Cache interface {
	Get(string, any) bool
	Set(string, any, time.Duration)
}

type cacheEntry struct {
	value      any
	expiration time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{items: map[string]cacheEntry{}}
}

func (c *MemoryCache) Get(key string, dest any) bool {
	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(entry.expiration) {
		return false
	}

	switch out := dest.(type) {
	case *string:
		v, ok := entry.value.(string)
		if !ok {
			return false
		}
		*out = v
		return true
	case *[]byte:
		v, ok := entry.value.([]byte)
		if !ok {
			return false
		}
		cp := make([]byte, len(v))
		copy(cp, v)
		*out = cp
		return true
	default:
		return false
	}
}

func (c *MemoryCache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cacheEntry{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
}
