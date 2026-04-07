package sessions

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
)

// BaseSession provides common logic for session implementations.
type BaseSession struct {
	key      string
	data     map[string]any
	modified bool
	backend  Backend
	mu       sync.RWMutex
	loaded   bool
}

// NewBaseSession creates a new BaseSession.
func NewBaseSession(key string, backend Backend) *BaseSession {
	return &BaseSession{
		key:     key,
		backend: backend,
		data:    make(map[string]any),
	}
}

// Load fetches data from the backend if it hasn't been loaded yet.
func (s *BaseSession) load(ctx context.Context) {
	if s.loaded || s.backend == nil {
		return
	}
	s.loaded = true
	if s.key != "" {
		data, err := s.backend.Load(ctx, s.key)
		if err == nil && data != nil {
			s.data = data
		}
	}
}

// Get retrieves a value from the session.
func (s *BaseSession) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.load(context.Background()) // Synchronous load fallback if used outside context flow

	val, ok := s.data[key]
	return val, ok
}

// Set adds a value to the session and marks it as modified.
func (s *BaseSession) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.load(context.Background())

	s.data[key] = value
	s.modified = true
}

// Delete removes a value from the session.
func (s *BaseSession) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.load(context.Background())

	if _, ok := s.data[key]; ok {
		delete(s.data, key)
		s.modified = true
	}
}

// Clear removes all values from the session.
func (s *BaseSession) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.load(context.Background())

	s.data = make(map[string]any)
	s.modified = true
}

// SessionKey returns the current session key.
func (s *BaseSession) SessionKey() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.key
}

// IsModified returns true if the session data has been changed.
func (s *BaseSession) IsModified() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.modified
}

// Save persists the session to the backend if modified.
func (s *BaseSession) Save(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.modified {
		return nil
	}

	if s.key == "" {
		s.key = generateSessionKey()
	}

	err := s.backend.Save(ctx, s.key, s.data, 1209600) // Default 2 weeks in seconds
	if err == nil {
		s.modified = false
	}
	return err
}

// Flush deletes the session from the backend and clears all data.
func (s *BaseSession) Flush(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.key != "" {
		if err := s.backend.Delete(ctx, s.key); err != nil {
			return err
		}
	}

	s.data = make(map[string]any)
	s.key = ""
	s.modified = false
	s.loaded = true
	return nil
}

// CycleKey regenerates the session key while keeping the data.
func (s *BaseSession) CycleKey(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldKey := s.key
	s.key = generateSessionKey()
	s.modified = true

	if oldKey != "" {
		if err := s.backend.Delete(ctx, oldKey); err != nil {
			return err
		}
	}

	return nil
}

func generateSessionKey() string {
	b := make([]byte, 16) // 32 hex chars
	rand.Read(b)
	return hex.EncodeToString(b)
}
