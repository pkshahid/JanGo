package middleware

import (
	"log"
	"runtime"

	godjangohttp "github.com/godjango/godjango/http"
	"github.com/godjango/godjango/http/views"
)

// PanicRecoveryMiddleware catches panics, logs the stack trace, and returns a 500 error page.
func PanicRecoveryMiddleware(next Handler) Handler {
	return func(req *godjangohttp.Request) (resp godjangohttp.Response) {
		defer func() {
			if r := recover(); r != nil {
				// Get stack trace
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				log.Printf("Panic recovered: %v\n%s\n", r, buf[:n])

				// Call ServerError view
				resp = views.ServerError(req)

				if err, ok := r.(error); ok {
					// Add internal error info to response if debugging
					if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
						if hr.StatusCode == 500 { // Just an extra check
							// We could append error text if DEBUG is enabled.
							// views.ServerError handles DEBUG logic. We don't overwrite its output
							// unless we want to inject the specific error.
							// Actually, let's just let ServerError handle it.
							_ = err
						}
					}
				}
			}
		}()

		return next(req)
	}
}
