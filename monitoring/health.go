package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pkshahid/JanGo/core/settings"
	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/orm/backends"
	"github.com/pkshahid/JanGo/cache"
)

// Check defines the interface for a health check.
type Check interface {
	Name() string
	Run(ctx context.Context) error
}

var (
	registryMu sync.RWMutex
	checks     = make(map[string]Check)
)

// RegisterCheck adds a health check to the global registry.
func RegisterCheck(check Check) {
	registryMu.Lock()
	defer registryMu.Unlock()
	checks[check.Name()] = check
}

// DBCheck verifies database connectivity.
type DBCheck struct{}

func (c *DBCheck) Name() string { return "database" }
func (c *DBCheck) Run(ctx context.Context) error {
	s := settings.Get()
	if len(s.DATABASES) == 0 {
		return nil // No databases configured, so it's technically "healthy" by default
	}

	backend, err := backends.GetBackend("default")
	if err != nil {
		return err
	}

	db := backend.DB()
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Ping the database
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}

// CacheCheck verifies cache connectivity.
type CacheCheck struct{}

func (c *CacheCheck) Name() string { return "cache" }
func (c *CacheCheck) Run(ctx context.Context) error {
	// Attempt to reach default cache if configured
	cch := cache.Get("default")
	if cch == nil {
		return nil // No cache configured, so it's "healthy"
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// A simple Set/Get or ping would suffice. The interface doesn't always have a Ping.
	// We'll try to get a dummy key which should return nil error (or miss, but not a connection error).
	_, err := cch.Get(ctx, "__health_check_dummy_key__")
	if err != nil && err.Error() != "cache: key not found" && err.Error() != "redis: nil" && err.Error() != "memcache: cache miss" {
		// Depending on backend, a miss is not an error for health purposes, but a connection refused is.
		return err
	}

	return nil
}

func init() {
	// Register default built-in checks
	RegisterCheck(&DBCheck{})
	RegisterCheck(&CacheCheck{})
}

// HealthView returns a JSON response indicating the status of all registered checks.
func HealthView(req *godjangohttp.Request) godjangohttp.Response {
	registryMu.RLock()
	var checkList []Check
	for _, c := range checks {
		checkList = append(checkList, c)
	}
	registryMu.RUnlock()

	ctx := req.Context
	status := "ok"
	statusCode := http.StatusOK
	results := make(map[string]string)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, check := range checkList {
		wg.Add(1)
		go func(c Check) {
			defer wg.Done()
			err := c.Run(ctx)

			mu.Lock()
			if err != nil {
				results[c.Name()] = err.Error()
				status = "error"
				statusCode = http.StatusServiceUnavailable
			} else {
				results[c.Name()] = "ok"
			}
			mu.Unlock()
		}(check)
	}

	wg.Wait()

	data := map[string]any{
		"status": status,
		"checks": results,
	}

	res := godjangohttp.NewJsonResponse(data)
	res.StatusCode = statusCode
	return res
}
