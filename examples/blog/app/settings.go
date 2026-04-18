package app

import (
	"github.com/pkshahid/JanGo/core/settings"
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
			"github.com/pkshahid/JanGo/admin",
			"github.com/pkshahid/JanGo/auth",
			"blog",
		},
	})
}
