package http

import (
	"bytes"
	"fmt"
	"net/http"
	"text/template"
)

// Dummy global template registry for rendering shortcuts until the full template engine is integrated.
// In a full implementation, this would invoke the GoDjango template engine.
var templates *template.Template

func init() {
	templates = template.New("base")
}

// Render renders a template with the given context and returns an HttpResponse.
func Render(req *Request, tmplName string, ctx map[string]any) *HttpResponse {
	// Add request to the context
	if ctx == nil {
		ctx = make(map[string]any)
	}
	ctx["request"] = req

	var buf bytes.Buffer
	// In the future this will use the framework's configured template backend
	// For now we use the stdlib html/template or text/template.
	// Since templates are likely loaded by the engine elsewhere, we just attempt to execute.
	err := templates.ExecuteTemplate(&buf, tmplName, ctx)
	if err != nil {
		// If template engine is not fully configured, return the error as a 500 response
		return NewHttpResponse(fmt.Sprintf("Template Error: %v", err), http.StatusInternalServerError)
	}

	return NewHttpResponse(buf.String(), http.StatusOK)
}

// Redirect returns a RedirectResponse to the specified URL.
func Redirect(to string, permanent bool) *RedirectResponse {
	return NewRedirectResponse(to, permanent)
}

// Get404 returns a generic 404 HttpResponse.
func Get404(req *Request) *HttpResponse {
	return HttpResponseNotFound("Not Found")
}
