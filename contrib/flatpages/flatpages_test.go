package flatpages

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()

	// Save a page
	page := &FlatPage{
		URL:     "/about/",
		Title:   "About Us",
		Content: "<p>We are a company.</p>",
	}
	err := store.Save(page)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.ID != 1 {
		t.Errorf("expected ID 1, got %d", page.ID)
	}

	// Get the page
	found, err := store.Get("/about/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Title != "About Us" {
		t.Errorf("expected 'About Us', got %q", found.Title)
	}

	// Get non-existent
	_, err = store.Get("/nonexistent/")
	if err == nil {
		t.Error("expected error for non-existent page")
	}
}

func TestMemoryStoreGetAll(t *testing.T) {
	store := NewMemoryStore()
	store.Save(&FlatPage{URL: "/about/", Title: "About"})
	store.Save(&FlatPage{URL: "/terms/", Title: "Terms"})

	pages, err := store.GetAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) != 2 {
		t.Errorf("expected 2 pages, got %d", len(pages))
	}
}

func TestMemoryStoreDelete(t *testing.T) {
	store := NewMemoryStore()
	store.Save(&FlatPage{URL: "/temp/", Title: "Temp"})

	err := store.Delete("/temp/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = store.Get("/temp/")
	if err == nil {
		t.Error("page should be deleted")
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/about/", "/about/"},
		{"/about", "/about/"},
		{"about/", "/about/"},
		{"about", "/about/"},
	}

	for _, tc := range tests {
		result := normalizeURL(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeURL(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestMiddlewareHandler(t *testing.T) {
	store := NewMemoryStore()
	store.Save(&FlatPage{
		URL:     "/about/",
		Title:   "About",
		Content: "<h1>About Page</h1>",
	})

	mw := NewMiddleware(store, nil)
	handler := mw.Handler()

	// Existing page
	req := httptest.NewRequest("GET", "/about/", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "About Page") {
		t.Error("response should contain page content")
	}
	if !strings.Contains(body, "<title>About</title>") {
		t.Error("response should contain page title")
	}

	// Non-existent page
	req = httptest.NewRequest("GET", "/nonexistent/", nil)
	rr = httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestDefaultStore(t *testing.T) {
	Clear()
	defer Clear()

	// Save via package-level function
	page := &FlatPage{
		URL:     "/global/",
		Title:   "Global",
		Content: "<p>Global page</p>",
	}
	err := Save(page)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.ID == 0 {
		t.Error("expected non-zero ID")
	}

	// Get via package-level function
	found, err := Get("/global/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Title != "Global" {
		t.Errorf("expected 'Global', got %q", found.Title)
	}

	// GetAll via package-level function
	all, err := GetAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 page, got %d", len(all))
	}

	// Delete via package-level function
	err = Delete("/global/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = Get("/global/")
	if err == nil {
		t.Error("page should be deleted")
	}

	// Clear
	Save(&FlatPage{URL: "/temp-clear/", Title: "Temp"})
	Clear()
	all, err = GetAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected 0 pages after Clear, got %d", len(all))
	}
}

func TestDefaultStoreSameInstance(t *testing.T) {
	Clear()
	defer Clear()

	s1 := DefaultStore()
	s2 := DefaultStore()
	if s1 != s2 {
		t.Error("DefaultStore should return the same instance")
	}
}

func TestMiddlewareWrap(t *testing.T) {
	store := NewMemoryStore()
	store.Save(&FlatPage{
		URL:     "/flat/",
		Title:   "Flat Page",
		Content: "<p>Content</p>",
	})

	mw := NewMiddleware(store, nil)

	// Create a handler that returns 404 for /flat/
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/normal/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Normal page"))
			return
		}
		http.NotFound(w, r)
	})

	wrapped := mw.Wrap(inner)

	// Normal page should pass through
	req := httptest.NewRequest("GET", "/normal/", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for normal page, got %d", rr.Code)
	}

	// Flatpage should be served on 404
	req = httptest.NewRequest("GET", "/flat/", nil)
	rr = httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)
	if !strings.Contains(rr.Body.String(), "Content") {
		t.Error("flatpage should be served as fallback")
	}
}
