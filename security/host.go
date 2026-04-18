package security

import (
	"strings"

	"github.com/pkshahid/JanGo/core/settings"
	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/http/middleware"
)

// AllowedHostsMiddleware checks the Host header against ALLOWED_HOSTS.
type AllowedHostsMiddleware struct{}

// NewAllowedHostsMiddleware creates a new AllowedHostsMiddleware.
func NewAllowedHostsMiddleware() *AllowedHostsMiddleware {
	return &AllowedHostsMiddleware{}
}

func (m *AllowedHostsMiddleware) Process(next middleware.Handler) middleware.Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		s := settings.Get()

		// If DEBUG is true and ALLOWED_HOSTS is empty, allow localhost.
		// Usually Django allows localhost implicitly if empty in DEBUG.
		allowed := s.ALLOWED_HOSTS
		if s.DEBUG && len(allowed) == 0 {
			allowed = []string{"127.0.0.1", "localhost", "[::1]"}
		}

		host := req.Host
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}

		isValid := false
		for _, allowedHost := range allowed {
			if allowedHost == "*" {
				isValid = true
				break
			}
			if strings.HasPrefix(allowedHost, ".") {
				// Wildcard subdomains e.g. .example.com
				if strings.HasSuffix(host, allowedHost) || host == allowedHost[1:] {
					isValid = true
					break
				}
			} else if host == allowedHost {
				isValid = true
				break
			}
		}

		if !isValid {
			return godjangohttp.HttpResponseBadRequest("Invalid HTTP_HOST header")
		}

		return next(req)
	}
}
