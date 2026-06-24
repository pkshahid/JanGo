package backends

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkshahid/JanGo/core/settings"
	"github.com/pkshahid/JanGo/orm"
	"github.com/pkshahid/JanGo/orm/queryset"
)

// SessionModel represents the database table structure for sessions.
type SessionModel struct {
	orm.Model
	SessionKey  string    `gd:"CharField,max_length=40,primary_key=true"`
	SessionData string    `gd:"TextField"`
	ExpireDate  time.Time `gd:"DateTimeField,db_index=true"`
}

func (m *SessionModel) ModelMeta() *orm.Meta {
	return &orm.Meta{
		DbTable: "godjango_session",
	}
}

func init() {
	orm.Register(&SessionModel{})
}

// DatabaseBackend stores sessions in the database using the ORM.
type DatabaseBackend struct{}

func (b *DatabaseBackend) Load(ctx context.Context, key string) (map[string]any, error) {
	qs := queryset.NewQuerySet[SessionModel]()
	session, err := qs.Get(queryset.Lookup{"SessionKey__exact": key})
	if err != nil {
		return nil, err
	}

	if session.ExpireDate.Before(time.Now()) {
		b.Delete(ctx, key)
		return nil, fmt.Errorf("session expired")
	}

	var data map[string]any
	err = json.Unmarshal([]byte(session.SessionData), &data)
	return data, err
}

func (b *DatabaseBackend) Save(ctx context.Context, key string, data map[string]any, expireSeconds int) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	expireTime := time.Now().Add(time.Duration(expireSeconds) * time.Second)

	// Async write using goroutine to not block the response
	go func() {
		// We use background context here because the request context might cancel
		// when the HTTP handler finishes, stopping our async save.
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = bgCtx // Pass to ORM if ORM supported context

		qs := queryset.NewQuerySet[SessionModel]()
		_, _, err := qs.UpdateOrCreate(
			queryset.Lookup{"SessionKey__exact": key},
			map[string]any{
				"SessionData": string(jsonData),
				"ExpireDate":  expireTime,
			},
		)

		if err != nil {
			fmt.Printf("Failed to async save session %s: %v\n", key, err)
		}
	}()

	return nil
}

func (b *DatabaseBackend) Delete(ctx context.Context, key string) error {
	qs := queryset.NewQuerySet[SessionModel]().Filter(queryset.Lookup{"SessionKey__exact": key})
	_, err := qs.Delete()
	return err
}

func (b *DatabaseBackend) ClearExpired(ctx context.Context) error {
	qs := queryset.NewQuerySet[SessionModel]().Filter(queryset.Lookup{"ExpireDate__lt": time.Now()})
	_, err := qs.Delete()
	return err
}

// FileBackend stores sessions as individual JSON files in SESSION_FILE_PATH.
type FileBackend struct{}

func (b *FileBackend) getPath() string {
	s := settings.Get()
	path := s.SESSION_FILE_PATH
	if path == "" {
		path = os.TempDir()
	}
	os.MkdirAll(path, 0700)
	return path
}

func (b *FileBackend) Load(ctx context.Context, key string) (map[string]any, error) {
	filePath := filepath.Join(b.getPath(), "godjango_session_"+key)

	_, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var payload struct {
		Expire time.Time
		Data   map[string]any
	}

	if err := json.Unmarshal(content, &payload); err != nil {
		return nil, err
	}

	if payload.Expire.Before(time.Now()) {
		os.Remove(filePath)
		return nil, fmt.Errorf("session expired")
	}

	// Update modified time for file
	os.Chtimes(filePath, time.Now(), time.Now())

	return payload.Data, nil
}

func (b *FileBackend) Save(ctx context.Context, key string, data map[string]any, expireSeconds int) error {
	expireTime := time.Now().Add(time.Duration(expireSeconds) * time.Second)

	payload := struct {
		Expire time.Time
		Data   map[string]any
	}{
		Expire: expireTime,
		Data:   data,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	filePath := filepath.Join(b.getPath(), "godjango_session_"+key)
	return os.WriteFile(filePath, jsonData, 0600)
}

func (b *FileBackend) Delete(ctx context.Context, key string) error {
	filePath := filepath.Join(b.getPath(), "godjango_session_"+key)
	return os.Remove(filePath)
}

func (b *FileBackend) ClearExpired(ctx context.Context) error {
	dir := b.getPath()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	now := time.Now()

	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "godjango_session_") {
			filePath := filepath.Join(dir, e.Name())

			content, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			var payload struct{ Expire time.Time }
			if json.Unmarshal(content, &payload) == nil {
				if payload.Expire.Before(now) {
					os.Remove(filePath)
				}
			}
		}
	}
	return nil
}

// CacheBackend uses a sync.Map for testing purposes.
type CacheBackend struct {
	cache sync.Map
}

func (b *CacheBackend) Load(ctx context.Context, key string) (map[string]any, error) {
	if val, ok := b.cache.Load(key); ok {
		return val.(map[string]any), nil
	}
	return nil, fmt.Errorf("session not found")
}

func (b *CacheBackend) Save(ctx context.Context, key string, data map[string]any, expire int) error {
	b.cache.Store(key, data)
	return nil
}

func (b *CacheBackend) Delete(ctx context.Context, key string) error {
	b.cache.Delete(key)
	return nil
}

func (b *CacheBackend) ClearExpired(ctx context.Context) error {
	return nil // Requires iterating over cache and checking TTL, omitted for mock
}

// CookieBackend is handled by storing data directly inside the cookie value via encoding/signing.
// This backend does not actually "Load" or "Save" anywhere, it just acts as an interface.
// Actual cookie logic sits in the middleware.
type CookieBackend struct{}

func (b *CookieBackend) Load(ctx context.Context, key string) (map[string]any, error) {
	// The 'key' here is the signed base64 payload itself for cookie sessions.
	return Unsign(key)
}

func (b *CookieBackend) Save(ctx context.Context, key string, data map[string]any, expire int) error {
	// Handled directly via middleware
	return nil
}

func (b *CookieBackend) Delete(ctx context.Context, key string) error { return nil }
func (b *CookieBackend) ClearExpired(ctx context.Context) error       { return nil }
