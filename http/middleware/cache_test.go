package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkshahid/JanGo/cache"
	"github.com/pkshahid/JanGo/core/settings"
	godjangohttp "github.com/pkshahid/JanGo/http"
)

func setupCacheTestSettings() {
	s := settings.Settings{
		SECRET_KEY:               "secret",
		ROOT_URLCONF:             "test",
		CACHE_MIDDLEWARE_ALIAS:   "default",
		CACHE_MIDDLEWARE_SECONDS: 60,
		CACHES: map[string]settings.CacheConfig{
			"default": {
				Backend:  "LocMemCache",
				Location: "cache-middleware-test",
			},
		},
	}
	settings.Configure(s)
}

func TestCacheMiddleware_BasicCaching(t *testing.T) {
	setupCacheTestSettings()

	c := cache.Get("default")
	c.Clear(context.Background())

	callCount := 0
	view := func(req *godjangohttp.Request) godjangohttp.Response {
		callCount++
		return godjangohttp.NewHttpResponse("cached page", http.StatusOK)
	}

	chain := NewChain(UpdateCacheMiddleware, FetchFromCacheMiddleware)
	handler := chain.Then(view)

	req1 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/page/", nil))
	resp1 := handler(req1)
	hr1 := resp1.(*godjangohttp.HttpResponse)
	body1, _ := io.ReadAll(hr1.Body)
	if string(body1) != "cached page" {
		t.Errorf("Expected 'cached page', got %s", string(body1))
	}
	if callCount != 1 {
		t.Errorf("Expected callCount 1 after first request, got %d", callCount)
	}

	req2 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/page/", nil))
	resp2 := handler(req2)
	hr2 := resp2.(*godjangohttp.HttpResponse)
	body2, _ := io.ReadAll(hr2.Body)
	if string(body2) != "cached page" {
		t.Errorf("Expected 'cached page' from cache, got %s", string(body2))
	}
	if callCount != 1 {
		t.Errorf("Expected callCount to remain 1 after cache hit, got %d", callCount)
	}
}

func TestCacheMiddleware_PostNotCached(t *testing.T) {
	setupCacheTestSettings()

	c := cache.Get("default")
	c.Clear(context.Background())

	callCount := 0
	view := func(req *godjangohttp.Request) godjangohttp.Response {
		callCount++
		return godjangohttp.NewHttpResponse("post result", http.StatusOK)
	}

	chain := NewChain(UpdateCacheMiddleware, FetchFromCacheMiddleware)
	handler := chain.Then(view)

	req1 := godjangohttp.NewRequest(httptest.NewRequest("POST", "/submit/", nil))
	handler(req1)
	req2 := godjangohttp.NewRequest(httptest.NewRequest("POST", "/submit/", nil))
	handler(req2)

	if callCount != 2 {
		t.Errorf("Expected callCount 2 for POST requests (not cached), got %d", callCount)
	}
}

func TestCacheMiddleware_HeadersPreserved(t *testing.T) {
	setupCacheTestSettings()

	c := cache.Get("default")
	c.Clear(context.Background())

	view := func(req *godjangohttp.Request) godjangohttp.Response {
		resp := godjangohttp.NewHttpResponse("with headers", http.StatusOK)
		resp.Headers.Set("X-Custom-Header", "custom-value")
		return resp
	}

	chain := NewChain(UpdateCacheMiddleware, FetchFromCacheMiddleware)
	handler := chain.Then(view)

	req1 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/headers/", nil))
	handler(req1)

	req2 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/headers/", nil))
	resp2 := handler(req2)
	hr2 := resp2.(*godjangohttp.HttpResponse)

	if hr2.Headers.Get("X-Custom-Header") != "custom-value" {
		t.Errorf("Expected X-Custom-Header to be preserved from cache, got %q", hr2.Headers.Get("X-Custom-Header"))
	}
}

func TestCacheMiddleware_VaryHeaders(t *testing.T) {
	setupCacheTestSettings()

	c := cache.Get("default")
	c.Clear(context.Background())

	callCount := 0
	view := func(req *godjangohttp.Request) godjangohttp.Response {
		callCount++
		resp := godjangohttp.NewHttpResponse("vary content", http.StatusOK)
		resp.Headers.Set("Vary", "Accept-Language")
		return resp
	}

	chain := NewChain(UpdateCacheMiddleware, FetchFromCacheMiddleware)
	handler := chain.Then(view)

	rawReq1 := httptest.NewRequest("GET", "/vary/", nil)
	rawReq1.Header.Set("Accept-Language", "en")
	req1 := godjangohttp.NewRequest(rawReq1)
	handler(req1)
	if callCount != 1 {
		t.Fatalf("Expected callCount 1, got %d", callCount)
	}

	rawReq2 := httptest.NewRequest("GET", "/vary/", nil)
	rawReq2.Header.Set("Accept-Language", "en")
	req2 := godjangohttp.NewRequest(rawReq2)
	handler(req2)
	if callCount != 1 {
		t.Errorf("Expected callCount 1 for same Accept-Language (cache hit), got %d", callCount)
	}

	rawReq3 := httptest.NewRequest("GET", "/vary/", nil)
	rawReq3.Header.Set("Accept-Language", "fr")
	req3 := godjangohttp.NewRequest(rawReq3)
	handler(req3)
	if callCount != 2 {
		t.Errorf("Expected callCount 2 for different Accept-Language (cache miss), got %d", callCount)
	}
}

func TestCacheMiddleware_NonCacheableStatus(t *testing.T) {
	setupCacheTestSettings()

	c := cache.Get("default")
	c.Clear(context.Background())

	callCount := 0
	view := func(req *godjangohttp.Request) godjangohttp.Response {
		callCount++
		return godjangohttp.NewHttpResponse("error", http.StatusInternalServerError)
	}

	chain := NewChain(UpdateCacheMiddleware, FetchFromCacheMiddleware)
	handler := chain.Then(view)

	req1 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/error/", nil))
	handler(req1)
	req2 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/error/", nil))
	handler(req2)

	if callCount != 2 {
		t.Errorf("Expected callCount 2 for 500 responses (not cached), got %d", callCount)
	}
}

func TestCacheMiddleware_DifferentURLsCachedSeparately(t *testing.T) {
	setupCacheTestSettings()

	c := cache.Get("default")
	c.Clear(context.Background())

	callCount := 0
	view := func(req *godjangohttp.Request) godjangohttp.Response {
		callCount++
		return godjangohttp.NewHttpResponse("content for "+req.Path, http.StatusOK)
	}

	chain := NewChain(UpdateCacheMiddleware, FetchFromCacheMiddleware)
	handler := chain.Then(view)

	req1 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/page-a/", nil))
	resp1 := handler(req1)
	hr1 := resp1.(*godjangohttp.HttpResponse)
	body1, _ := io.ReadAll(hr1.Body)
	if string(body1) != "content for /page-a/" {
		t.Errorf("Unexpected body: %s", string(body1))
	}

	req2 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/page-b/", nil))
	resp2 := handler(req2)
	hr2 := resp2.(*godjangohttp.HttpResponse)
	body2, _ := io.ReadAll(hr2.Body)
	if string(body2) != "content for /page-b/" {
		t.Errorf("Unexpected body: %s", string(body2))
	}

	if callCount != 2 {
		t.Errorf("Expected callCount 2 for different URLs, got %d", callCount)
	}

	req3 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/page-a/", nil))
	handler(req3)
	if callCount != 2 {
		t.Errorf("Expected callCount 2 after cache hit on first URL, got %d", callCount)
	}
}
