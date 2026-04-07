package admin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/godjango/godjango/auth"
	godjangohttp "github.com/godjango/godjango/http"
	"github.com/godjango/godjango/http/urls"
	"github.com/godjango/godjango/orm"
)

// AdminSite encapsulates an instance of the admin application.
type AdminSite struct {
	Name             string
	SiteHeader       string
	SiteTitle        string
	IndexTitle       string
	_registry        map[string]*ModelAdmin
	appDict          map[string][]*ModelAdmin
}

// NewAdminSite creates a new AdminSite.
func NewAdminSite(name string) *AdminSite {
	if name == "" {
		name = "admin"
	}
	return &AdminSite{
		Name:       name,
		SiteHeader: "GoDjango administration",
		SiteTitle:  "GoDjango site admin",
		IndexTitle: "Site administration",
		_registry:  make(map[string]*ModelAdmin),
		appDict:    make(map[string][]*ModelAdmin),
	}
}

// Register maps a model to a ModelAdmin instance on the site.
func (s *AdminSite) Register(model any, adminConfig *ModelAdmin) error {
	info, err := orm.GetModelInfo(model)
	if err != nil {
		return err
	}

	key := strings.ToLower(info.Name)

	if _, exists := s._registry[key]; exists {
		return fmt.Errorf("model %s is already registered", info.Name)
	}

	if adminConfig == nil {
		adminConfig, err = NewModelAdmin(model)
		if err != nil {
			return err
		}
	}

	s._registry[key] = adminConfig

	// Mock grouping by app name. A real framework infers this from the module path or explicit AppName.
	appName := "main" // fallback
	// For testing and realism, we'll try to extract package name as app name
	pkgParts := strings.Split(info.Type.PkgPath(), "/")
	if len(pkgParts) > 0 && pkgParts[len(pkgParts)-1] != "" {
		appName = pkgParts[len(pkgParts)-1]
	}

	s.appDict[appName] = append(s.appDict[appName], adminConfig)

	return nil
}

// URLs generates the URLconf routing table for the entire admin site.
func (s *AdminSite) URLs() *urls.URLconf {
	patterns := []*urls.URLPattern{
		urls.Path("", s.adminView(s.index), "index", nil),
		urls.Path("<slug:app_label>/", s.adminView(s.appIndex), "app_list", nil),
		urls.Path("<slug:app_label>/<slug:model_name>/", s.adminView(s.changelistView), "changelist", nil),
		urls.Path("<slug:app_label>/<slug:model_name>/add/", s.adminView(s.addView), "add", nil),
		urls.Path("<slug:app_label>/<slug:model_name>/<int:object_id>/change/", s.adminView(s.changeView), "change", nil),
		urls.Path("<slug:app_label>/<slug:model_name>/<int:object_id>/delete/", s.adminView(s.deleteView), "delete", nil),
		urls.Path("<slug:app_label>/<slug:model_name>/<int:object_id>/history/", s.adminView(s.historyView), "history", nil),
		urls.Path("static/<path:path>", s.ServeStatic, "static", nil),
	}

	return &urls.URLconf{
		Patterns:  patterns,
		AppName:   s.Name,
		Namespace: s.Name,
	}
}

// adminView wraps an admin handler to enforce authentication and staff status.
func (s *AdminSite) adminView(view urls.ViewFunc) urls.ViewFunc {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		if req.User == nil || !req.User.IsAuthenticated() {
			loginURL := "/admin/login/"
			return godjangohttp.NewRedirectResponse(fmt.Sprintf("%s?next=%s", loginURL, req.URL.RequestURI()), false)
		}
		if !req.User.IsStaff() {
			// Not a staff user, throw 403 or redirect
			return godjangohttp.NewHttpResponse("You do not have permission to access this page.", http.StatusForbidden)
		}
		return view(req)
	}
}

func (s *AdminSite) getModel(appLabel, modelName string) (*ModelAdmin, error) {
	// Simple mock lookup. In a full system, app configs are queried accurately.
	if modelAdmin, ok := s._registry[strings.ToLower(modelName)]; ok {
		return modelAdmin, nil
	}
	return nil, fmt.Errorf("Model %s not found in app %s", modelName, appLabel)
}

// DefaultAdminSite provides the globally registered admin site instance.
var DefaultAdminSite = NewAdminSite("admin")
