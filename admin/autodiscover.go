package admin

import (
	"fmt"

	"github.com/pkshahid/JanGo/core/apps"
)

// AdminRegistrar is implemented by applications that wish to register
// models with an AdminSite during autodiscover. Apps embed an
// AdminRegistrarFunc or implement the method directly on their AppConfig.
//
// Example:
//
//	type BlogApp struct{}
//	func (a *BlogApp) Name() string { return "blog" }
//	func (a *BlogApp) Label() string { return "blog" }
//	func (a *BlogApp) Path() string { return "." }
//	func (a *BlogApp) Models() []apps.ModelInfo { return nil }
//	func (a *BlogApp) Ready() {}
//	func (a *BlogApp) RegisterAdmin(site *admin.AdminSite) {
//	    site.Register(Post{}, &admin.ModelAdmin{...})
//	}
type AdminRegistrar interface {
	RegisterAdmin(site *AdminSite)
}

// AdminRegistrarFunc is a function adapter for AdminRegistrar.
type AdminRegistrarFunc func(site *AdminSite)

func (f AdminRegistrarFunc) RegisterAdmin(site *AdminSite) { f(site) }

// Autodiscover iterates over all installed applications (via the app registry)
// and calls RegisterAdmin on any AppConfig that implements AdminRegistrar.
// This mirrors Django's admin.autodiscover() pattern: each app is given the
// opportunity to register its models with the admin site in a single,
// deterministic pass.
//
// Apps that do not implement AdminRegistrar are silently skipped, preserving
// backward compatibility with apps that register via init().
func (s *AdminSite) Autodiscover() error {
	allApps := apps.All()
	if len(allApps) == 0 {
		return nil
	}

	for _, app := range allApps {
		registrar, ok := app.(AdminRegistrar)
		if !ok {
			continue
		}
		registrar.RegisterAdmin(s)
	}

	return nil
}

// AutodiscoverDefault runs Autodiscover against DefaultAdminSite.
func AutodiscoverDefault() error {
	return DefaultAdminSite.Autodiscover()
}

// MustAutodiscover calls Autodiscover and panics on error.
func (s *AdminSite) MustAutodiscover() {
	if err := s.Autodiscover(); err != nil {
		panic(fmt.Sprintf("admin: autodiscover failed: %v", err))
	}
}
