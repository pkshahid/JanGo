package cache

import (
	"container/heap"
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/godjango/godjango/core/settings"
)

func init() {
	RegisterBackend("LocMemCache", NewLocMemCache)
}

// LocMemCache provides a fast, thread-safe, in-memory cache.
type LocMemCache struct {
	items    sync.Map
	expiry   *expiryHeap
	expiryMu sync.Mutex
	cleanup  *time.Ticker
	stop     chan struct{}
}

// expiryItem represents an item on the priority queue.
type expiryItem struct {
	key    string
	expire time.Time
	index  int
}

// expiryHeap implements heap.Interface and holds expiryItems.
type expiryHeap []*expiryItem

func (h expiryHeap) Len() int           { return len(h) }
func (h expiryHeap) Less(i, j int) bool { return h[i].expire.Before(h[j].expire) }
func (h expiryHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}
func (h *expiryHeap) Push(x any) {
	n := len(*h)
	item := x.(*expiryItem)
	item.index = n
	*h = append(*h, item)
}
func (h *expiryHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]
	return item
}

type locMemItem struct {
	value  any
	expire time.Time
	ei     *expiryItem
	mu     sync.Mutex
}

func NewLocMemCache(alias string, config settings.CacheConfig) Cache {
	h := &expiryHeap{}
	heap.Init(h)

	c := &LocMemCache{
		expiry:  h,
		cleanup: time.NewTicker(time.Second * 5), // Background cleanup interval
		stop:    make(chan struct{}),
	}

	// Start background cleanup
	go c.startCleanup()
	return c
}

func (c *LocMemCache) startCleanup() {
	for {
		select {
		case <-c.cleanup.C:
			c.cull()
		case <-c.stop:
			c.cleanup.Stop()
			return
		}
	}
}

func (c *LocMemCache) cull() {
	c.expiryMu.Lock()
	defer c.expiryMu.Unlock()

	now := time.Now()
	for c.expiry.Len() > 0 {
		item := (*c.expiry)[0]
		if now.Before(item.expire) {
			break
		}
		// Item has expired, remove from heap and map
		heap.Pop(c.expiry)
		c.items.Delete(item.key)
	}
}

func (c *LocMemCache) Get(ctx context.Context, key string) (any, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	val, ok := c.items.Load(key)
	if !ok {
		return nil, ErrCacheMiss
	}

	item := val.(*locMemItem)
	item.mu.Lock()
	defer item.mu.Unlock()

	// Check expiration inline to be accurate up to the millisecond
	if !item.expire.IsZero() && time.Now().After(item.expire) {
		c.items.Delete(key)
		return nil, ErrCacheMiss
	}

	return item.value, nil
}

func (c *LocMemCache) Set(ctx context.Context, key string, value any, timeout time.Duration) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	var expire time.Time
	if timeout > 0 {
		expire = time.Now().Add(timeout)
	}

	newItem := &locMemItem{
		value:  value,
		expire: expire,
	}

	c.expiryMu.Lock()
	if timeout > 0 {
		ei := &expiryItem{key: key, expire: expire}
		newItem.ei = ei
		heap.Push(c.expiry, ei)
	}

	// Overwrite previous item if it existed
	if oldVal, ok := c.items.LoadOrStore(key, newItem); ok {
		// Key existed, replace value and handle old heap entry
		c.items.Store(key, newItem)
		oldItem := oldVal.(*locMemItem)
		oldItem.mu.Lock()
		if oldItem.ei != nil && oldItem.ei.index != -1 {
			heap.Remove(c.expiry, oldItem.ei.index)
		}
		oldItem.mu.Unlock()
	}
	c.expiryMu.Unlock()

	return nil
}

func (c *LocMemCache) Delete(ctx context.Context, key string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if val, ok := c.items.LoadAndDelete(key); ok {
		item := val.(*locMemItem)
		c.expiryMu.Lock()
		if item.ei != nil && item.ei.index != -1 {
			heap.Remove(c.expiry, item.ei.index)
			item.ei.index = -1
		}
		c.expiryMu.Unlock()
	}
	return nil
}

func (c *LocMemCache) Clear(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.expiryMu.Lock()
	c.items.Clear()
	*c.expiry = make(expiryHeap, 0)
	heap.Init(c.expiry)
	c.expiryMu.Unlock()
	return nil
}

func (c *LocMemCache) GetMany(ctx context.Context, keys []string) (map[string]any, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	results := make(map[string]any)
	for _, key := range keys {
		val, err := c.Get(ctx, key)
		if err == nil {
			results[key] = val
		}
	}
	return results, nil
}

func (c *LocMemCache) SetMany(ctx context.Context, values map[string]any, timeout time.Duration) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	for key, value := range values {
		if err := c.Set(ctx, key, value, timeout); err != nil {
			return err
		}
	}
	return nil
}

func (c *LocMemCache) DeleteMany(ctx context.Context, keys []string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	for _, key := range keys {
		if err := c.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (c *LocMemCache) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	// We need a lock per key or rely on atomic. Using atomic value for int64.
	// But since `value` is `any`, we must serialize it. For LocMem, we can just mutate.
	for {
		val, ok := c.items.Load(key)
		if !ok {
			// Key doesn't exist, start at delta
			if err := c.Set(ctx, key, delta, 0); err != nil {
				return 0, err
			}
			return delta, nil
		}

		item := val.(*locMemItem)
		item.mu.Lock()

		if !item.expire.IsZero() && time.Now().After(item.expire) {
			item.mu.Unlock()
			c.items.Delete(key)
			continue // try again to set new value
		}

		var num int64
		switch v := item.value.(type) {
		case int:
			num = int64(v)
		case int64:
			num = v
		case int32:
			num = int64(v)
		case *int64:
			// Using sync/atomic if stored as pointer
			newVal := atomic.AddInt64(v, delta)
			item.mu.Unlock()
			return newVal, nil
		default:
			item.mu.Unlock()
			return 0, errors.New("value is not an integer")
		}

		num += delta
		item.value = num
		item.mu.Unlock()
		return num, nil
	}
}

func (c *LocMemCache) Decr(ctx context.Context, key string, delta int64) (int64, error) {
	return c.Incr(ctx, key, -delta)
}

func (c *LocMemCache) GetOrSet(ctx context.Context, key string, defaultFunc func() any, timeout time.Duration) (any, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	val, err := c.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	if err != ErrCacheMiss {
		return nil, err
	}

	newVal := defaultFunc()
	err = c.Set(ctx, key, newVal, timeout)
	if err != nil {
		return nil, err
	}

	return newVal, nil
}

func (c *LocMemCache) Has(ctx context.Context, key string) (bool, error) {
	_, err := c.Get(ctx, key)
	if err == nil {
		return true, nil
	}
	if err == ErrCacheMiss {
		return false, nil
	}
	return false, err
}
