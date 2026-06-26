package middleware

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkshahid/JanGo/cache"
	"github.com/pkshahid/JanGo/core/settings"
	godjangohttp "github.com/pkshahid/JanGo/http"
)

// cacheHeaderPrefix is the key prefix for storing Vary header lists.
const cacheHeaderPrefix = "middleware.cache.cache_header."

// cachePagePrefix is the key prefix for storing cached page content.
const cachePagePrefix = "middleware.cache.cache_page."

// getCacheMiddlewareConfig reads cache middleware settings with sensible defaults.
func getCacheMiddlewareConfig() (alias string, timeout time.Duration, keyPrefix string) {
	s := settings.Get()

	alias = s.CACHE_MIDDLEWARE_ALIAS
	if alias == "" {
		alias = "default"
	}

	seconds := s.CACHE_MIDDLEWARE_SECONDS
	if seconds <= 0 {
		seconds = 600
	}
	timeout = time.Duration(seconds) * time.Second

	keyPrefix = s.CACHE_MIDDLEWARE_KEY_PREFIX
	return
}

// generateMiddlewareCacheKey builds a cache key from the request URL, method,
// key prefix, and any Vary headers.
func generateMiddlewareCacheKey(req *godjangohttp.Request, keyPrefix string, varyHeaders []string) string {
	urlStr := req.URL.String()
	for _, h := range varyHeaders {
		urlStr += "|" + h + "=" + req.Header.Get(h)
	}
	hash := md5.Sum([]byte(urlStr))
	return cachePagePrefix + keyPrefix + "." + hex.EncodeToString(hash[:])
}

// generateMiddlewareHeaderKey builds the key under which the Vary header list
// for a given URL is stored.
func generateMiddlewareHeaderKey(req *godjangohttp.Request, keyPrefix string) string {
	hash := md5.Sum([]byte(req.URL.String()))
	return cacheHeaderPrefix + keyPrefix + "." + hex.EncodeToString(hash[:])
}

// FetchFromCacheMiddleware checks the cache for a cached page on GET/HEAD
// requests. If a cache hit is found, it short-circuits the middleware chain
// and returns the cached response.
//
// In Django's middleware ordering, FetchFromCacheMiddleware is placed after
// UpdateCacheMiddleware so that it runs later on the request path (closer to
// the view) but earlier on the response path.
func FetchFromCacheMiddleware(next Handler) Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		if req.Method != http.MethodGet && req.Method != http.MethodHead {
			return next(req)
		}

		alias, _, keyPrefix := getCacheMiddlewareConfig()
		c := cache.Get(alias)

		// Phase 1: Retrieve Vary headers for this URL
		headerKey := generateMiddlewareHeaderKey(req, keyPrefix)
		var varyHeaders []string
		if hVal, err := c.Get(req.Context, headerKey); err == nil {
			switch hList := hVal.(type) {
			case []string:
				varyHeaders = append(varyHeaders, hList...)
			case []any:
				for _, h := range hList {
					if hs, ok := h.(string); ok {
						varyHeaders = append(varyHeaders, hs)
					}
				}
			}
		}

		// Phase 2: Try to fetch the cached page
		key := generateMiddlewareCacheKey(req, keyPrefix, varyHeaders)
		val, err := c.Get(req.Context, key)
		if err != nil {
			return next(req)
		}

		respMap, ok := val.(map[string]any)
		if !ok {
			return next(req)
		}

		return reconstructResponse(respMap)
	}
}

// UpdateCacheMiddleware caches the response for GET/HEAD requests after the
// view (and inner middlewares) have produced it. It should be placed first in
// the middleware chain so that it wraps everything else, allowing it to
// intercept the final response.
func UpdateCacheMiddleware(next Handler) Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		resp := next(req)

		if req.Method != http.MethodGet && req.Method != http.MethodHead {
			return resp
		}

		// Only cache successful responses
		if !isCacheable(resp) {
			return resp
		}

		alias, timeout, keyPrefix := getCacheMiddlewareConfig()
		c := cache.Get(alias)

		// Extract Vary headers from the response
		varyHeaders := getVaryHeaders(resp)
		if len(varyHeaders) == 0 {
			varyHeaders = []string{}
		}

		// Store the Vary header list
		headerKey := generateMiddlewareHeaderKey(req, keyPrefix)
		c.Set(req.Context, headerKey, varyHeaders, timeout)

		// Serialize and store the response
		key := generateMiddlewareCacheKey(req, keyPrefix, varyHeaders)
		cacheData := serializeResponse(resp)
		if cacheData != nil {
			c.Set(req.Context, key, cacheData, timeout)
		}

		return resp
	}
}

// isCacheable returns true for HttpResponse and JsonResponse with a 2xx or 3xx
// status code.
func isCacheable(resp godjangohttp.Response) bool {
	var statusCode int
	switch r := resp.(type) {
	case *godjangohttp.HttpResponse:
		statusCode = r.StatusCode
	case *godjangohttp.JsonResponse:
		statusCode = r.StatusCode
	default:
		return false
	}
	return statusCode >= 200 && statusCode < 400
}

// getVaryHeaders parses the Vary header from the response.
func getVaryHeaders(resp godjangohttp.Response) []string {
	var vary string
	switch r := resp.(type) {
	case *godjangohttp.HttpResponse:
		vary = r.Headers.Get("Vary")
	case *godjangohttp.JsonResponse:
		vary = r.Headers.Get("Vary")
	}
	if vary == "" {
		return nil
	}
	parts := strings.Split(vary, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

// serializeResponse reads the response body and headers into a cacheable map.
// The body is replaced with a fresh reader so the response remains usable.
func serializeResponse(resp godjangohttp.Response) map[string]any {
	switch r := resp.(type) {
	case *godjangohttp.HttpResponse:
		if r.Body == nil {
			return nil
		}
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return nil
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return map[string]any{
			"status":  r.StatusCode,
			"content": string(bodyBytes),
			"headers": convertHeadersToMap(r.Headers),
		}
	case *godjangohttp.JsonResponse:
		if r.Body == nil {
			return nil
		}
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return nil
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return map[string]any{
			"status":  r.StatusCode,
			"content": string(bodyBytes),
			"headers": convertHeadersToMap(r.Headers),
		}
	}
	return nil
}

// reconstructResponse rebuilds an HttpResponse from a cached map.
func reconstructResponse(respMap map[string]any) godjangohttp.Response {
	content, _ := respMap["content"].(string)
	statusFloat, _ := respMap["status"].(float64)
	status := int(statusFloat)
	if status == 0 {
		status = http.StatusOK
	}

	resp := godjangohttp.NewHttpResponse(content, status)
	if headers, ok := respMap["headers"].(map[string]any); ok {
		for k, vAny := range headers {
			if vList, ok := vAny.([]any); ok {
				for _, vStr := range vList {
					if s, ok := vStr.(string); ok {
						resp.Headers.Add(k, s)
					}
				}
			}
		}
	}
	return resp
}

// convertHeadersToMap converts an http.Header into a map[string]any suitable
// for cache serialization.
func convertHeadersToMap(h http.Header) map[string]any {
	res := make(map[string]any)
	for k, v := range h {
		list := make([]any, len(v))
		for i, s := range v {
			list[i] = s
		}
		res[k] = list
	}
	return res
}
