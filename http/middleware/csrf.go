package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/godjango/godjango/core/settings"
	godjangohttp "github.com/godjango/godjango/http"
)

type csrfContextKey string
const csrfExemptKey = csrfContextKey("csrf_exempt")

// CsrfExempt marks a handler as exempt from CSRF validation.
func CsrfExempt(handler Handler) Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		ctx := context.WithValue(req.Context, csrfExemptKey, true)
		req.Request = req.Request.WithContext(ctx)
		req.Context = ctx
		return handler(req)
	}
}

// CsrfViewMiddleware implements double-submit cookie CSRF protection.
func CsrfViewMiddleware(next Handler) Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		s := settings.Get()
		cookieName := s.CSRF_COOKIE_NAME
		if cookieName == "" {
			cookieName = "csrftoken"
		}

		// Extract or generate token
		csrfCookie := req.COOKIES[cookieName]
		if csrfCookie == "" {
			bytes := make([]byte, 32)
			rand.Read(bytes)
			csrfCookie = hex.EncodeToString(bytes)
		}

		// Attach token to request so templates can use it
		req.META["CSRF_TOKEN"] = csrfCookie

		// Check if exempt
		if exempt, _ := req.Context.Value(csrfExemptKey).(bool); exempt {
			resp := next(req)
			return setCsrfCookie(resp, cookieName, csrfCookie)
		}

		// Safe methods don't need CSRF
		if req.Method == http.MethodGet || req.Method == http.MethodHead || req.Method == http.MethodOptions || req.Method == http.MethodTrace {
			resp := next(req)
			return setCsrfCookie(resp, cookieName, csrfCookie)
		}

		// Validate token
		reqToken := req.POST.Get("csrfmiddlewaretoken")
		if reqToken == "" {
			reqToken = req.Header.Get("X-CSRFToken")
		}

		if reqToken == "" || reqToken != csrfCookie {
			return godjangohttp.HttpResponseForbidden("CSRF token missing or incorrect.")
		}

		resp := next(req)
		return setCsrfCookie(resp, cookieName, csrfCookie)
	}
}

func setCsrfCookie(resp godjangohttp.Response, name, value string) godjangohttp.Response {
	if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
		cookie := &http.Cookie{
			Name:     name,
			Value:    value,
			Path:     "/",
			HttpOnly: false, // Must be accessible to JS for X-CSRFToken
			SameSite: http.SameSiteLaxMode,
		}
		hr.Headers.Add("Set-Cookie", cookie.String())
	}
	return resp
}
