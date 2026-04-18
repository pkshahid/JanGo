package exceptions

import (
	"net/http"

	"github.com/godjango/godjango/core/settings"
	godjangohttp "github.com/godjango/godjango/http"
	"github.com/godjango/godjango/http/urls"
)

// Default error handlers for production.

func Handler400(req *godjangohttp.Request) godjangohttp.Response {
	return godjangohttp.NewHttpResponse("<h1>Bad Request (400)</h1>", http.StatusBadRequest)
}

func Handler403(req *godjangohttp.Request) godjangohttp.Response {
	return godjangohttp.NewHttpResponse("<h1>Forbidden (403)</h1>", http.StatusForbidden)
}

func Handler404(req *godjangohttp.Request) godjangohttp.Response {
	return godjangohttp.NewHttpResponse("<h1>Page not found (404)</h1>", http.StatusNotFound)
}

func Handler500(req *godjangohttp.Request) godjangohttp.Response {
	return godjangohttp.NewHttpResponse("<h1>Server Error (500)</h1>", http.StatusInternalServerError)
}

// Resolver is a utility to resolve which error handler to call.
type Resolver struct{}

func (r *Resolver) resolveCustom(handlerName string, req *godjangohttp.Request) godjangohttp.Response {
	router := urls.GetGlobalRouter()
	if router != nil {
		// A full implementation would look up the view function by its string name via an app registry.
		// Since we don't have a string-to-func registry in the router directly, we return nil
		// to fallback to the default.
	}
	return nil
}

func (r *Resolver) Resolve404(req *godjangohttp.Request) godjangohttp.Response {
	s := settings.Get()
	if s.HANDLER_404 != "" {
		if resp := r.resolveCustom(s.HANDLER_404, req); resp != nil {
			return resp
		}
	}
	return Handler404(req)
}

func (r *Resolver) Resolve500(req *godjangohttp.Request) godjangohttp.Response {
	s := settings.Get()
	if s.HANDLER_500 != "" {
		if resp := r.resolveCustom(s.HANDLER_500, req); resp != nil {
			return resp
		}
	}
	return Handler500(req)
}

func (r *Resolver) Resolve403(req *godjangohttp.Request) godjangohttp.Response {
	s := settings.Get()
	if s.HANDLER_403 != "" {
		if resp := r.resolveCustom(s.HANDLER_403, req); resp != nil {
			return resp
		}
	}
	return Handler403(req)
}

func (r *Resolver) Resolve400(req *godjangohttp.Request) godjangohttp.Response {
	s := settings.Get()
	if s.HANDLER_400 != "" {
		if resp := r.resolveCustom(s.HANDLER_400, req); resp != nil {
			return resp
		}
	}
	return Handler400(req)
}
