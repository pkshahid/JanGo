package static_test

import (
	"github.com/pkshahid/JanGo/core/settings"
	"github.com/pkshahid/JanGo/static/management"
	"github.com/pkshahid/JanGo/static/middleware"
	"github.com/pkshahid/JanGo/template"
	godjangotags "github.com/pkshahid/JanGo/template/tags"

	"os"
	"strings"
	"testing"
)

func setupSettings() {
	s := settings.Settings{
		STATIC_URL:          "/static/",
		STATIC_ROOT:         "test_static_root",
		STATICFILES_DIRS:    []string{"test_static_dirs"},
		STATICFILES_STORAGE: "ManifestStaticFilesStorage",
		SECRET_KEY:          "test",
		ROOT_URLCONF:        "test",
	}
	settings.Configure(s)

	os.MkdirAll("test_static_dirs", 0755)
	os.WriteFile("test_static_dirs/style.css", []byte("body { color: red; }"), 0644)
}

func cleanupSettings() {
	os.RemoveAll("test_static_root")
	os.RemoveAll("test_static_dirs")
}

func TestStaticTemplateTag(t *testing.T) {
	setupSettings()
	defer cleanupSettings()

	engine := template.NewEngine(nil, false)
	lib := template.NewLibrary()
	godjangotags.RegisterWebTags(lib)
	engine.RegisterLibrary("static", lib)

	coreLib := template.NewLibrary()
	godjangotags.RegisterInheritanceTags(coreLib)
	engine.AddBuiltin(coreLib)

	// Test normal path
	tmpl, err := engine.FromString(`{% load static %}{% static "style.css" %}`)
	if err != nil {
		t.Fatalf("Failed to compile template: %v", err)
	}

	ctx := template.NewContext(map[string]any{})
	output, err := tmpl.Render(ctx)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if output != "/static/style.css" {
		t.Errorf("Expected /static/style.css, got %q", output)
	}

	// Test as variable
	tmpl2, err := engine.FromString(`{% load static %}{% static "style.css" as mystatic %}{{ mystatic }}`)
	if err != nil {
		t.Fatalf("Failed to compile template: %v", err)
	}
	output2, err := tmpl2.Render(ctx)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}
	if output2 != "/static/style.css" {
		t.Errorf("Expected '/static/style.css' when using 'as', got %q", output2)
	}
	val, ok := ctx.Get("mystatic")
	if !ok || val != "/static/style.css" {
		t.Errorf("Expected context variable 'mystatic' to be /static/style.css, got %v", val)
	}
}

func TestCollectStaticCommand(t *testing.T) {
	setupSettings()
	defer cleanupSettings()

	cmd := &management.CollectStaticCommand{}
	cmd.Execute(nil, []string{})

	// Check if file was copied
	if _, err := os.Stat("test_static_root/style.css"); os.IsNotExist(err) {
		t.Errorf("style.css was not copied to STATIC_ROOT")
	}

	// Check if manifest was created
	if _, err := os.Stat("test_static_root/staticfiles.json"); os.IsNotExist(err) {
		t.Errorf("staticfiles.json was not created")
	}

	// Check if hashed file was created
	files, _ := os.ReadDir("test_static_root")
	foundHashed := false
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "style.") && strings.HasSuffix(f.Name(), ".css") && f.Name() != "style.css" {
			foundHashed = true
			break
		}
	}
	if !foundHashed {
		t.Errorf("Hashed style.css was not created")
	}
}

func TestWhiteNoiseMiddleware(t *testing.T) {
	setupSettings()
	defer cleanupSettings()

	cmd := &management.CollectStaticCommand{}
	cmd.Execute(nil, []string{})

	wm := middleware.NewWhiteNoiseMiddleware()

	if wm == nil {
		t.Errorf("Middleware is nil")
	}
}
