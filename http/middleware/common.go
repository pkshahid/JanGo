package middleware

import (
	"net/http"
	"strings"

	"github.com/pkshahid/JanGo/core/settings"
	godjangohttp "github.com/pkshahid/JanGo/http"
)

// CommonMiddleware handles append slash, content-length, and conditional GETs.
func CommonMiddleware(next Handler) Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		s := settings.Get()

		// Append Slash Redirect
		if s.APPEND_SLASH && req.Method != http.MethodPost && !strings.HasSuffix(req.Path, "/") {
			// Basic append slash implementation.
			// Ideally, we'd check if the un-slashed path matched a route first.
			// For simplicity, we just append if APPEND_SLASH is true and no trailing slash exists.

			// A correct Django implementation attempts a route match with slash,
			// if unslashed fails. We will assume the basic behavior for now.

			redirectURL := req.Path + "/"
			if req.URL.RawQuery != "" {
				redirectURL += "?" + req.URL.RawQuery
			}
			return godjangohttp.NewRedirectResponse(redirectURL, true)
		}

		resp := next(req)

		// Set Content-Length if body is available and not streaming
		if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
			// This is simplistic since Body is an io.ReadCloser and we'd need to buffer to get length.
			// We skip auto content-length calculation if body is dynamic, or rely on net/http to set it,
			// but we can set it if it's already provided or if we replace the body with a buffered one.

			// ETag / Last-Modified logic (Conditional GET)
			if req.Method == http.MethodGet {
				etag := hr.Headers.Get("ETag")
				if etag != "" && req.Header.Get("If-None-Match") == etag {
					hr.StatusCode = http.StatusNotModified
					hr.Body = nil
				}
			}
		}

		return resp
	}
}
