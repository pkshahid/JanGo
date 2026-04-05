package middleware

import (
	godjangohttp "github.com/godjango/godjango/http"
)

// Handler represents the core HTTP request handler type.
type Handler func(req *godjangohttp.Request) godjangohttp.Response

// MiddlewareFunc represents a middleware which wraps a Handler.
type MiddlewareFunc func(next Handler) Handler

// Chain is a helper to compose middlewares into a single Handler.
type Chain struct {
	middlewares []MiddlewareFunc
}

// NewChain creates a new Chain with the given middlewares.
// The order matches Django's setting: the first middleware in the slice
// wraps the second, which wraps the third, and so on.
// This means the first middleware runs first on request, and last on response.
func NewChain(middlewares ...MiddlewareFunc) Chain {
	return Chain{middlewares: middlewares}
}

// Append adds middlewares to the end of the chain.
func (c Chain) Append(middlewares ...MiddlewareFunc) Chain {
	newMiddlewares := make([]MiddlewareFunc, 0, len(c.middlewares)+len(middlewares))
	newMiddlewares = append(newMiddlewares, c.middlewares...)
	newMiddlewares = append(newMiddlewares, middlewares...)
	return Chain{middlewares: newMiddlewares}
}

// Then applies the middleware chain to the given final handler.
func (c Chain) Then(handler Handler) Handler {
	// Apply middlewares from last to first to build the onion
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i](handler)
	}
	return handler
}

// Middleware is an interface representing Django's middleware hook model.
// Implementations can optionally implement ProcessRequest, ProcessView, ProcessResponse, ProcessException.
// This struct will be adapted to MiddlewareFunc by adapt()
type Middleware interface {
	ProcessRequest(req *godjangohttp.Request) godjangohttp.Response
	ProcessView(req *godjangohttp.Request, viewFunc Handler, args []string, kwargs map[string]any) godjangohttp.Response
	ProcessResponse(req *godjangohttp.Request, resp godjangohttp.Response) godjangohttp.Response
	ProcessException(req *godjangohttp.Request, err error) godjangohttp.Response
}

// AdaptMiddleware creates a MiddlewareFunc from a Middleware interface.
func AdaptMiddleware(m Middleware) MiddlewareFunc {
	return func(next Handler) Handler {
		return func(req *godjangohttp.Request) godjangohttp.Response {
			if resp := m.ProcessRequest(req); resp != nil {
				// If process_request returns a response, skip the rest of the chain
				return m.ProcessResponse(req, resp)
			}

			// ProcessView is typically called by the router before executing the view.
			// Since our middleware wraps the whole handler, simulating process_view here
			// requires extracting view info from ResolverMatch if available.
			// For simplicity in the generic chain, we just run next() and handle exceptions.

			// Execute next handler (which may panic/error depending on implementation)
			// In Go we usually use recover for exceptions.
			var resp godjangohttp.Response
			func() {
				defer func() {
					if r := recover(); r != nil {
						var err error
						if e, ok := r.(error); ok {
							err = e
						}
						// Process exception
						if excResp := m.ProcessException(req, err); excResp != nil {
							resp = excResp
						} else {
							panic(r) // re-panic if not handled
						}
					}
				}()
				resp = next(req)
			}()

			return m.ProcessResponse(req, resp)
		}
	}
}
