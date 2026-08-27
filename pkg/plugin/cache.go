package plugin

import (
	"sync"
	"time"
)

// maxCacheEntries bounds the memory held by a single datasource instance.
const maxCacheEntries = 1000

type cacheEntry struct {
	value   []byte
	failed  bool
	expires time.Time
}

// ttlCache is a small time-based cache scoped to a single datasource
// instance, so entries can never be shared between different servers or
// credentials.
type ttlCache struct {
	mu      sync.Mutex
	entries map[string]cacheEntry
}

func newTTLCache() *ttlCache {
	return &ttlCache{entries: map[string]cacheEntry{}}
}

// get returns the cached value and whether the entry recorded a failure.
func (c *ttlCache) get(key string) ([]byte, bool, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false, false
	}

	if time.Now().After(entry.expires) {
		delete(c.entries, key)
		return nil, false, false
	}

	return entry.value, entry.failed, true
}

func (c *ttlCache) set(key string, value []byte, failed bool, ttl time.Duration) {
	if ttl <= 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= maxCacheEntries {
		now := time.Now()
		for k, v := range c.entries {
			if now.After(v.expires) {
				delete(c.entries, k)
			}
		}
		if len(c.entries) >= maxCacheEntries {
			c.entries = map[string]cacheEntry{}
		}
	}

	c.entries[key] = cacheEntry{value: value, failed: failed, expires: time.Now().Add(ttl)}
}
