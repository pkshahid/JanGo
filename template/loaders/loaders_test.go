package loaders

import (
	"embed"
	"os"
	"path/filepath"
	"testing"
)

//go:embed testdata/*
var testFS embed.FS

func TestFilesystemLoader(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "godjango_loaders")
	defer os.RemoveAll(tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "test.html"), []byte("FS Content"), 0644)

	loader := NewFilesystemLoader([]string{tmpDir})
	content, err := loader.Load("test.html")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}
	if content != "FS Content" {
		t.Errorf("Expected 'FS Content', got %s", content)
	}

	_, err = loader.Load("missing.html")
	if err == nil {
		t.Errorf("Expected error for missing file")
	}
}

func TestAppDirectoriesLoader(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "godjango_apps")
	defer os.RemoveAll(tmpDir)

	appDir := filepath.Join(tmpDir, "myapp")
	tmplDir := filepath.Join(appDir, "templates")
	os.MkdirAll(tmplDir, 0755)

	os.WriteFile(filepath.Join(tmplDir, "app.html"), []byte("App Content"), 0644)

	loader := NewAppDirectoriesLoader([]string{appDir})
	content, err := loader.Load("app.html")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}
	if content != "App Content" {
		t.Errorf("Expected 'App Content', got %s", content)
	}
}

func TestCachedLoader(t *testing.T) {
	locmem := NewLocmemLoader(map[string]string{
		"test.html": "Locmem Content",
	})
	cached := NewCachedLoader(locmem)

	content, _ := cached.Load("test.html")
	if content != "Locmem Content" {
		t.Errorf("Expected 'Locmem Content', got %s", content)
	}

	// Change underlying locmem to verify cache works
	locmem.Templates["test.html"] = "Changed Content"

	content2, _ := cached.Load("test.html")
	if content2 != "Locmem Content" {
		t.Errorf("Cache failed, got %s", content2)
	}
}

func TestLocmemLoader(t *testing.T) {
	loader := NewLocmemLoader(map[string]string{
		"test.html": "Locmem Content",
	})

	content, err := loader.Load("test.html")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}
	if content != "Locmem Content" {
		t.Errorf("Expected 'Locmem Content', got %s", content)
	}
}

func TestEmbedLoader(t *testing.T) {
	// We use the dummy embed.FS from this test file package.
	// Since we haven't created testdata/ yet, we will just test missing file handling.
	loader := NewEmbedLoader(testFS, "testdata")

	_, err := loader.Load("missing.html")
	if err == nil {
		t.Errorf("Expected error for missing file in embed FS")
	}
}
