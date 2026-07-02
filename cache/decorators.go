package cache

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/http/views"
)

// CachePage caches the output of the decorated view for the given timeout.
// getVaryHeaders parses the Vary header from a response, if present.
// Since we don't know the response's Vary headers until we render it, a robust framework does two-phase caching.
// For this simplified implementation, we'll assume any `Vary` headers appended to `req.Header` or we just rely on standard ones.
// A common trick is caching the `Vary` header names themselves under a separate key.
// Let's implement two-phase caching!

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

			// Phase 1: Retrieve Vary headers from cache
			headerKey := "views.decorators.cache.cache_header." + func() string { hash := md5.Sum([]byte(req.URL.String())); return hex.EncodeToString(hash[:]) }()
			varyHeaders := []string{}

			if hVal, err := c.Get(req.Context, headerKey); err == nil {
				if hList, ok := hVal.([]any); ok {
					for _, h := range hList {
						if hs, ok := h.(string); ok {
							varyHeaders = append(varyHeaders, hs)
						}
					}
				}
			}

			key := generateCacheKeyWithHeaders(req, varyHeaders)

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

			// Phase 2: Save Vary headers and content to cache
			var currentVary string
			if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
				currentVary = hr.Headers.Get("Vary")
				bodyBytes, err := io.ReadAll(hr.Body)
				hr.Body.Close()
				if err == nil {
					hr.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					cacheData := map[string]any{
						"status":  hr.StatusCode,
						"content": string(bodyBytes),
						"headers": convertHeadersToMap(hr.Headers),
					}

					// Re-evaluate key with new Vary headers if they changed
					newVaryHeaders := parseVaryHeader(currentVary)
					newKey := generateCacheKeyWithHeaders(req, newVaryHeaders)

					c.Set(req.Context, headerKey, newVaryHeaders, timeout)
					c.Set(req.Context, newKey, cacheData, timeout)
				}
			} else if jr, ok := resp.(*godjangohttp.JsonResponse); ok {
				currentVary = jr.Headers.Get("Vary")
				bodyBytes, err := io.ReadAll(jr.Body)
				jr.Body.Close()
				if err == nil {
					jr.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					cacheData := map[string]any{
						"status":  jr.StatusCode,
						"content": string(bodyBytes),
						"headers": convertHeadersToMap(jr.Headers),
					}

					newVaryHeaders := parseVaryHeader(currentVary)
					newKey := generateCacheKeyWithHeaders(req, newVaryHeaders)

					c.Set(req.Context, headerKey, newVaryHeaders, timeout)
					c.Set(req.Context, newKey, cacheData, timeout)
				}
			}

			return resp
		}
	}
}

func parseVaryHeader(vary string) []string {
	if vary == "" {
		return nil
	}
	parts := strings.Split(vary, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
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

// generateCacheKeyWithHeaders creates a unique key considering known vary headers.
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
