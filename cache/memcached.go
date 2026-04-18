package cache

import (
	"context"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/pkshahid/JanGo/core/settings"
)

func init() {
	RegisterBackend("MemcachedCache", NewMemcachedCache)
}

type MemcachedCache struct {
	client *memcache.Client
	prefix string
}

func NewMemcachedCache(alias string, config settings.CacheConfig) Cache {
	loc := config.Location
	if loc == "" {
		loc = "127.0.0.1:11211"
	}
	prefix, _ := config.Options["KEY_PREFIX"].(string)

	return &MemcachedCache{
		client: memcache.New(loc),
		prefix: prefix,
	}
}

func (c *MemcachedCache) makeKey(key string) string {
	if c.prefix != "" {
		return c.prefix + ":" + key
	}
	return key
}

func (c *MemcachedCache) Get(ctx context.Context, key string) (any, error) {
	item, err := c.client.Get(c.makeKey(key))
	if err == memcache.ErrCacheMiss {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}
	return Deserialize(item.Value)
}

func (c *MemcachedCache) Set(ctx context.Context, key string, value any, timeout time.Duration) error {
	data, err := Serialize(value)
	if err != nil {
		return err
	}

	// memcache expiration is int32 seconds
	expiration := int32(timeout.Seconds())
	if timeout == 0 {
		expiration = 0 // never expire
	}

	item := &memcache.Item{
		Key:        c.makeKey(key),
		Value:      data,
		Expiration: expiration,
	}
	return c.client.Set(item)
}

func (c *MemcachedCache) Delete(ctx context.Context, key string) error {
	return c.client.Delete(c.makeKey(key))
}

func (c *MemcachedCache) Clear(ctx context.Context) error {
	return c.client.FlushAll()
}

func (c *MemcachedCache) GetMany(ctx context.Context, keys []string) (map[string]any, error) {
	prefixedKeys := make([]string, len(keys))
	for i, k := range keys {
		prefixedKeys[i] = c.makeKey(k)
	}

	items, err := c.client.GetMulti(prefixedKeys)
	if err != nil {
		return nil, err
	}

	results := make(map[string]any)
	for _, k := range keys {
		if item, ok := items[c.makeKey(k)]; ok {
			if val, err := Deserialize(item.Value); err == nil {
				results[k] = val
			}
		}
	}
	return results, nil
}

func (c *MemcachedCache) SetMany(ctx context.Context, values map[string]any, timeout time.Duration) error {
	for k, v := range values {
		if err := c.Set(ctx, k, v, timeout); err != nil {
			return err
		}
	}
	return nil
}

func (c *MemcachedCache) DeleteMany(ctx context.Context, keys []string) error {
	for _, k := range keys {
		if err := c.Delete(ctx, k); err != nil && err != memcache.ErrCacheMiss {
			return err
		}
	}
	return nil
}

func (c *MemcachedCache) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	newVal, err := c.client.Increment(c.makeKey(key), uint64(delta))
	if err == memcache.ErrCacheMiss {
		// Initialize
		err = c.client.Set(&memcache.Item{Key: c.makeKey(key), Value: []byte("0")})
		if err != nil {
			return 0, err
		}
		newVal, err = c.client.Increment(c.makeKey(key), uint64(delta))
	}
	return int64(newVal), err
}

func (c *MemcachedCache) Decr(ctx context.Context, key string, delta int64) (int64, error) {
	newVal, err := c.client.Decrement(c.makeKey(key), uint64(delta))
	if err == memcache.ErrCacheMiss {
		// Initialize
		err = c.client.Set(&memcache.Item{Key: c.makeKey(key), Value: []byte("0")})
		if err != nil {
			return 0, err
		}
		// Technically memcached decrements don't go below 0, but mimicking API
		newVal, err = c.client.Decrement(c.makeKey(key), uint64(delta))
	}
	return int64(newVal), err
}

func (c *MemcachedCache) GetOrSet(ctx context.Context, key string, defaultFunc func() any, timeout time.Duration) (any, error) {
	val, err := c.Get(ctx, key)
	if err == nil {
		return val, nil
	}
	if err != ErrCacheMiss {
		return nil, err
	}

	newVal := defaultFunc()
	if err := c.Set(ctx, key, newVal, timeout); err != nil {
		return nil, err
	}
	return newVal, nil
}

func (c *MemcachedCache) Has(ctx context.Context, key string) (bool, error) {
	_, err := c.client.Get(c.makeKey(key))
	if err == memcache.ErrCacheMiss {
		return false, nil
	}
	return err == nil, err
}
