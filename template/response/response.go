package response

import (
	"net/http"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/template"
	"github.com/pkshahid/JanGo/template/context"
)

// Default processors used if none configured on the engine (for prototype simplicity)
var DefaultProcessors = []context.ProcessorFunc{
	context.RequestContextProcessor,
	context.AuthContextProcessor,
	context.MessagesContextProcessor,
	context.StaticContextProcessor,
	context.DebugContextProcessor,
	context.CsrfContextProcessor,
}

// RenderToString renders a template to a string.
// If req is provided, context processors are applied.
func RenderToString(engine *template.Engine, templateName string, baseContext map[string]any, req *godjangohttp.Request) (string, error) {
	tmpl, err := engine.GetTemplate(templateName)
	if err != nil {
		return "", err
	}

	var ctx *template.Context
	if req != nil {
		ctx = context.BuildRequestContext(req, DefaultProcessors, baseContext)
	} else {
		ctx = template.NewContext(baseContext)
	}

	return tmpl.Render(ctx)
}

// Render renders a template and returns an HttpResponse.
func Render(engine *template.Engine, req *godjangohttp.Request, templateName string, baseContext map[string]any) *godjangohttp.HttpResponse {
	content, err := RenderToString(engine, templateName, baseContext, req)
	if err != nil {
		return godjangohttp.NewHttpResponse("Template Render Error: "+err.Error(), http.StatusInternalServerError)
	}
	return godjangohttp.NewHttpResponse(content, http.StatusOK)
}

// TemplateResponse is a lazy HttpResponse that renders when written.
type TemplateResponse struct {
	Engine       *template.Engine
	Request      *godjangohttp.Request
	TemplateName string
	ContextData  map[string]any

	// Underlying HttpResponse properties
	StatusCode int
	Headers    http.Header

	rendered bool
	content  []byte
}

func NewTemplateResponse(engine *template.Engine, req *godjangohttp.Request, templateName string, contextData map[string]any) *TemplateResponse {
	return &TemplateResponse{
		Engine:       engine,
		Request:      req,
		TemplateName: templateName,
		ContextData:  contextData,
		StatusCode:   http.StatusOK,
		Headers:      make(http.Header),
	}
}

// Resolve renders the template into the content buffer if not already rendered.
func (r *TemplateResponse) Resolve() {
	if r.rendered {
		return
	}
	content, err := RenderToString(r.Engine, r.TemplateName, r.ContextData, r.Request)
	if err != nil {
		r.content = []byte("Template Render Error: " + err.Error())
		r.StatusCode = http.StatusInternalServerError
	} else {
		r.content = []byte(content)
	}
	r.rendered = true
}

// Write implements the godjangohttp.Response interface.
func (r *TemplateResponse) Write(w http.ResponseWriter) {
	r.Resolve()

	for k, v := range r.Headers {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}
	w.WriteHeader(r.StatusCode)
	w.Write(r.content)
}

// SimpleTemplateResponse is a lazy HttpResponse without request context execution.
type SimpleTemplateResponse struct {
	Engine       *template.Engine
	TemplateName string
	ContextData  map[string]any

	StatusCode int
	Headers    http.Header

	rendered bool
	content  []byte
}

func NewSimpleTemplateResponse(engine *template.Engine, templateName string, contextData map[string]any) *SimpleTemplateResponse {
	return &SimpleTemplateResponse{
		Engine:       engine,
		TemplateName: templateName,
		ContextData:  contextData,
		StatusCode:   http.StatusOK,
		Headers:      make(http.Header),
	}
}

func (r *SimpleTemplateResponse) Resolve() {
	if r.rendered {
		return
	}
	content, err := RenderToString(r.Engine, r.TemplateName, r.ContextData, nil)
	if err != nil {
		r.content = []byte("Template Render Error: " + err.Error())
		r.StatusCode = http.StatusInternalServerError
	} else {
		r.content = []byte(content)
	}
	r.rendered = true
}

func (r *SimpleTemplateResponse) Write(w http.ResponseWriter) {
	r.Resolve()

	for k, v := range r.Headers {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}
	w.WriteHeader(r.StatusCode)
	w.Write(r.content)
}
