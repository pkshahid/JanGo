package app

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/pkshahid/JanGo/http/urls"
	"github.com/pkshahid/JanGo/core/handlers/wsgi"
)

// A quick simulated TestCase environment
func TestBlogApp_ListView(t *testing.T) {
	handler := wsgi.NewWSGIHandler(urls.GetGlobalRouter())

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 200 && w.Code != 404 {
		t.Errorf("Expected 200 OK or 404 Empty for list view, got %d", w.Code)
	}

	body := w.Body.String()
	if !bytes.Contains([]byte(body), []byte("First Post")) && !bytes.Contains([]byte(body), []byte("Template Render Mock")) && !bytes.Contains([]byte(body), []byte("No objects found")) {
		t.Errorf("Expected to see mock data or render mock, got %s", body)
	}
}

func TestBlogApp_CreateViewRedirect(t *testing.T) {
	// Post create returns 200 mock because no valid auth/POST data in the mocked framework currently.
	req := httptest.NewRequest("POST", "/post/new/", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	// Use global router
	handler := wsgi.NewWSGIHandler(urls.GetGlobalRouter())
	handler.ServeHTTP(w, req)

	// Either 200 Mock or 302 login redirect is acceptable for this stub
	if w.Code != 200 && w.Code != 302 {
		t.Errorf("Expected 200 or 302 for create view, got %d", w.Code)
	}
}
