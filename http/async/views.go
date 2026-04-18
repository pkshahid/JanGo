package async

import (
	godjangohttp "github.com/godjango/godjango/http"
	"github.com/godjango/godjango/http/middleware"
)

// AsyncView represents a view that returns a channel of Response.
// This allows the request goroutine to block until the view resolves.
type AsyncView func(req *godjangohttp.Request) <-chan godjangohttp.Response

// AsyncHandler wraps an AsyncView to make it a standard Handler.
// It waits on the channel, respecting the request's context cancellation.
func AsyncHandler(view AsyncView) middleware.Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		ch := view(req)

		select {
		case <-req.Context.Done():
			// Client disconnected or timed out
			return godjangohttp.HttpResponseBadRequest("Request Cancelled")
		case resp := <-ch:
			return resp
		}
	}
}
