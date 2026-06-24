// Package redirects implements Django's redirects framework.
// It provides database-stored URL redirects with middleware support.
package redirects

import (
	"net/http"
	"strings"
	"sync"
)

// Redirect represents a stored URL redirect rule.
// Equivalent to Django's Redirect model.
type Redirect struct {
	ID          int
	OldPath     string
	NewPath     string
	StatusCode  int  // 301 (permanent) or 302 (temporary)
}

// Store defines the interface for redirect persistence.
type Store interface {
	Get(path string) (*Redirect, error)
	GetAll() ([]*Redirect, error)
	Add(oldPath, newPath string, permanent bool) (*Redirect, error)
	Remove(oldPath string) error
}

// MemoryStore is an in-memory implementation of the redirect store.
type MemoryStore struct {
	mu        sync.RWMutex
	redirects map[string]*Redirect
	nextID    int
}

// NewMemoryStore creates a new in-memory redirect store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		redirects: make(map[string]*Redirect),
		nextID:    1,
	}
}

func (s *MemoryStore) Get(path string) (*Redirect, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path = normalizePath(path)
	r, exists := s.redirects[path]
	if !exists {
		return nil, ErrNotFound
	}
	return r, nil
}

func (s *MemoryStore) GetAll() ([]*Redirect, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Redirect, 0, len(s.redirects))
	for _, r := range s.redirects {
		result = append(result, r)
	}
	return result, nil
}

func (s *MemoryStore) Add(oldPath, newPath string, permanent bool) (*Redirect, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldPath = normalizePath(oldPath)
	statusCode := http.StatusFound // 302
	if permanent {
		statusCode = http.StatusMovedPermanently // 301
	}

	r := &Redirect{
		ID:         s.nextID,
		OldPath:    oldPath,
		NewPath:    newPath,
		StatusCode: statusCode,
	}
	s.redirects[oldPath] = r
	s.nextID++
	return r, nil
}

func (s *MemoryStore) Remove(oldPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldPath = normalizePath(oldPath)
	delete(s.redirects, oldPath)
	return nil
}

// ErrNotFound is returned when a redirect is not found.
var ErrNotFound = &NotFoundError{}

type NotFoundError struct{}

func (e *NotFoundError) Error() string {
	return "redirects: no redirect found for path"
}

// Middleware intercepts requests and performs redirects if a matching rule exists.
type Middleware struct {
	Store Store
}

// NewMiddleware creates a new redirect middleware.
func NewMiddleware(store Store) *Middleware {
	return &Middleware{Store: store}
}

// Wrap wraps an HTTP handler with redirect checking.
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirect, err := m.Store.Get(r.URL.Path)
		if err == nil {
			http.Redirect(w, r, redirect.NewPath, redirect.StatusCode)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Handler returns an HTTP handler that only handles redirects.
func (m *Middleware) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redirect, err := m.Store.Get(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, redirect.NewPath, redirect.StatusCode)
	}
}

// FallbackMiddleware only redirects on 404 (after the main handler).
type FallbackMiddleware struct {
	Store Store
}

// NewFallbackMiddleware creates middleware that only redirects if the main handler returns 404.
func NewFallbackMiddleware(store Store) *FallbackMiddleware {
	return &FallbackMiddleware{Store: store}
}

// Wrap wraps an HTTP handler with fallback redirect checking.
func (m *FallbackMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)

		if rec.statusCode == http.StatusNotFound {
			redirect, err := m.Store.Get(r.URL.Path)
			if err == nil {
				http.Redirect(w, r, redirect.NewPath, redirect.StatusCode)
				return
			}
			// Write the 404
			http.NotFound(w, r)
		}
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.written {
		r.statusCode = code
		if code != http.StatusNotFound {
			r.ResponseWriter.WriteHeader(code)
		}
		r.written = true
	}
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.statusCode == http.StatusNotFound {
		return len(b), nil
	}
	return r.ResponseWriter.Write(b)
}

func normalizePath(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}
