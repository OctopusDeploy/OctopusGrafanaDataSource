package plugin

import (
	"sync"
	"time"
)

// maxCacheEntries and maxCacheBytes bound the memory held by a single
// datasource instance. Individual responses are already capped at
// maxResponseBytes, which is below the byte budget, so a single entry can
// always be stored.
const maxCacheEntries = 1000
const maxCacheBytes = 256 << 20

type cacheEntry struct {
	value   []byte
	failed  bool
	expires time.Time
}

// ttlCache is a small time-based cache scoped to a single datasource
// instance, so entries can never be shared between different servers or
// credentials.
type ttlCache struct {
	mu         sync.Mutex
	entries    map[string]cacheEntry
	totalBytes int
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
		c.deleteLocked(key)
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

	c.deleteLocked(key)

	if len(c.entries) >= maxCacheEntries || c.totalBytes+len(value) > maxCacheBytes {
		now := time.Now()
		for k, v := range c.entries {
			if now.After(v.expires) {
				c.deleteLocked(k)
			}
		}
		if len(c.entries) >= maxCacheEntries || c.totalBytes+len(value) > maxCacheBytes {
			c.entries = map[string]cacheEntry{}
			c.totalBytes = 0
		}
	}

	c.entries[key] = cacheEntry{value: value, failed: failed, expires: time.Now().Add(ttl)}
	c.totalBytes += len(value)
}

// deleteLocked removes an entry and its byte count. The caller must hold mu.
func (c *ttlCache) deleteLocked(key string) {
	if entry, ok := c.entries[key]; ok {
		c.totalBytes -= len(entry.value)
		delete(c.entries, key)
	}
}
