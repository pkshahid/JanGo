package sessions

import "context"

// Session represents a user session.
type Session interface {
	Get(key string) (any, bool)
	Set(key string, value any)
	Delete(key string)
	Clear()
	SessionKey() string
	IsModified() bool
	Flush(ctx context.Context) error
	CycleKey(ctx context.Context) error
	Save(ctx context.Context) error
}

// Backend defines the storage mechanisms for a session.
type Backend interface {
	Load(ctx context.Context, key string) (map[string]any, error)
	Save(ctx context.Context, key string, data map[string]any, expire int) error
	Delete(ctx context.Context, key string) error
	ClearExpired(ctx context.Context) error
}
