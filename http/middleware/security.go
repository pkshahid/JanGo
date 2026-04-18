package middleware

import (
	"fmt"
	"strings"

	"github.com/pkshahid/JanGo/core/settings"
	godjangohttp "github.com/pkshahid/JanGo/http"
)

// SecurityMiddleware adds security headers and redirects to HTTPS.
func SecurityMiddleware(next Handler) Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		s := settings.Get()

		// HTTPS Redirect
		if s.SECURE_SSL_REDIRECT && req.Header.Get("X-Forwarded-Proto") != "https" && req.URL.Scheme != "https" {
			host := req.Host
			redirectURL := fmt.Sprintf("https://%s%s", host, req.URL.RequestURI())
			return godjangohttp.NewRedirectResponse(redirectURL, true)
		}

		resp := next(req)

		// HSTS
		if s.SECURE_HSTS_SECONDS > 0 {
			hsts := fmt.Sprintf("max-age=%d", s.SECURE_HSTS_SECONDS)
			if s.SECURE_HSTS_INCLUDE_SUBDOMAINS {
				hsts += "; includeSubDomains"
			}
			if s.SECURE_HSTS_PRELOAD {
				hsts += "; preload"
			}
			if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
				hr.Headers.Set("Strict-Transport-Security", hsts)
			}
		}

		// Content-Type Options
		if s.SECURE_CONTENT_TYPE_NOSNIFF {
			if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
				hr.Headers.Set("X-Content-Type-Options", "nosniff")
			}
		}

		// X-Frame-Options
		if s.X_FRAME_OPTIONS != "" {
			if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
				hr.Headers.Set("X-Frame-Options", strings.ToUpper(s.X_FRAME_OPTIONS))
			}
		}

		// Referrer Policy
		if s.SECURE_REFERRER_POLICY != "" {
			if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
				hr.Headers.Set("Referrer-Policy", s.SECURE_REFERRER_POLICY)
			}
		}

		return resp
	}
}
