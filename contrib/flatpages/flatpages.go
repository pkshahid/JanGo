// Package flatpages implements Django's flatpages framework.
// It provides simple in-memory stored pages served from configurable URLs.
// Persistence is MemoryStore only — no ORM-backed database storage is used.
package flatpages

import (
	"html/template"
	"net/http"
	"strings"
	"sync"
)

// FlatPage represents a simple content page stored in memory.
// Equivalent to Django's FlatPage model.
type FlatPage struct {
	ID                   int
	URL                  string
	Title                string
	Content              string
	EnableComments       bool
	TemplateName         string
	RegistrationRequired bool
}

// Store defines the interface for flatpage persistence.
type Store interface {
	Get(url string) (*FlatPage, error)
	GetAll() ([]*FlatPage, error)
	Save(page *FlatPage) error
	Delete(url string) error
}

// MemoryStore is the in-memory implementation of Store.
// This is the only supported backend — no ORM-backed store is provided.
type MemoryStore struct {
	mu     sync.RWMutex
	pages  map[string]*FlatPage
	nextID int
}

// NewMemoryStore creates a new in-memory flatpage store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		pages:  make(map[string]*FlatPage),
		nextID: 1,
	}
}

func (s *MemoryStore) Get(url string) (*FlatPage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Normalize URL
	url = normalizeURL(url)
	page, exists := s.pages[url]
	if !exists {
		return nil, ErrPageNotFound
	}
	return page, nil
}

func (s *MemoryStore) GetAll() ([]*FlatPage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pages := make([]*FlatPage, 0, len(s.pages))
	for _, page := range s.pages {
		pages = append(pages, page)
	}
	return pages, nil
}

func (s *MemoryStore) Save(page *FlatPage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	page.URL = normalizeURL(page.URL)
	if page.ID == 0 {
		page.ID = s.nextID
		s.nextID++
	}
	s.pages[page.URL] = page
	return nil
}

func (s *MemoryStore) Delete(url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	url = normalizeURL(url)
	delete(s.pages, url)
	return nil
}

// Clear removes all flatpages from the store (for testing/reset).
func (s *MemoryStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pages = make(map[string]*FlatPage)
	s.nextID = 1
}

// --- Global default store ---

var defaultStore = NewMemoryStore()

// DefaultStore returns the package-level default MemoryStore instance.
func DefaultStore() *MemoryStore {
	return defaultStore
}

// Get retrieves a flatpage by URL from the default store.
func Get(url string) (*FlatPage, error) {
	return defaultStore.Get(url)
}

// GetAll returns all flatpages from the default store.
func GetAll() ([]*FlatPage, error) {
	return defaultStore.GetAll()
}

// Save saves a flatpage to the default store.
func Save(page *FlatPage) error {
	return defaultStore.Save(page)
}

// Delete removes a flatpage from the default store.
func Delete(url string) error {
	return defaultStore.Delete(url)
}

// Clear removes all flatpages from the default store (for testing/reset).
func Clear() {
	defaultStore.Clear()
}

// ErrPageNotFound is returned when a flatpage is not found.
var ErrPageNotFound = &PageNotFoundError{}

type PageNotFoundError struct{}

func (e *PageNotFoundError) Error() string {
	return "flatpages: page not found"
}

// Middleware provides flatpage fallback handling for 404 responses.
// If the request URL matches a flatpage, it renders that page instead of returning 404.
type Middleware struct {
	Store    Store
	Template *template.Template
}

// NewMiddleware creates a flatpage middleware.
func NewMiddleware(store Store, tmpl *template.Template) *Middleware {
	return &Middleware{
		Store:    store,
		Template: tmpl,
	}
}

// Wrap wraps an HTTP handler with flatpage fallback.
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use a response recorder to detect 404s
		rec := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)

		// If not a 404, the response was already written
		if rec.statusCode != http.StatusNotFound {
			return
		}

		// Try to serve a flatpage
		page, err := m.Store.Get(r.URL.Path)
		if err != nil {
			// No flatpage either, write the 404
			http.NotFound(w, r)
			return
		}

		m.renderPage(w, page)
	})
}

// Handler returns an HTTP handler that directly serves flatpages.
func (m *Middleware) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, err := m.Store.Get(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		m.renderPage(w, page)
	}
}

func (m *Middleware) renderPage(w http.ResponseWriter, page *FlatPage) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if m.Template != nil {
		data := map[string]interface{}{
			"flatpage": page,
			"title":    page.Title,
			"content":  template.HTML(page.Content),
		}
		tmplName := page.TemplateName
		if tmplName == "" {
			tmplName = "flatpages/default.html"
		}
		err := m.Template.ExecuteTemplate(w, tmplName, data)
		if err != nil {
			// Fallback to simple rendering
			_, _ = w.Write([]byte(page.Content))
		}
		return
	}

	// Simple rendering without template
	_, _ = w.Write([]byte("<html><head><title>" + page.Title + "</title></head><body>"))
	_, _ = w.Write([]byte(page.Content))
	_, _ = w.Write([]byte("</body></html>"))
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (r *responseRecorder) WriteHeader(code int) {
	if !r.written {
		r.statusCode = code
		if code != http.StatusNotFound {
			r.ResponseWriter.WriteHeader(code)
		}
		r.written = true
	}
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.statusCode == http.StatusNotFound {
		return len(b), nil // discard 404 body
	}
	return r.ResponseWriter.Write(b)
}

func normalizeURL(url string) string {
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	return url
}
