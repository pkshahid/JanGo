package asgi

import (
	"fmt"
	"net/http"
	"sync"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/http/middleware"
	"github.com/pkshahid/JanGo/http/ws"
)

// Router interface defines the contract for looking up views by path.
type Router interface {
	Resolve(path string) (any, map[string]string)
}

var globalRouter Router

// SetGlobalRouter sets the router for the handler.
func SetGlobalRouter(r Router) {
	globalRouter = r
}

// AsgiHandler acts as the main entry point for ASGI/async compliant requests.
// It fully supports net/http (since Go is inherently async) and handles WS upgrades.
type AsgiHandler struct {
	MiddlewareChain middleware.Chain
	reqPool         sync.Pool
}

// NewAsgiHandler creates a new AsgiHandler.
func NewAsgiHandler(chain middleware.Chain) *AsgiHandler {
	return &AsgiHandler{
		MiddlewareChain: chain,
		reqPool: sync.Pool{
			New: func() any {
				return &godjangohttp.Request{}
			},
		},
	}
}

func (h *AsgiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Construct the GoDjango request
	req := godjangohttp.NewRequest(r)

	// Route the request to check if it's a websocket upgrade
	if globalRouter != nil {
		view, _ := globalRouter.Resolve(req.Path)
		// If it's a WebSocket view, handle the upgrade and return immediately
		if wsView, ok := view.(ws.WebSocketView); ok {
			ws.ServeWebSocket(wsView, req, w)
			return
		}
	}

	// It's a standard HTTP request, run through middleware
	var handler middleware.Handler = func(req *godjangohttp.Request) godjangohttp.Response {
		if globalRouter == nil {
			return godjangohttp.HttpResponseNotFound("Router not configured")
		}
		view, _ := globalRouter.Resolve(req.Path)
		if httpView, ok := view.(middleware.Handler); ok {
			return httpView(req)
		}
		return godjangohttp.HttpResponseNotFound("Not Found")
	}

	finalHandler := h.MiddlewareChain.Then(handler)

	// Since we support HTTP/2 Push (if the underlying ResponseWriter supports it),
	// we could optionally check for Push here and inject it into the context or request.
	if pusher, ok := w.(http.Pusher); ok {
		// Mock storing pusher, though standard `net/http` handles this.
		// e.g. req.Context = context.WithValue(req.Context, "pusher", pusher)
		_ = pusher
	}

	// Execute
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("panic: %v\n", err)
			http.Error(w, "Internal Server Error", 500)
		}
	}()

	resp := finalHandler(req)
	if resp != nil {
		resp.Write(w)
	} else {
		http.Error(w, "Internal Server Error", 500)
	}
}
