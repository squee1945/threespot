package storage

import (
	"context"
	"sync"
	"time"
)

type fakeCache struct {
	mu      sync.Mutex
	entries map[string]*entry
}

var _ Cache = (*memcacheCache)(nil) // Ensure interface is implemented.

type entry struct {
	value      string
	expiration time.Time
}

// NewFakeCache creates simple memory-backed cache.
func NewFakeCache() Cache {
	return &fakeCache{
		entries: make(map[string]*entry),
	}
}

func (c *fakeCache) Get(ctx context.Context, key string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if !ok {
		return "", ErrCacheMiss
	}
	if time.Now().Before(entry.expiration) {
		return entry.value, nil
	}
	delete(c.entries, key)
	return "", ErrCacheMiss
}

func (c *fakeCache) Set(ctx context.Context, key, value string, timeout time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry := &entry{value: value}
	if timeout == 0 {
		entry.expiration = time.Now().Add(365 * 24 * time.Hour)
	} else {
		entry.expiration = time.Now().Add(timeout)
	}
	c.entries[key] = entry
	return nil
}
