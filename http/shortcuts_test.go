package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRender(t *testing.T) {
	// Register a simple template for testing
	templates.New("test.html").Parse("Hello {{.name}}")

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	req := NewRequest(r)

	resp := Render(req, "test.html", map[string]any{"name": "World"})

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	recorder := httptest.NewRecorder()
	resp.Write(recorder)

	if recorder.Body.String() != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", recorder.Body.String())
	}
}

func TestRedirect(t *testing.T) {
	resp := Redirect("/new-url", true)
	if resp.StatusCode != http.StatusMovedPermanently {
		t.Errorf("expected 301, got %d", resp.StatusCode)
	}
	if resp.URL != "/new-url" {
		t.Errorf("expected /new-url, got %s", resp.URL)
	}
	if !resp.Permanent {
		t.Errorf("expected permanent to be true")
	}
}

func TestGet404(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	req := NewRequest(r)

	resp := Get404(req)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
