package middleware

import (
	"github.com/godjango/godjango/auth"
	godjangohttp "github.com/godjango/godjango/http"
)

// AuthenticationMiddleware attaches a User to the request.
// It depends on SessionMiddleware being executed first.
func AuthenticationMiddleware(next Handler) Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		req.User = auth.GetUserFromRequest(req)
		return next(req)
	}
}
