package admin

import (
	"bytes"
	"embed"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	godjangohttp "github.com/godjango/godjango/http"
)

//go:embed templates/* static/*
var adminFS embed.FS

var adminTemplates *template.Template

func init() {
	// Pre-parse the embedded html/templates on startup
	funcs := template.FuncMap{
		"lower": strings.ToLower,
	}
	adminTemplates = template.Must(template.New("").Funcs(funcs).ParseFS(adminFS, "templates/admin/*.html"))
}

// render is a helper function to render `html/template` blocks seamlessly.
// Unlike user templates which use the Django-style template engine,
// internal framework admin views often leverage standard Go templating to reduce cyclic dependencies
// or parsing overhead. This mimics Django's static admin view generation logic efficiently.
func render(req *godjangohttp.Request, tmplName string, data map[string]any) godjangohttp.Response {
	// Add globals
	if data == nil {
		data = make(map[string]any)
	}
	data["request"] = req
	data["site_header"] = DefaultAdminSite.SiteHeader
	data["site_title"] = DefaultAdminSite.SiteTitle

	// Clone the base template so we can define the required block dynamically based on tmplName
	tmpl, err := adminTemplates.Clone()
	if err != nil {
		return godjangohttp.NewHttpResponse("Admin clone error: "+err.Error(), http.StatusInternalServerError)
	}

	// Add a small wrapper that includes the target template as "content"
	wrapper := `{{template "` + tmplName + `" .}}`
	_, err = tmpl.New("content_wrapper").Parse(wrapper)
	if err != nil {
		return godjangohttp.NewHttpResponse("Admin wrapper parse error: "+err.Error(), http.StatusInternalServerError)
	}

	// Standard html/template rendering
	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "base.html", data)
	if err != nil {
		return godjangohttp.NewHttpResponse("Admin template render error: "+err.Error(), http.StatusInternalServerError)
	}

	return godjangohttp.NewHttpResponse(buf.String(), http.StatusOK)
}

// serveStatic serves the embedded CSS/JS files
func serveStatic(req *godjangohttp.Request, path string) godjangohttp.Response {
	content, err := adminFS.ReadFile(filepath.Join("static", path))
	if err != nil {
		return godjangohttp.HttpResponseNotFound("File not found")
	}

	resp := godjangohttp.NewHttpResponse(string(content), http.StatusOK)
	if filepath.Ext(path) == ".css" {
		resp.Headers.Set("Content-Type", "text/css")
	} else if filepath.Ext(path) == ".js" {
		resp.Headers.Set("Content-Type", "application/javascript")
	}
	return resp
}
