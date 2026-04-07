package cache_test

import (
	"context"
	godjangohttp "github.com/godjango/godjango/http"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/godjango/godjango/cache"
	"github.com/godjango/godjango/core/settings"
)

func initSettings() {
	s := settings.Settings{
		CACHES: map[string]settings.CacheConfig{
			"default": {
				Backend:  "LocMemCache",
				Location: "unique-snowflake",
			},
			"dummy": {
				Backend:  "DummyCache",
				Location: "dummy",
			},
		},
		SECRET_KEY:   "secret",
		ROOT_URLCONF: "urls",
	}
	settings.Configure(s)
}

func TestLocMemCache(t *testing.T) {
	initSettings()
	c := cache.Get("default")
	ctx := context.Background()

	c.Clear(ctx)

	err := c.Set(ctx, "test_key", "test_value", time.Minute*5)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := c.Get(ctx, "test_key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if val.(string) != "test_value" {
		t.Errorf("Expected test_value, got %v", val)
	}

	has, err := c.Has(ctx, "test_key")
	if err != nil || !has {
		t.Errorf("Expected Has to be true")
	}

	c.Delete(ctx, "test_key")
	_, err = c.Get(ctx, "test_key")
	if err != cache.ErrCacheMiss {
		t.Errorf("Expected ErrCacheMiss, got %v", err)
	}

	// Test Incr / Decr
	c.Incr(ctx, "counter", 1)
	val, _ = c.Get(ctx, "counter")
	if val.(int64) != 1 {
		t.Errorf("Expected counter to be 1, got %v", val)
	}

	c.Incr(ctx, "counter", 5)
	val, _ = c.Get(ctx, "counter")
	if val.(int64) != 6 {
		t.Errorf("Expected counter to be 6, got %v", val)
	}

	c.Decr(ctx, "counter", 2)
	val, _ = c.Get(ctx, "counter")
	if val.(int64) != 4 {
		t.Errorf("Expected counter to be 4, got %v", val)
	}

	// Test SetMany
	values := map[string]any{"a": 1, "b": 2, "c": 3}
	c.SetMany(ctx, values, time.Minute)

	manyVal, _ := c.GetMany(ctx, []string{"a", "b", "c", "d"})
	if len(manyVal) != 3 {
		t.Errorf("Expected 3 items in GetMany")
	}

	// Test Expiration
	c.Set(ctx, "expired", "yes", time.Millisecond*10)
	time.Sleep(time.Millisecond * 20)

	_, err = c.Get(ctx, "expired")
	if err != cache.ErrCacheMiss {
		t.Errorf("Expected ErrCacheMiss after expiration, got %v", err)
	}
}

func TestDummyCache(t *testing.T) {
	initSettings()
	c := cache.Get("dummy")
	ctx := context.Background()

	err := c.Set(ctx, "dummy_key", "value", time.Minute)
	if err != nil {
		t.Fatalf("Dummy Set failed")
	}

	_, err = c.Get(ctx, "dummy_key")
	if err != cache.ErrCacheMiss {
		t.Errorf("Expected CacheMiss from DummyCache, got %v", err)
	}
}

func TestCacheDecorators(t *testing.T) {
	initSettings()
	c := cache.Get("default")
	ctx := context.Background()
	c.Clear(ctx)

	callCount := 0
	view := func(req *godjangohttp.Request) godjangohttp.Response {
		callCount++
		return godjangohttp.NewHttpResponse("cached content", 200)
	}

	cachedView := cache.CachePage(time.Minute)(view)

	req, _ := http.NewRequest("GET", "http://test.com/cache-me", nil)
	gdReq := godjangohttp.NewRequest(req)
	gdReq.Context = ctx // Override context

	// First call should execute view
	resp1 := cachedView(gdReq)
	if callCount != 1 {
		t.Errorf("Expected callCount 1, got %d", callCount)
	}

	hr, _ := resp1.(*godjangohttp.HttpResponse)
	body, _ := ioutil.ReadAll(hr.Body)
	if string(body) != "cached content" {
		t.Errorf("Expected cached content, got %s", string(body))
	}

	// Second call should hit cache
	resp2 := cachedView(gdReq)
	if callCount != 1 {
		t.Errorf("Expected callCount to remain 1, got %d", callCount)
	}

	hr2, _ := resp2.(*godjangohttp.HttpResponse)
	body2, _ := ioutil.ReadAll(hr2.Body)
	if string(body2) != "cached content" {
		t.Errorf("Expected cached content from cache, got %s", string(body2))
	}
}

func TestVaryOnHeaders(t *testing.T) {
	view := func(req *godjangohttp.Request) godjangohttp.Response {
		return godjangohttp.NewHttpResponse("content", 200)
	}

	varyView := cache.VaryOnHeaders("Accept-Language", "Cookie")(view)

	req, _ := http.NewRequest("GET", "/", nil)
	gdReq := godjangohttp.NewRequest(req)

	resp := varyView(gdReq)
	hr, _ := resp.(*godjangohttp.HttpResponse)

	vary := hr.Headers.Get("Vary")
	if !strings.Contains(vary, "Accept-Language") || !strings.Contains(vary, "Cookie") {
		t.Errorf("Expected Vary header to contain Accept-Language and Cookie, got %s", vary)
	}
}

func TestVaryOnCookie(t *testing.T) {
	view := func(req *godjangohttp.Request) godjangohttp.Response {
		return godjangohttp.NewHttpResponse("content", 200)
	}

	cookieView := cache.VaryOnCookie()(view)

	req, _ := http.NewRequest("GET", "/", nil)
	gdReq := godjangohttp.NewRequest(req)

	resp := cookieView(gdReq)
	hr, _ := resp.(*godjangohttp.HttpResponse)

	vary := hr.Headers.Get("Vary")
	if vary != "Cookie" {
		t.Errorf("Expected Vary header to be Cookie, got %s", vary)
	}
}

type customStruct struct {
	Foo string
	Bar int
}

func TestSerialization(t *testing.T) {
	cache.RegisterType("cache_test.customStruct", customStruct{})

	c := customStruct{Foo: "hello", Bar: 42}

	data, err := cache.Serialize(c)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	val, err := cache.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	res, ok := val.(customStruct)
	if !ok {
		t.Fatalf("Expected customStruct, got %T", val)
	}

	if res.Foo != "hello" || res.Bar != 42 {
		t.Errorf("Struct data corrupted: %+v", res)
	}

	// Test pointer
	dataPtr, _ := cache.Serialize(&c)
	valPtr, _ := cache.Deserialize(dataPtr)
	resPtr, ok := valPtr.(customStruct)
	if !ok {
		t.Fatalf("Expected customStruct from pointer serialization, got %T", valPtr)
	}
	if resPtr.Foo != "hello" {
		t.Errorf("Pointer struct data corrupted")
	}
}
