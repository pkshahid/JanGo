package cache

import (
	"context"
	"time"

	"github.com/godjango/godjango/core/settings"
	"github.com/redis/go-redis/v9"
)

func init() {
	RegisterBackend("RedisCache", NewRedisCache)
}

type RedisCache struct {
	client *redis.Client
	prefix string
}

func NewRedisCache(alias string, config settings.CacheConfig) Cache {
	opt := &redis.Options{
		Addr: config.Location,
	}
	if pass, ok := config.Options["PASSWORD"].(string); ok {
		opt.Password = pass
	}
	if db, ok := config.Options["DB"].(int); ok {
		opt.DB = db
	}

	prefix, _ := config.Options["KEY_PREFIX"].(string)

	return &RedisCache{
		client: redis.NewClient(opt),
		prefix: prefix,
	}
}

func (c *RedisCache) makeKey(key string) string {
	if c.prefix != "" {
		return c.prefix + ":" + key
	}
	return key
}

func (c *RedisCache) Get(ctx context.Context, key string) (any, error) {
	val, err := c.client.Get(ctx, c.makeKey(key)).Bytes()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}
	return Deserialize(val)
}

func (c *RedisCache) Set(ctx context.Context, key string, value any, timeout time.Duration) error {
	data, err := Serialize(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.makeKey(key), data, timeout).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.makeKey(key)).Err()
}

func (c *RedisCache) Clear(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

func (c *RedisCache) GetMany(ctx context.Context, keys []string) (map[string]any, error) {
	prefixedKeys := make([]string, len(keys))
	for i, k := range keys {
		prefixedKeys[i] = c.makeKey(k)
	}

	vals, err := c.client.MGet(ctx, prefixedKeys...).Result()
	if err != nil {
		return nil, err
	}

	results := make(map[string]any)
	for i, k := range keys {
		if vals[i] != nil {
			if strVal, ok := vals[i].(string); ok {
				if res, err := Deserialize([]byte(strVal)); err == nil {
					results[k] = res
				}
			}
		}
	}
	return results, nil
}

func (c *RedisCache) SetMany(ctx context.Context, values map[string]any, timeout time.Duration) error {
	pipe := c.client.Pipeline()
	for k, v := range values {
		data, err := Serialize(v)
		if err != nil {
			return err
		}
		pipe.Set(ctx, c.makeKey(k), data, timeout)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisCache) DeleteMany(ctx context.Context, keys []string) error {
	prefixedKeys := make([]string, len(keys))
	for i, k := range keys {
		prefixedKeys[i] = c.makeKey(k)
	}
	return c.client.Del(ctx, prefixedKeys...).Err()
}

func (c *RedisCache) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	return c.client.IncrBy(ctx, c.makeKey(key), delta).Result()
}

func (c *RedisCache) Decr(ctx context.Context, key string, delta int64) (int64, error) {
	return c.client.DecrBy(ctx, c.makeKey(key), delta).Result()
}

func (c *RedisCache) GetOrSet(ctx context.Context, key string, defaultFunc func() any, timeout time.Duration) (any, error) {
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

func (c *RedisCache) Has(ctx context.Context, key string) (bool, error) {
	res, err := c.client.Exists(ctx, c.makeKey(key)).Result()
	if err != nil {
		return false, err
	}
	return res > 0, nil
}
