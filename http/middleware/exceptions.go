package middleware

import (
	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/http/exceptions"
)

// ExceptionMiddleware transforms recognized error panics into standard HTTP responses.
// In GoDjango, since Handler returns Response, standard exceptions like Http404
// are either explicitly returned or panic'd. This catches them if they panic.
func ExceptionMiddleware(next Handler) Handler {
	return func(req *godjangohttp.Request) (resp godjangohttp.Response) {
		defer func() {
			if r := recover(); r != nil {
				resolver := &exceptions.Resolver{}
				switch r.(type) {
				case *exceptions.Http404, exceptions.Http404:
					resp = resolver.Resolve404(req)
				case *exceptions.PermissionDenied, exceptions.PermissionDenied:
					resp = resolver.Resolve403(req)
				case *exceptions.BadRequest, exceptions.BadRequest:
					resp = resolver.Resolve400(req)
				case *exceptions.SuspiciousOperation, exceptions.SuspiciousOperation:
					resp = resolver.Resolve400(req)
				default:
					// Re-panic for PanicRecoveryMiddleware to catch as a 500
					panic(r)
				}
			}
		}()
		return next(req)
	}
}
