package backends

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/pkshahid/JanGo/core/settings"
	"github.com/pkshahid/JanGo/orm"
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

// GetConnection returns the raw *sql.DB connection for the given database alias.
// This is useful for management commands that need direct database access.
func GetConnection(alias string) (*sql.DB, error) {
	b, err := GetBackend(alias)
	if err != nil {
		return nil, err
	}
	return b.DB(), nil
}

// RouteForRead returns the database alias to use for reading the given model.
// It consults registered routers in order.
func RouteForRead(model *orm.ModelInfo) string {
	managersMu.RLock()
	defer managersMu.RUnlock()

	for _, r := range routers {
		if alias := r.DBForRead(model); alias != "" {
			return alias
		}
	}
	return "default"
}

// RouteForWrite returns the database alias to use for writing the given model.
// It consults registered routers in order.
func RouteForWrite(model *orm.ModelInfo) string {
	managersMu.RLock()
	defer managersMu.RUnlock()

	for _, r := range routers {
		if alias := r.DBForWrite(model); alias != "" {
			return alias
		}
	}
	return "default"
}

// AllowMigrate checks if migration is allowed on the given database for the model.
func AllowMigrate(db string, model *orm.ModelInfo) bool {
	managersMu.RLock()
	defer managersMu.RUnlock()

	for _, r := range routers {
		if !r.AllowMigrate(db, model) {
			return false
		}
	}
	return true
}

// ClearRouters removes all registered routers (for testing).
func ClearRouters() {
	managersMu.Lock()
	defer managersMu.Unlock()
	routers = nil
}

// DefaultRouter routes everything to "default".
type DefaultRouter struct{}

func (DefaultRouter) DBForRead(model *orm.ModelInfo) string             { return "default" }
func (DefaultRouter) DBForWrite(model *orm.ModelInfo) string            { return "default" }
func (DefaultRouter) AllowMigrate(db string, model *orm.ModelInfo) bool { return true }

// AppRouter routes models based on their app label.
type AppRouter struct {
	Routes map[string]string // app_label -> database alias
}

// NewAppRouter creates a router that maps app labels to databases.
func NewAppRouter(routes map[string]string) *AppRouter {
	return &AppRouter{Routes: routes}
}

func (r *AppRouter) DBForRead(model *orm.ModelInfo) string {
	if model == nil {
		return ""
	}
	if alias, ok := r.Routes[model.AppLabel]; ok {
		return alias
	}
	return ""
}

func (r *AppRouter) DBForWrite(model *orm.ModelInfo) string {
	if model == nil {
		return ""
	}
	if alias, ok := r.Routes[model.AppLabel]; ok {
		return alias
	}
	return ""
}

func (r *AppRouter) AllowMigrate(db string, model *orm.ModelInfo) bool {
	if model == nil {
		return true
	}
	if alias, ok := r.Routes[model.AppLabel]; ok {
		return alias == db
	}
	return true
}

// ReadReplicaRouter routes reads to replicas and writes to primary.
type ReadReplicaRouter struct {
	Primary  string
	Replicas []string
	counter  int
}

// NewReadReplicaRouter creates a router with primary/replica routing.
func NewReadReplicaRouter(primary string, replicas ...string) *ReadReplicaRouter {
	return &ReadReplicaRouter{
		Primary:  primary,
		Replicas: replicas,
	}
}

func (r *ReadReplicaRouter) DBForRead(model *orm.ModelInfo) string {
	if len(r.Replicas) == 0 {
		return r.Primary
	}
	// Round-robin across replicas
	idx := r.counter % len(r.Replicas)
	r.counter++
	return r.Replicas[idx]
}

func (r *ReadReplicaRouter) DBForWrite(model *orm.ModelInfo) string {
	return r.Primary
}

func (r *ReadReplicaRouter) AllowMigrate(db string, model *orm.ModelInfo) bool {
	return db == r.Primary
}

// ModelRouter routes specific models to specific databases.
type ModelRouter struct {
	Routes map[string]string // model name (lowercase) -> database alias
}

// NewModelRouter creates a router that maps model names to databases.
func NewModelRouter(routes map[string]string) *ModelRouter {
	return &ModelRouter{Routes: routes}
}

func (r *ModelRouter) DBForRead(model *orm.ModelInfo) string {
	if model == nil {
		return ""
	}
	if alias, ok := r.Routes[model.Name]; ok {
		return alias
	}
	return ""
}

func (r *ModelRouter) DBForWrite(model *orm.ModelInfo) string {
	if model == nil {
		return ""
	}
	if alias, ok := r.Routes[model.Name]; ok {
		return alias
	}
	return ""
}

func (r *ModelRouter) AllowMigrate(db string, model *orm.ModelInfo) bool {
	if model == nil {
		return true
	}
	if alias, ok := r.Routes[model.Name]; ok {
		return alias == db
	}
	return true
}
