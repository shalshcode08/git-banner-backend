package github

import (
	"sync"
	"time"
)

type cacheEntry struct {
	value     any
	expiresAt time.Time
}

// Cache is a simple in-memory TTL cache backed by sync.Map.
type Cache struct {
	m   sync.Map
	ttl time.Duration
}

// NewCache creates a Cache with the given TTL.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{ttl: ttl}
}

// Get returns the cached value and true if it exists and has not expired.
func (c *Cache) Get(key string) (any, bool) {
	v, ok := c.m.Load(key)
	if !ok {
		return nil, false
	}
	entry := v.(cacheEntry)
	if time.Now().After(entry.expiresAt) {
		c.m.Delete(key)
		return nil, false
	}
	return entry.value, true
}

// Set stores value under key with the cache TTL.
func (c *Cache) Set(key string, value any) {
	c.m.Store(key, cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	})
}
