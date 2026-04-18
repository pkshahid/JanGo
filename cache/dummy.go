package cache

import (
	"context"
	"time"

	"github.com/pkshahid/JanGo/core/settings"
)

func init() {
	RegisterBackend("DummyCache", NewDummyCache)
}

// DummyCache implements a no-op cache.
type DummyCache struct{}

func NewDummyCache(alias string, config settings.CacheConfig) Cache {
	return &DummyCache{}
}

func (c *DummyCache) Get(ctx context.Context, key string) (any, error) {
	return nil, ErrCacheMiss
}

func (c *DummyCache) Set(ctx context.Context, key string, value any, timeout time.Duration) error {
	return nil
}

func (c *DummyCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (c *DummyCache) Clear(ctx context.Context) error {
	return nil
}

func (c *DummyCache) GetMany(ctx context.Context, keys []string) (map[string]any, error) {
	return make(map[string]any), nil
}

func (c *DummyCache) SetMany(ctx context.Context, values map[string]any, timeout time.Duration) error {
	return nil
}

func (c *DummyCache) DeleteMany(ctx context.Context, keys []string) error {
	return nil
}

func (c *DummyCache) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	return delta, nil
}

func (c *DummyCache) Decr(ctx context.Context, key string, delta int64) (int64, error) {
	return -delta, nil
}

func (c *DummyCache) GetOrSet(ctx context.Context, key string, defaultFunc func() any, timeout time.Duration) (any, error) {
	return defaultFunc(), nil
}

func (c *DummyCache) Has(ctx context.Context, key string) (bool, error) {
	return false, nil
}
