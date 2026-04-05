package settings

import (
	"os"
	"sync"
	"testing"
)

func resetSettings() {
	settingsOnce = sync.Once{}
	globalSettings = nil
	registryMu.Lock()
	registry = make(map[string]Settings)
	registryMu.Unlock()
}

func TestValidation(t *testing.T) {
	t.Cleanup(resetSettings)

	s := Settings{}
	err := Configure(s)
	if err == nil {
		t.Fatal("expected validation error for empty SECRET_KEY and ROOT_URLCONF")
	}

	resetSettings()
	s.SECRET_KEY = "secret"
	err = Configure(s)
	if err == nil {
		t.Fatal("expected validation error for empty ROOT_URLCONF")
	}

	resetSettings()
	s.ROOT_URLCONF = "myproject.urls"
	err = Configure(s)
	if err != nil {
		t.Fatalf("did not expect validation error, got: %v", err)
	}
}

func TestEnvironmentOverrides(t *testing.T) {
	t.Cleanup(resetSettings)

	os.Setenv("DEBUG", "true")
	os.Setenv("SECRET_KEY", "env_secret")
	os.Setenv("ALLOWED_HOSTS", "localhost, 127.0.0.1, example.com")
	os.Setenv("EMAIL_PORT", "1025")
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	defer func() {
		os.Unsetenv("DEBUG")
		os.Unsetenv("SECRET_KEY")
		os.Setenv("ALLOWED_HOSTS", "")
		os.Unsetenv("EMAIL_PORT")
		os.Unsetenv("DATABASE_URL")
	}()

	s := Settings{
		SECRET_KEY:   "code_secret",
		ROOT_URLCONF: "urls",
		DATABASES: map[string]DatabaseConfig{
			"default": {
				Engine: "postgres",
			},
		},
	}

	err := Configure(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := Get()
	if got.DEBUG != true {
		t.Errorf("DEBUG not overridden, got %v", got.DEBUG)
	}
	if got.SECRET_KEY != "env_secret" {
		t.Errorf("SECRET_KEY not overridden, got %v", got.SECRET_KEY)
	}
	if len(got.ALLOWED_HOSTS) != 3 || got.ALLOWED_HOSTS[1] != "127.0.0.1" {
		t.Errorf("ALLOWED_HOSTS not overridden correctly, got %v", got.ALLOWED_HOSTS)
	}
	if got.EMAIL_PORT != 1025 {
		t.Errorf("EMAIL_PORT not overridden, got %v", got.EMAIL_PORT)
	}
	if db, ok := got.DATABASES["default"]; !ok || db.DSN != "postgres://user:pass@localhost:5432/db" {
		t.Errorf("DATABASES.DSN not overridden, got %v", got.DATABASES["default"].DSN)
	}
	if db, ok := got.DATABASES["default"]; ok && db.Engine != "postgres" {
		t.Errorf("DATABASES.Engine was overwritten incorrectly, got %v", db.Engine)
	}
}

func TestRegisterAndLoad(t *testing.T) {
	t.Cleanup(resetSettings)

	Register("myproject.settings", Settings{
		SECRET_KEY:   "registered_secret",
		ROOT_URLCONF: "registered_urls",
	})

	os.Setenv("GODJANGO_SETTINGS_MODULE", "myproject.settings")
	defer os.Unsetenv("GODJANGO_SETTINGS_MODULE")

	err := Load()
	if err != nil {
		t.Fatalf("unexpected error loading settings: %v", err)
	}

	got := Get()
	if got.SECRET_KEY != "registered_secret" {
		t.Errorf("got incorrect SECRET_KEY: %v", got.SECRET_KEY)
	}
}
