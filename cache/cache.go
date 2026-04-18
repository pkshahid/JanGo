package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pkshahid/JanGo/core/settings"
)

var (
	ErrCacheMiss = errors.New("key not found in cache")
)

// Cache defines the standard interface for a caching backend.
type Cache interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any, timeout time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	GetMany(ctx context.Context, keys []string) (map[string]any, error)
	SetMany(ctx context.Context, values map[string]any, timeout time.Duration) error
	DeleteMany(ctx context.Context, keys []string) error
	Incr(ctx context.Context, key string, delta int64) (int64, error)
	Decr(ctx context.Context, key string, delta int64) (int64, error)
	GetOrSet(ctx context.Context, key string, defaultFunc func() any, timeout time.Duration) (any, error)
	Has(ctx context.Context, key string) (bool, error)
}

// BackendFactory is a function that creates a new Cache instance from configuration.
type BackendFactory func(alias string, config settings.CacheConfig) Cache

var (
	factories = make(map[string]BackendFactory)
	caches    = make(map[string]Cache)
	mu        sync.RWMutex
	once      sync.Once
)

// RegisterBackend registers a new caching backend factory.
func RegisterBackend(name string, factory BackendFactory) {
	mu.Lock()
	defer mu.Unlock()
	factories[name] = factory
}

// Init initializes all configured caches based on settings.
func Init() {
	once.Do(func() {
		s := settings.Get()
		mu.Lock()
		defer mu.Unlock()

		for alias, config := range s.CACHES {
			factory, ok := factories[config.Backend]
			if !ok {
				// Default to dummy cache if not found or misconfigured
				factory = factories["DummyCache"]
				if factory == nil {
					continue
				}
			}
			caches[alias] = factory(alias, config)
		}

		// Ensure "default" cache always exists
		if _, ok := caches["default"]; !ok {
			if factory, ok := factories["LocMemCache"]; ok {
				caches["default"] = factory("default", settings.CacheConfig{
					Location: "default",
				})
			}
		}
	})
}

// Get returns a configured Cache by its alias. Returns the default cache if empty alias.
func Get(alias string) Cache {
	Init()

	if alias == "" {
		alias = "default"
	}

	mu.RLock()
	defer mu.RUnlock()

	c, ok := caches[alias]
	if !ok {
		panic(fmt.Sprintf("cache alias %q is not configured", alias))
	}
	return c
}
