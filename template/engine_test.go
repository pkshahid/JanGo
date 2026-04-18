package template

import (
	"os"
	"path/filepath"
	"testing"
	"net/http"
	"net/http/httptest"

	godjangohttp "github.com/pkshahid/JanGo/http"
)

func TestEngine(t *testing.T) {
	// Setup dummy template file
	tmpDir, err := os.MkdirTemp("", "godjango_templates")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tmplPath := filepath.Join(tmpDir, "hello.html")
	err = os.WriteFile(tmplPath, []byte("Hello {{ name }}"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	engine := NewEngine([]string{tmpDir}, false)

	tmpl, err := engine.GetTemplate("hello.html")
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}

	ctx := NewContext(map[string]any{"name": "World"})
	res, err := tmpl.Render(ctx)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if res != "Hello World" {
		t.Errorf("Expected 'Hello World', got %q", res)
	}

	// Test RenderToResponse
	req := godjangohttp.NewRequest(httptest.NewRequest("GET", "/", nil))
	resp := tmpl.RenderToResponse(req, ctx)
	hr := resp.(*godjangohttp.HttpResponse)

	if hr.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", hr.StatusCode)
	}
}
