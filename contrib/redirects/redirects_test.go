package redirects

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()

	// Add permanent redirect
	r, err := store.Add("/old-page", "/new-page", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.StatusCode != http.StatusMovedPermanently {
		t.Errorf("expected 301, got %d", r.StatusCode)
	}

	// Add temporary redirect
	r2, err := store.Add("/temp", "/other", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r2.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", r2.StatusCode)
	}

	// Get redirect
	found, err := store.Get("/old-page")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.NewPath != "/new-page" {
		t.Errorf("expected '/new-page', got %q", found.NewPath)
	}

	// Get non-existent
	_, err = store.Get("/nonexistent")
	if err == nil {
		t.Error("expected error for non-existent redirect")
	}
}

func TestMemoryStoreGetAll(t *testing.T) {
	store := NewMemoryStore()
	store.Add("/a", "/b", true)
	store.Add("/c", "/d", false)

	all, err := store.GetAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 redirects, got %d", len(all))
	}
}

func TestMemoryStoreRemove(t *testing.T) {
	store := NewMemoryStore()
	store.Add("/old", "/new", true)

	err := store.Remove("/old")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = store.Get("/old")
	if err == nil {
		t.Error("redirect should be removed")
	}
}

func TestMiddleware(t *testing.T) {
	store := NewMemoryStore()
	store.Add("/old-page", "/new-page", true)

	mw := NewMiddleware(store)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrapped := mw.Wrap(inner)

	// Request to redirected URL
	req := httptest.NewRequest("GET", "/old-page", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusMovedPermanently {
		t.Errorf("expected 301, got %d", rr.Code)
	}
	location := rr.Header().Get("Location")
	if location != "/new-page" {
		t.Errorf("expected Location '/new-page', got %q", location)
	}

	// Request to non-redirected URL
	req = httptest.NewRequest("GET", "/normal", nil)
	rr = httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestFallbackMiddleware(t *testing.T) {
	store := NewMemoryStore()
	store.Add("/moved", "/destination", false)

	mw := NewFallbackMiddleware(store)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/exists" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("exists"))
			return
		}
		http.NotFound(w, r)
	})

	wrapped := mw.Wrap(inner)

	// Normal request passes through
	req := httptest.NewRequest("GET", "/exists", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	// 404 with redirect rule should redirect
	req = httptest.NewRequest("GET", "/moved", nil)
	rr = httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rr.Code)
	}
}

func TestHandler(t *testing.T) {
	store := NewMemoryStore()
	store.Add("/redirect-me", "/target", true)

	mw := NewMiddleware(store)
	handler := mw.Handler()

	req := httptest.NewRequest("GET", "/redirect-me", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusMovedPermanently {
		t.Errorf("expected 301, got %d", rr.Code)
	}

	req = httptest.NewRequest("GET", "/no-redirect", nil)
	rr = httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}
