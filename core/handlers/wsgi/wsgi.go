package wsgi

import (
	"net/http"
	"sync"

	godjangohttp "github.com/godjango/godjango/http"
	"github.com/godjango/godjango/http/middleware"
	"github.com/godjango/godjango/http/urls"
	"github.com/godjango/godjango/http/views"
)

// WSGIHandler implements http.Handler and acts as the entry point.
type WSGIHandler struct {
	middlewareChain middleware.Chain
	requestPool     sync.Pool
	router          *urls.Router
}

// NewWSGIHandler creates a new WSGIHandler.
// In a full implementation, it would construct the middleware chain based on settings.MIDDLEWARE.
func NewWSGIHandler(router *urls.Router, middlewares ...middleware.MiddlewareFunc) *WSGIHandler {
	return &WSGIHandler{
		middlewareChain: middleware.NewChain(middlewares...),
		router:          router,
		requestPool: sync.Pool{
			New: func() any {
				return &godjangohttp.Request{}
			},
		},
	}
}

// ServeHTTP handles the incoming HTTP request.
func (h *WSGIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Re-use request objects
	req := h.requestPool.Get().(*godjangohttp.Request)
	defer h.requestPool.Put(req)

	// Initialize the Request by building a new one and copying to the pooled instance
	newReq := godjangohttp.NewRequest(r)
	*req = *newReq

	// Create the final handler that routes the request
	finalHandler := func(request *godjangohttp.Request) godjangohttp.Response {
		match, err := h.router.Match(request.Path)
		if err != nil {
			return views.PageNotFound(request, nil)
		}

		request.ResolverMatch = match

		// Merge URL kwargs into the request or pass them to the view.
		// Wait, ViewFunc takes (*Request). So we attach them to the Request or ResolverMatch.
		// They are already in match.Kwargs.

		return match.Func(request)
	}

	// Apply middleware chain
	chainedHandler := h.middlewareChain.Then(finalHandler)

	// Execute
	resp := chainedHandler(req)

	// Write response
	if resp != nil {
		resp.Write(w)
	} else {
		// Fallback
		http.Error(w, "Empty response", http.StatusInternalServerError)
	}
}
