package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/appengine/memcache"
)

var (
	// ErrCacheMiss means the item was not found in the cache.
	ErrCacheMiss = errors.New("Not in cache")
)

// Cache stores key/value pairs inexpensively.
type Cache interface {
	// Get gets a value for a key. Returns ErrCacheMiss if not found.
	Get(ctx context.Context, key string) (string, error)
	// Set sets a value for a key. Use timeout=0 for "forever".
	// Of course, since this is a "cache", nothing is guaranted and the item may disappear at any time.
	Set(ctx context.Context, key, value string, timeout time.Duration) error
	// Clear clears the cache key.
	Clear(ctx context.Context, key string) error
}

type memcacheCache struct{}

var _ Cache = (*memcacheCache)(nil) // Ensure interface is implemented.

// NewMemcacheCache creates a cache backed by appengine.memcache.
func NewMemcacheCache() Cache {
	return &memcacheCache{}
}

func (c *memcacheCache) Get(ctx context.Context, key string) (string, error) {
	item, err := memcache.Get(ctx, key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return "", ErrCacheMiss
		}
		return "", fmt.Errorf("memcache.Get(): %v", err)
	}
	return string(item.Value), nil
}

func (c *memcacheCache) Set(ctx context.Context, key, value string, timeout time.Duration) error {
	item := &memcache.Item{
		Key:        key,
		Value:      []byte(value),
		Expiration: timeout,
	}
	if err := memcache.Set(ctx, item); err != nil {
		return fmt.Errorf("memcache.Set(): %v", err)
	}
	return nil
}

func (c *memcacheCache) Clear(ctx context.Context, key string) error {
	if err := memcache.Delete(ctx, key); err != nil {
		return fmt.Errorf("memcache.Delete(): %v", err)
	}
	return nil
}
