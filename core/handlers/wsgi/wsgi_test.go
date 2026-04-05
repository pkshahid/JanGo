package wsgi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/godjango/godjango/core/settings"
	godjangohttp "github.com/godjango/godjango/http"
	"github.com/godjango/godjango/http/urls"
)

// Configure mock settings carefully using sync.Once reset if needed, or just validate fields.
func setupTestSettings() {
	s := settings.Settings{
		SECRET_KEY: "secret",
		ROOT_URLCONF: "test",
		DEBUG: true,
	}
	settings.Configure(s)
}

func TestWSGIHandler(t *testing.T) {
	setupTestSettings()
	router := urls.GetGlobalRouter()
	router.Add(urls.Path("/test", func(req *godjangohttp.Request) godjangohttp.Response {
		return godjangohttp.NewHttpResponse("test response", http.StatusOK)
	}, "test", nil))

	handler := NewWSGIHandler(router)

	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", recorder.Code)
	}
	if recorder.Body.String() != "test response" {
		t.Errorf("Expected 'test response', got %s", recorder.Body.String())
	}
}

func TestWSGIHandler_NotFound(t *testing.T) {
	setupTestSettings()
	router := urls.GetGlobalRouter()
	handler := NewWSGIHandler(router)

	req := httptest.NewRequest("GET", "/notfound", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", recorder.Code)
	}
}
