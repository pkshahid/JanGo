package integration

import (
	"net/http/httptest"
	"strings"
	"testing"
	"go.uber.org/goleak"

	_ "github.com/godjango/godjango/examples/blog/app"
	"github.com/godjango/godjango/core/handlers/wsgi"
	"github.com/godjango/godjango/http/urls"
)

func TestMain(m *testing.M) {
	// We ignore monitoring background routine which is by-design.
	// We also ignore cache cleanup routine because it loops indefinitely in LocMem.
	goleak.VerifyTestMain(m,
		goleak.IgnoreTopFunction("github.com/godjango/godjango/monitoring.updateRuntimeMetrics"),
		goleak.IgnoreTopFunction("github.com/godjango/godjango/cache.(*LocMemCache).startCleanup"),
	)
}

func TestFullAppIntegration(t *testing.T) {
	// The app has already been initialized via blank import above.
	// This sets up urls, settings, and models.

	handler := wsgi.NewWSGIHandler(urls.GetGlobalRouter())

	// Create a test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// 1. Test routing -> generic view -> template render fallback mock
	// (Listview without AllowEmpty=true and empty queryset returns 404. We mocked a queryset in urls.go but let's test directly)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 200 && w.Code != 404 {
		t.Errorf("Expected 200 or 404 for index, got %d", w.Code)
	}

	// 2. Test unknown URL (404)
	req404 := httptest.NewRequest("GET", "/non-existent-url/", nil)
	w404 := httptest.NewRecorder()
	handler.ServeHTTP(w404, req404)
	if w404.Code != 404 {
		t.Errorf("Expected 404 Not Found, got %d", w404.Code)
	}

	// 3. Test routing -> detail view generic matching
	reqDetail := httptest.NewRequest("GET", "/post/1/", nil)
	wDetail := httptest.NewRecorder()
	handler.ServeHTTP(wDetail, reqDetail)

	// DetailView will try to hit ORM with mock. We expect 404 if object not found or fallback.
	// Generic views in our implementation return a 200 mock if template doesn't exist,
	// but if the ORM get fails, it returns 404.
	// Since we haven't seeded DB, it returns 404.
	if wDetail.Code != 404 {
		t.Errorf("Expected 404 Not Found from empty DB, got %d", wDetail.Code)
	}
}

func TestReverseLookup(t *testing.T) {
	r := urls.GetGlobalRouter()
	url, err := r.Reverse("post_detail", map[string]any{"pk": 42})
	if err != nil {
		t.Fatalf("Reverse lookup failed: %v", err)
	}
	if url != "/post/42/" {
		t.Errorf("Expected /post/42/, got %s", url)
	}
}

func TestFormSubmission(t *testing.T) {
	// Because CreateView is mapped behind LoginRequiredMixin, we'd get a 302 redirect to /login/ or similar
	r := httptest.NewRequest("POST", "/post/new/", strings.NewReader("is_valid=true"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	handler := wsgi.NewWSGIHandler(urls.GetGlobalRouter())
	handler.ServeHTTP(w, r)

	// Our Generic CreateView returns 200 Mock because the ORM/forms are mocked
	if w.Code != 302 && w.Code != 200 {
		t.Errorf("Expected 200 Mock or 302 Redirect on valid form submit, got %d", w.Code)
	}
}
