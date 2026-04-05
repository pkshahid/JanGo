package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/godjango/godjango/core/settings"
	godjangohttp "github.com/godjango/godjango/http"
)

// In-memory session store for prototype purposes.
// A real implementation would use caching or database backends.
var sessionStore = struct {
	sync.RWMutex
	data map[string]map[string]any
}{data: make(map[string]map[string]any)}

type MemorySession struct {
	id   string
	data map[string]any
	mu   sync.RWMutex
}

func (s *MemorySession) Get(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

func (s *MemorySession) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *MemorySession) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *MemorySession) save() {
	sessionStore.Lock()
	defer sessionStore.Unlock()

	// Copy data to avoid concurrent map read/write during save
	s.mu.RLock()
	copyData := make(map[string]any)
	for k, v := range s.data {
		copyData[k] = v
	}
	s.mu.RUnlock()

	sessionStore.data[s.id] = copyData
}

// SessionMiddleware loads and saves session data.
func SessionMiddleware(next Handler) Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		s := settings.Get()
		cookieName := s.SESSION_COOKIE_NAME
		if cookieName == "" {
			cookieName = "sessionid"
		}

		sessionID := req.COOKIES[cookieName]

		sessionStore.RLock()
		data, exists := sessionStore.data[sessionID]
		sessionStore.RUnlock()

		if !exists || sessionID == "" {
			bytes := make([]byte, 32)
			rand.Read(bytes)
			sessionID = hex.EncodeToString(bytes)
			data = make(map[string]any)
		}

		memSession := &MemorySession{
			id:   sessionID,
			data: data,
		}

		req.Session = memSession

		resp := next(req)

		// Save session
		memSession.save()

		// Set cookie if HttpResponse
		if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
			cookie := &http.Cookie{
				Name:     cookieName,
				Value:    sessionID,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			}
			hr.Headers.Add("Set-Cookie", cookie.String())
		}

		return resp
	}
}
