package views

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/godjango/godjango/core/settings"
	godjangohttp "github.com/godjango/godjango/http"
)

// LoginRequired ensures the user is authenticated, otherwise redirects to the login page.
func LoginRequired(view ViewFunc) ViewFunc {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		if req.User == nil || !req.User.IsAuthenticated() {
			s := settings.Get()
			loginURL := s.LOGIN_URL
			if loginURL == "" {
				loginURL = "/login/" // Fallback default
			}
			// In Django, it adds ?next=...
			next := req.URL.RequestURI()
			redirectURL := fmt.Sprintf("%s?next=%s", loginURL, next)
			return godjangohttp.NewRedirectResponse(redirectURL, false)
		}
		return view(req)
	}
}

// PermissionRequired ensures the user has a specific permission, otherwise returns 403.
func PermissionRequired(perm string, view ViewFunc) ViewFunc {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		if req.User == nil || !req.User.IsAuthenticated() {
			s := settings.Get()
			loginURL := s.LOGIN_URL
			if loginURL == "" {
				loginURL = "/login/"
			}
			return godjangohttp.NewRedirectResponse(loginURL+"?next="+req.URL.RequestURI(), false)
		}
		if !req.User.HasPerm(perm) {
			return PermissionDenied(req, nil)
		}
		return view(req)
	}
}

// RequireHTTPMethods allows only specific HTTP methods.
func RequireHTTPMethods(methods []string) func(ViewFunc) ViewFunc {
	return func(view ViewFunc) ViewFunc {
		return func(req *godjangohttp.Request) godjangohttp.Response {
			for _, m := range methods {
				if strings.EqualFold(req.Method, m) {
					return view(req)
				}
			}
			resp := godjangohttp.NewHttpResponse("Method Not Allowed", http.StatusMethodNotAllowed)
			resp.Headers.Set("Allow", strings.Join(methods, ", "))
			return resp
		}
	}
}

// RequireGET allows only GET requests.
func RequireGET(view ViewFunc) ViewFunc {
	return RequireHTTPMethods([]string{http.MethodGet})(view)
}

// RequirePOST allows only POST requests.
func RequirePOST(view ViewFunc) ViewFunc {
	return RequireHTTPMethods([]string{http.MethodPost})(view)
}

// RequireSafe allows only GET and HEAD requests.
func RequireSafe(view ViewFunc) ViewFunc {
	return RequireHTTPMethods([]string{http.MethodGet, http.MethodHead})(view)
}

// CacheControl sets Cache-Control headers via a map of directives.
func CacheControl(kwargs map[string]any) func(ViewFunc) ViewFunc {
	return func(view ViewFunc) ViewFunc {
		return func(req *godjangohttp.Request) godjangohttp.Response {
			resp := view(req)
			if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
				var directives []string

				for k, v := range kwargs {
					if b, isBool := v.(bool); isBool {
						if b {
							directives = append(directives, k)
						}
					} else {
						directives = append(directives, fmt.Sprintf("%s=%v", k, v))
					}
				}

				if len(directives) > 0 {
					existing := hr.Headers.Get("Cache-Control")
					if existing != "" {
						hr.Headers.Set("Cache-Control", existing+", "+strings.Join(directives, ", "))
					} else {
						hr.Headers.Set("Cache-Control", strings.Join(directives, ", "))
					}
				}
			}
			return resp
		}
	}
}

// NeverCache adds headers to prevent caching.
func NeverCache(view ViewFunc) ViewFunc {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		resp := view(req)
		if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
			hr.Headers.Set("Cache-Control", "max-age=0, no-cache, no-store, must-revalidate")
			hr.Headers.Set("Pragma", "no-cache")
			hr.Headers.Set("Expires", "0") // Set Expires to past or 0
		}
		return resp
	}
}
