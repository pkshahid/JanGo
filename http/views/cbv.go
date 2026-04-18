package views

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	godjangohttp "github.com/pkshahid/JanGo/http"
)

// View is the interface for all Class-Based Views.
type View interface {
	Dispatch(req *godjangohttp.Request) godjangohttp.Response
}

// BaseView provides reflection-based HTTP method dispatching.
// Custom views should embed BaseView and implement Get, Post, etc.
type BaseView struct{}

// AsView creates a ViewFunc from a View interface.
func AsView(v View) ViewFunc {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		return v.Dispatch(req)
	}
}

// Dispatch routes the request to the method matching the HTTP method name (e.g. Get, Post).
func (v *BaseView) Dispatch(req *godjangohttp.Request, instance any) godjangohttp.Response {
	methodName := strings.Title(strings.ToLower(req.Method))

	val := reflect.ValueOf(instance)
	method := val.MethodByName(methodName)

	if !method.IsValid() {
		// Method not allowed
		return v.httpMethodNotAllowed(req, val)
	}

	// Call the method with req as argument
	in := []reflect.Value{reflect.ValueOf(req)}
	results := method.Call(in)

	if len(results) > 0 {
		if resp, ok := results[0].Interface().(godjangohttp.Response); ok {
			return resp
		}
	}

	return ServerError(req)
}

func (v *BaseView) httpMethodNotAllowed(req *godjangohttp.Request, instance reflect.Value) godjangohttp.Response {
	allowed := []string{}
	methods := []string{"Get", "Post", "Put", "Patch", "Delete", "Head", "Options", "Trace"}

	for _, m := range methods {
		if instance.MethodByName(m).IsValid() {
			allowed = append(allowed, strings.ToUpper(m))
		}
	}

	resp := godjangohttp.NewHttpResponse(fmt.Sprintf("Method %s Not Allowed", req.Method), http.StatusMethodNotAllowed)
	resp.Headers.Set("Allow", strings.Join(allowed, ", "))
	return resp
}

// TemplateView renders a given template with context.
type TemplateView struct {
	BaseView
	TemplateName string
	ContextData  map[string]any
}

func (v *TemplateView) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	return v.BaseView.Dispatch(req, v)
}

func (v *TemplateView) Get(req *godjangohttp.Request) godjangohttp.Response {
	ctx := v.GetContextData(req)
	return godjangohttp.Render(req, v.TemplateName, ctx)
}

func (v *TemplateView) GetContextData(req *godjangohttp.Request) map[string]any {
	if v.ContextData == nil {
		return make(map[string]any)
	}
	return v.ContextData
}

// RedirectView redirects to a specific URL.
type RedirectView struct {
	BaseView
	URL          string
	Permanent    bool
	PatternName  string
}

func (v *RedirectView) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	return v.BaseView.Dispatch(req, v)
}

func (v *RedirectView) Get(req *godjangohttp.Request) godjangohttp.Response {
	url := v.GetRedirectURL(req)
	if url != "" {
		return godjangohttp.NewRedirectResponse(url, v.Permanent)
	}
	return godjangohttp.HttpResponseNotFound("Not Found")
}

func (v *RedirectView) GetRedirectURL(req *godjangohttp.Request) string {
	if v.URL != "" {
		return v.URL
	}
	// Note: in a real implementation we would resolve v.PatternName via the URL Router.
	// Since views package shouldn't import urls package (to avoid circular deps),
	// we assume the reverse lookup logic happens elsewhere or is passed in.
	return ""
}
