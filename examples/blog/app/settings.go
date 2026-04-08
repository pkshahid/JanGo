package app

import (
	"github.com/godjango/godjango/core/settings"
)

func init() {
	// Configure settings directly
	_ = settings.Configure(settings.Settings{
		DEBUG:        true,
		SECRET_KEY:   "django-insecure-example-blog-key-1234567890",
		ROOT_URLCONF: "blog.urls",
		DATABASES: map[string]settings.DatabaseConfig{
			"default": {
				Engine: "sqlite3",
				Name:   "file::memory:?cache=shared",
			},
		},
		INSTALLED_APPS: []string{
			"github.com/godjango/godjango/admin",
			"github.com/godjango/godjango/auth",
			"blog",
		},
	})
}
