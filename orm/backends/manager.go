package backends

import (
	"fmt"
	"sync"
	"time"

	"github.com/godjango/godjango/core/settings"
)

var (
	managersMu sync.RWMutex
	backends   = make(map[string]Backend)
	routers    []DatabaseRouter
)

// RegisterBackend maps a backend string to a constructor function.
var backendConstructors = make(map[string]func() Backend)

// Register registers a backend constructor for a database engine string.
func Register(engine string, constructor func() Backend) {
	backendConstructors[engine] = constructor
}

// AddRouter registers a DatabaseRouter to the global routing pipeline.
func AddRouter(router DatabaseRouter) {
	managersMu.Lock()
	defer managersMu.Unlock()
	routers = append(routers, router)
}

// Init sets up database connections based on settings.DATABASES.
func Init() error {
	s := settings.Get()

	managersMu.Lock()
	defer managersMu.Unlock()

	for alias, config := range s.DATABASES {
		constructor, ok := backendConstructors[config.Engine]
		if !ok {
			return fmt.Errorf("orm: unsupported database engine: %s", config.Engine)
		}

		backend := constructor()
		err := backend.Connect(config)
		if err != nil {
			return fmt.Errorf("orm: failed to connect to database %s: %v", alias, err)
		}

		// Configure pool limits
		db := backend.DB()
		if db != nil {
			if config.MAX_CONN > 0 {
				db.SetMaxOpenConns(config.MAX_CONN)
			}
			if config.MAX_IDLE > 0 {
				db.SetMaxIdleConns(config.MAX_IDLE)
			}
			if config.CONN_MAX_LIFETIME > 0 {
				db.SetConnMaxLifetime(time.Duration(config.CONN_MAX_LIFETIME) * time.Second)
			}
		}

		backends[alias] = backend
	}

	return nil
}

// Close closes all database connections.
func Close() error {
	managersMu.Lock()
	defer managersMu.Unlock()

	var firstErr error
	for alias, backend := range backends {
		if err := backend.Close(); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("failed to close %s: %v", alias, err)
			}
		}
		delete(backends, alias)
	}
	return firstErr
}

// GetBackend returns the backend for the given alias.
func GetBackend(alias string) (Backend, error) {
	if alias == "" {
		alias = "default"
	}

	managersMu.RLock()
	defer managersMu.RUnlock()

	if b, ok := backends[alias]; ok {
		return b, nil
	}

	return nil, fmt.Errorf("orm: database alias %s not found", alias)
}

// DefaultRouter routes everything to "default".
type DefaultRouter struct{}

func (DefaultRouter) DBForRead(model any) string  { return "default" }
func (DefaultRouter) DBForWrite(model any) string { return "default" }
func (DefaultRouter) AllowMigrate(db string, model any) bool { return true }
