package settings

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type DatabaseConfig struct {
	Engine            string `env:"DATABASE_ENGINE"`
	Name              string `env:"DATABASE_NAME"`
	User              string `env:"DATABASE_USER"`
	Password          string `env:"DATABASE_PASSWORD"`
	Host              string `env:"DATABASE_HOST"`
	Port              int    `env:"DATABASE_PORT"`
	DSN               string `env:"DATABASE_URL"`
	MAX_CONN          int    `env:"DATABASE_MAX_CONN"`
	MAX_IDLE          int    `env:"DATABASE_MAX_IDLE"`
	CONN_MAX_LIFETIME int    `env:"DATABASE_CONN_MAX_LIFETIME"` // in seconds
}

type TemplateConfig struct {
	Backend string
	Dirs    []string
	AppDirs bool
	Options map[string]any
}

type CacheConfig struct {
	Backend  string
	Location string
	Options  map[string]any
}

type Settings struct {
	DEBUG               bool     `env:"DEBUG"`
	SECRET_KEY          string   `env:"SECRET_KEY"`
	ALLOWED_HOSTS       []string `env:"ALLOWED_HOSTS"`
	INSTALLED_APPS      []string
	MIDDLEWARE          []string
	ROOT_URLCONF        string
	TEMPLATES           []TemplateConfig
	DATABASES           map[string]DatabaseConfig
	STATIC_URL          string `env:"STATIC_URL"`
	STATIC_ROOT         string `env:"STATIC_ROOT"`
	MEDIA_URL           string `env:"MEDIA_URL"`
	MEDIA_ROOT          string `env:"MEDIA_ROOT"`
	DEFAULT_FILE_STORAGE string `env:"DEFAULT_FILE_STORAGE"`

	// S3 Settings
	AWS_STORAGE_BUCKET_NAME string `env:"AWS_STORAGE_BUCKET_NAME"`
	AWS_S3_REGION_NAME      string `env:"AWS_S3_REGION_NAME"`
	AWS_ACCESS_KEY_ID       string `env:"AWS_ACCESS_KEY_ID"`
	AWS_SECRET_ACCESS_KEY   string `env:"AWS_SECRET_ACCESS_KEY"`
	AWS_S3_ENDPOINT_URL     string `env:"AWS_S3_ENDPOINT_URL"`
	STATICFILES_DIRS    []string
	STATICFILES_STORAGE string
	STATICFILES_FINDERS []string
	CACHES              map[string]CacheConfig
	SESSION_ENGINE      string
	AUTH_USER_MODEL     string
	LOGIN_URL           string
	LOGIN_REDIRECT_URL  string
	TIME_ZONE           string `env:"TIME_ZONE"`
	USE_TZ              bool   `env:"USE_TZ"`
	LANGUAGE_CODE       string `env:"LANGUAGE_CODE"`
	EMAIL_BACKEND       string
	EMAIL_HOST          string `env:"EMAIL_HOST"`
	EMAIL_PORT          int    `env:"EMAIL_PORT"`
	LOGGING             map[string]any

	// Middleware specific settings
	SECURE_SSL_REDIRECT            bool   `env:"SECURE_SSL_REDIRECT"`
	SECURE_HSTS_SECONDS            int    `env:"SECURE_HSTS_SECONDS"`
	SECURE_HSTS_INCLUDE_SUBDOMAINS bool   `env:"SECURE_HSTS_INCLUDE_SUBDOMAINS"`
	SECURE_HSTS_PRELOAD            bool   `env:"SECURE_HSTS_PRELOAD"`
	SECURE_CONTENT_TYPE_NOSNIFF    bool   `env:"SECURE_CONTENT_TYPE_NOSNIFF"`
	X_FRAME_OPTIONS                string `env:"X_FRAME_OPTIONS"`
	SECURE_REFERRER_POLICY         string `env:"SECURE_REFERRER_POLICY"`
	APPEND_SLASH                   bool   `env:"APPEND_SLASH"`
	CSRF_COOKIE_NAME               string `env:"CSRF_COOKIE_NAME"`
	SESSION_COOKIE_NAME            string `env:"SESSION_COOKIE_NAME"`
	SESSION_COOKIE_AGE             int    `env:"SESSION_COOKIE_AGE"`
	SESSION_COOKIE_DOMAIN          string `env:"SESSION_COOKIE_DOMAIN"`
	SESSION_COOKIE_SECURE          bool   `env:"SESSION_COOKIE_SECURE"`
	SESSION_COOKIE_HTTPONLY        bool   `env:"SESSION_COOKIE_HTTPONLY"`
	SESSION_COOKIE_SAMESITE        string `env:"SESSION_COOKIE_SAMESITE"`
	SESSION_FILE_PATH              string `env:"SESSION_FILE_PATH"`
}

var (
	globalSettings *Settings
	settingsOnce   sync.Once
	registry       = make(map[string]Settings)
	registryMu     sync.RWMutex
)

// Register records a settings configuration under a module path.
// This supports loading settings referenced by GODJANGO_SETTINGS_MODULE.
func Register(module string, s Settings) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[module] = s
}

// Load attempts to configure settings using the module specified in GODJANGO_SETTINGS_MODULE.
func Load() error {
	module := os.Getenv("GODJANGO_SETTINGS_MODULE")
	if module == "" {
		return errors.New("GODJANGO_SETTINGS_MODULE is not set")
	}

	registryMu.RLock()
	s, ok := registry[module]
	registryMu.RUnlock()

	if !ok {
		return fmt.Errorf("settings module %q not registered", module)
	}

	return Configure(s)
}

// Configure sets up the global settings using the provided struct.
func Configure(s Settings) error {
	var err error
	settingsOnce.Do(func() {
		applyEnvOverrides(&s)
		if e := s.Validate(); e != nil {
			err = e
			return
		}
		globalSettings = &s
	})
	return err
}

// Get returns the globally configured settings.
// It will panic if Configure or Load has not been successfully called.
func Get() *Settings {
	if globalSettings == nil {
		panic("settings are not configured. Call Configure() or Load() first.")
	}
	return globalSettings
}

// Validate checks required fields on Settings.
func (s *Settings) Validate() error {
	if s.SECRET_KEY == "" {
		return errors.New("SECRET_KEY must not be empty")
	}
	if s.ROOT_URLCONF == "" {
		return errors.New("ROOT_URLCONF must not be empty")
	}
	return nil
}

func applyEnvOverrides(s *Settings) {
	val := reflect.ValueOf(s).Elem()
	applyEnvToStruct(val)

	// Additionally, check for overrides inside the DATABASES map, particularly the DSN.
	// Since Go maps with struct values aren't addressable, we need to rebuild the map if we override.
	if s.DATABASES != nil {
		for key, dbConfig := range s.DATABASES {
			dbVal := reflect.ValueOf(&dbConfig).Elem()
			if applyEnvToStruct(dbVal) {
				s.DATABASES[key] = dbConfig
			}
		}
	}
}

// applyEnvToStruct sets struct fields based on `env` tags.
// Returns true if any field was modified.
func applyEnvToStruct(val reflect.Value) bool {
	modified := false
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		fieldInfo := typ.Field(i)
		envTag := fieldInfo.Tag.Get("env")
		if envTag == "" {
			continue
		}

		envVal, exists := os.LookupEnv(envTag)
		if !exists {
			continue
		}

		fieldVal := val.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		modified = true
		switch fieldVal.Kind() {
		case reflect.String:
			fieldVal.SetString(envVal)
		case reflect.Bool:
			b, err := strconv.ParseBool(envVal)
			if err == nil {
				fieldVal.SetBool(b)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			num, err := strconv.ParseInt(envVal, 10, 64)
			if err == nil {
				fieldVal.SetInt(num)
			}
		case reflect.Slice:
			if fieldVal.Type().Elem().Kind() == reflect.String {
				parts := strings.Split(envVal, ",")
				for j := range parts {
					parts[j] = strings.TrimSpace(parts[j])
				}
				fieldVal.Set(reflect.ValueOf(parts))
			}
		}
	}
	return modified
}
