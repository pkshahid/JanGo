package admin

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/pkshahid/JanGo/forms"
	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/http/urls"
)

func (s *AdminSite) index(req *godjangohttp.Request) godjangohttp.Response {
	ctx := map[string]any{
		"title":    s.IndexTitle,
		"app_dict": s.appDict,
	}
	return render(req, "index.html", ctx)
}

func (s *AdminSite) appIndex(req *godjangohttp.Request) godjangohttp.Response {
	// A real framework filters appDict by req.ResolverMatch.Kwargs["app_label"]
	return godjangohttp.NewRedirectResponse("/admin/", false)
}

func (s *AdminSite) changelistView(req *godjangohttp.Request) godjangohttp.Response {
	rm := req.ResolverMatch.(*urls.ResolverMatch)
	app := rm.Kwargs["app_label"].(string)
	model := rm.Kwargs["model_name"].(string)

	admin, err := s.getModel(app, model)
	if err != nil {
		return godjangohttp.HttpResponseNotFound("Model not found")
	}

	// Resolve filters: convert ListFilter entries into ListFilterer instances
	filters := ResolveFilters(admin)
	filterSpecs := make([]FilterSpec, 0, len(filters))
	for _, f := range filters {
		choices := f.Choices(req)
		filterSpecs = append(filterSpecs, FilterSpec{
			Title:     f.Title(),
			Choices:   choices,
			HasActive: hasActiveChoice(choices),
		})
	}

	// 1. Query Data
	// In Go, since we use generics, instantiating a QuerySet generically at runtime via reflection
	// is complex. We will mock the fetch using `RawQuerySet` for execution abstraction,
	// or rely on a `AdminQuery` hook.
	// For this prototype, we'll return an empty list or mock data unless we use reflection accurately.
	// Since we mock DB, we'll return a mock count and empty table.

	ctx := map[string]any{
		"title":        fmt.Sprintf("Select %s to change", admin.ModelInfo.Name),
		"model_admin":  admin,
		"columns":      admin.ListDisplay,
		"results":      []map[string]any{}, // Mock empty rows
		"count":        0,
		"actions":      len(admin.Actions) > 0,
		"has_filters":  len(filterSpecs) > 0,
		"filter_specs": filterSpecs,
	}

	// 2. Handle Actions
	if req.Method == http.MethodPost {
		actionName := req.POST.Get("action")
		if actionName != "" {
			// Find and execute action
			// e.g. deleteSelectedAction
			return godjangohttp.NewRedirectResponse(req.URL.RequestURI(), false)
		}
	}

	return render(req, "change_list.html", ctx)
}

func (s *AdminSite) addView(req *godjangohttp.Request) godjangohttp.Response {
	rm := req.ResolverMatch.(*urls.ResolverMatch)
	app := rm.Kwargs["app_label"].(string)
	model := rm.Kwargs["model_name"].(string)

	admin, err := s.getModel(app, model)
	if err != nil {
		return godjangohttp.HttpResponseNotFound("Model not found")
	}

	// Create new instance of model
	instance := reflect.New(admin.ModelInfo.Type).Interface()

	// Handle form
	mf, err := forms.NewModelForm(instance, admin.Fields, admin.ReadonlyFields)
	if err != nil {
		return godjangohttp.NewHttpResponse(err.Error(), 500)
	}

	if req.Method == http.MethodPost {

		// Convert url.Values to map[string]any for form binding
		postData := make(map[string]any)
		for k, v := range req.POST {
			if len(v) > 0 {
				postData[k] = v[0]
			}
		}

		mf.Bind(postData, req.FILES)
		if mf.IsValid() {
			saved, _ := mf.Save(true)
			admin.SaveModel(req, saved, mf, false)

			// Log addition
			// qs.Create(&LogEntry{...})

			return godjangohttp.NewRedirectResponse("../", false)
		}
	}

	ctx := map[string]any{
		"title":       fmt.Sprintf("Add %s", admin.ModelInfo.Name),
		"model_admin": admin,
		"form_html":   mf.Render(),
		"errors":      mf.NonFieldErrors(),
		"change":      false,
	}

	return render(req, "change_form.html", ctx)
}

func (s *AdminSite) changeView(req *godjangohttp.Request) godjangohttp.Response {
	rm := req.ResolverMatch.(*urls.ResolverMatch)
	app := rm.Kwargs["app_label"].(string)
	model := rm.Kwargs["model_name"].(string)
	idStr := rm.Kwargs["object_id"].(string)

	admin, err := s.getModel(app, model)
	if err != nil {
		return godjangohttp.HttpResponseNotFound("Model not found")
	}

	// Fetch instance
	instance := reflect.New(admin.ModelInfo.Type).Interface()
	// Mock fetch obj by idStr
	_ = idStr

	mf, err := forms.NewModelForm(instance, admin.Fields, admin.ReadonlyFields)
	if err != nil {
		return godjangohttp.NewHttpResponse(err.Error(), 500)
	}

	if req.Method == http.MethodPost {
		// Convert url.Values to map[string]any for form binding
		postData := make(map[string]any)
		for k, v := range req.POST {
			if len(v) > 0 {
				postData[k] = v[0]
			}
		}

		mf.Bind(postData, req.FILES)
		if mf.IsValid() {
			saved, _ := mf.Save(true)
			admin.SaveModel(req, saved, mf, true)
			return godjangohttp.NewRedirectResponse("../../", false)
		}
	}

	ctx := map[string]any{
		"title":            fmt.Sprintf("Change %s", admin.ModelInfo.Name),
		"model_admin":      admin,
		"form_html":        mf.Render(),
		"errors":           mf.NonFieldErrors(),
		"change":           true,
		"view_on_site_url": admin.ViewOnSiteURL(instance),
	}

	return render(req, "change_form.html", ctx)
}

func (s *AdminSite) deleteView(req *godjangohttp.Request) godjangohttp.Response {
	rm := req.ResolverMatch.(*urls.ResolverMatch)
	app := rm.Kwargs["app_label"].(string)
	model := rm.Kwargs["model_name"].(string)

	admin, err := s.getModel(app, model)
	if err != nil {
		return godjangohttp.HttpResponseNotFound("Model not found")
	}

	if req.Method == http.MethodPost && req.POST.Get("post") == "yes" {
		instance := reflect.New(admin.ModelInfo.Type).Interface()
		admin.DeleteModel(req, instance)
		return godjangohttp.NewRedirectResponse("../../", false)
	}

	ctx := map[string]any{
		"title":       fmt.Sprintf("Delete %s", admin.ModelInfo.Name),
		"model_admin": admin,
		"object_name": "Object #ID",
	}

	return render(req, "delete_confirmation.html", ctx)
}

func (s *AdminSite) historyView(req *godjangohttp.Request) godjangohttp.Response {
	return godjangohttp.NewHttpResponse("History view stub", 200)
}

// ServeStatic wraps static file serving for admin
func (s *AdminSite) ServeStatic(req *godjangohttp.Request) godjangohttp.Response {
	rm := req.ResolverMatch.(*urls.ResolverMatch)
	path := rm.Kwargs["path"].(string)
	return serveStatic(req, path)
}
