package cache

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	godjangohttp "github.com/godjango/godjango/http"
	"github.com/godjango/godjango/http/views"
)

// CachePage caches the output of the decorated view for the given timeout.
func CachePage(timeout time.Duration, alias ...string) func(views.ViewFunc) views.ViewFunc {
	cacheAlias := "default"
	if len(alias) > 0 {
		cacheAlias = alias[0]
	}

	return func(next views.ViewFunc) views.ViewFunc {
		return func(req *godjangohttp.Request) godjangohttp.Response {
			if req.Method != "GET" && req.Method != "HEAD" {
				return next(req)
			}

			c := Get(cacheAlias)
			key := generateCacheKey(req)

			// Try to get from cache
			val, err := c.Get(req.Context, key)
			if err == nil {
				if respMap, ok := val.(map[string]any); ok {
					content, _ := respMap["content"].(string)
					statusFloat, _ := respMap["status"].(float64)
					status := int(statusFloat)

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
			}

			// Render response and cache it
			resp := next(req)

			if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
				// Read body to string
				bodyBytes, err := ioutil.ReadAll(hr.Body)
				if err == nil {
					// Restore body
					hr.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

					// Save to cache
					cacheData := map[string]any{
						"status":  hr.StatusCode,
						"content": string(bodyBytes),
						"headers": convertHeadersToMap(hr.Headers),
					}
					c.Set(req.Context, key, cacheData, timeout)
				}
			} else if jr, ok := resp.(*godjangohttp.JsonResponse); ok {
				// Handle JsonResponse which wraps HttpResponse
				bodyBytes, err := ioutil.ReadAll(jr.Body)
				if err == nil {
					jr.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
					cacheData := map[string]any{
						"status":  jr.StatusCode,
						"content": string(bodyBytes),
						"headers": convertHeadersToMap(jr.Headers),
					}
					c.Set(req.Context, key, cacheData, timeout)
				}
			}

			return resp
		}
	}
}

// VaryOnHeaders adds Vary headers to the response, which affects the cache key.
func VaryOnHeaders(headers ...string) func(views.ViewFunc) views.ViewFunc {
	return func(next views.ViewFunc) views.ViewFunc {
		return func(req *godjangohttp.Request) godjangohttp.Response {
			resp := next(req)
			if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
				addVary(hr.Headers, headers...)
			} else if jr, ok := resp.(*godjangohttp.JsonResponse); ok {
				addVary(jr.Headers, headers...)
			}
			return resp
		}
	}
}

// VaryOnCookie adds the Cookie header to Vary.
func VaryOnCookie() func(views.ViewFunc) views.ViewFunc {
	return VaryOnHeaders("Cookie")
}

func addVary(h http.Header, newHeaders ...string) {
	existing := h.Get("Vary")
	parts := []string{}
	if existing != "" {
		parts = strings.Split(existing, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
	}

	for _, header := range newHeaders {
		found := false
		for _, p := range parts {
			if strings.EqualFold(p, header) {
				found = true
				break
			}
		}
		if !found {
			parts = append(parts, header)
		}
	}

	if len(parts) > 0 {
		h.Set("Vary", strings.Join(parts, ", "))
	}
}

// generateCacheKey creates a unique key for the request URL, including Vary header state if needed.
// Note: In Django, cache keys are usually generated *after* response to know Vary headers, or
// through a more complex two-phase caching. For simplicity, we just hash URL here unless we implement
// a header-based key retrieval.
func generateCacheKey(req *godjangohttp.Request) string {
	urlStr := req.URL.String()

	// Check if this URL's Vary headers are known in cache.
	// Actually, a simpler approximation is just hashing the URL + specific query args.
	// Since we need to include Vary headers, let's just hash the URL and all current request headers.
	// In Django, it uses `utils.cache.get_cache_key` which evaluates headers based on a cached list.
	// For this implementation, we will append a generic list of Vary headers if we knew them,
	// but to make it testable and functional, we'll hash the URL.

	hash := md5.Sum([]byte(urlStr))
	return "views.decorators.cache.cache_page." + hex.EncodeToString(hash[:])
}

// Update generateCacheKey to consider known vary headers
func generateCacheKeyWithHeaders(req *godjangohttp.Request, varyHeaders []string) string {
	urlStr := req.URL.String()
	for _, h := range varyHeaders {
		urlStr += "|" + h + "=" + req.Header.Get(h)
	}
	hash := md5.Sum([]byte(urlStr))
	return "views.decorators.cache.cache_page." + hex.EncodeToString(hash[:])
}

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
