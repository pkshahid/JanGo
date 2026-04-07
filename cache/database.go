package cache

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/godjango/godjango/core/settings"
	"github.com/godjango/godjango/orm/backends"
)

func init() {
	RegisterBackend("DatabaseCache", NewDatabaseCache)
}

type DatabaseCache struct {
	dbAlias   string
	tableName string
	prefix    string
	cleanup   *time.Ticker
	stop      chan struct{}
}

func NewDatabaseCache(alias string, config settings.CacheConfig) Cache {
	tableName, ok := config.Options["LOCATION"].(string)
	if !ok || tableName == "" {
		tableName = "cache_table"
	}
	dbAlias, ok := config.Options["DB_ALIAS"].(string)
	if !ok || dbAlias == "" {
		dbAlias = "default"
	}
	prefix, _ := config.Options["KEY_PREFIX"].(string)

	c := &DatabaseCache{
		dbAlias:   dbAlias,
		tableName: tableName,
		prefix:    prefix,
		cleanup:   time.NewTicker(time.Minute * 10), // Background cleanup interval
		stop:      make(chan struct{}),
	}

	// Start background cleanup
	go c.startCleanup()
	return c
}

func (c *DatabaseCache) makeKey(key string) string {
	if c.prefix != "" {
		return c.prefix + ":" + key
	}
	return key
}

func (c *DatabaseCache) startCleanup() {
	for {
		select {
		case <-c.cleanup.C:
			c.cull()
		case <-c.stop:
			c.cleanup.Stop()
			return
		}
	}
}

func (c *DatabaseCache) getDB() (*sql.DB, error) {
	backend, err := backends.GetBackend(c.dbAlias)
	if err != nil {
		return nil, err
	}
	return backend.DB(), nil
}

func (c *DatabaseCache) cull() {
	db, err := c.getDB()
	if err != nil {
		return
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE expires < ?", c.tableName)
	db.ExecContext(context.Background(), query, time.Now().UTC())
}

func (c *DatabaseCache) Get(ctx context.Context, key string) (any, error) {
	db, err := c.getDB()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT value, expires FROM %s WHERE cache_key = ?", c.tableName)
	row := db.QueryRowContext(ctx, query, c.makeKey(key))

	var valueData []byte
	var expires time.Time

	err = row.Scan(&valueData, &expires)
	if err == sql.ErrNoRows {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}

	if !expires.IsZero() && time.Now().UTC().After(expires) {
		// Key expired, delete it and return miss
		c.Delete(ctx, key)
		return nil, ErrCacheMiss
	}

	return Deserialize(valueData)
}

func (c *DatabaseCache) Set(ctx context.Context, key string, value any, timeout time.Duration) error {
	db, err := c.getDB()
	if err != nil {
		return err
	}

	data, err := Serialize(value)
	if err != nil {
		return err
	}

	var expires time.Time
	if timeout == 0 {
		// Use a far future date for "never expire"
		expires = time.Now().UTC().AddDate(100, 0, 0)
	} else {
		expires = time.Now().UTC().Add(timeout)
	}

	s := settings.Get()
	dbConfig := s.DATABASES[c.dbAlias]

	var query string
	if dbConfig.Engine == "mysql" {
		query = fmt.Sprintf("INSERT INTO %s (cache_key, value, expires) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE value=VALUES(value), expires=VALUES(expires)", c.tableName)
	} else {
		query = fmt.Sprintf("INSERT INTO %s (cache_key, value, expires) VALUES (?, ?, ?) ON CONFLICT(cache_key) DO UPDATE SET value=excluded.value, expires=excluded.expires", c.tableName)
	}
	_, err = db.ExecContext(ctx, query, c.makeKey(key), string(data), expires)

	return err
}

func (c *DatabaseCache) Delete(ctx context.Context, key string) error {
	db, err := c.getDB()
	if err != nil {
		return err
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE cache_key = ?", c.tableName)
	_, err = db.ExecContext(ctx, query, c.makeKey(key))
	return err
}

func (c *DatabaseCache) Clear(ctx context.Context) error {
	db, err := c.getDB()
	if err != nil {
		return err
	}
	query := fmt.Sprintf("DELETE FROM %s", c.tableName)
	_, err = db.ExecContext(ctx, query)
	return err
}

func (c *DatabaseCache) GetMany(ctx context.Context, keys []string) (map[string]any, error) {
	if len(keys) == 0 {
		return make(map[string]any), nil
	}

	db, err := c.getDB()
	if err != nil {
		return nil, err
	}

	prefixedKeys := make([]any, len(keys))
	placeholders := make([]string, len(keys))
	for i, k := range keys {
		prefixedKeys[i] = c.makeKey(k)
		placeholders[i] = "?"
	}

	query := fmt.Sprintf("SELECT cache_key, value, expires FROM %s WHERE cache_key IN (%s)", c.tableName, strings.Join(placeholders, ","))
	rows, err := db.QueryContext(ctx, query, prefixedKeys...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]any)
	now := time.Now().UTC()

	prefixLen := len(c.prefix)
	if c.prefix != "" {
		prefixLen++ // for the colon
	}

	for rows.Next() {
		var key string
		var data []byte
		var expires time.Time

		if err := rows.Scan(&key, &data, &expires); err != nil {
			continue
		}

		if !expires.IsZero() && now.After(expires) {
			continue // Skip expired keys
		}

		val, err := Deserialize(data)
		if err == nil {
			originalKey := key[prefixLen:]
			results[originalKey] = val
		}
	}
	return results, nil
}

func (c *DatabaseCache) SetMany(ctx context.Context, values map[string]any, timeout time.Duration) error {
	// A robust implementation would use a transaction and batch inserts.
	for k, v := range values {
		if err := c.Set(ctx, k, v, timeout); err != nil {
			return err
		}
	}
	return nil
}

func (c *DatabaseCache) DeleteMany(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	db, err := c.getDB()
	if err != nil {
		return err
	}

	prefixedKeys := make([]any, len(keys))
	placeholders := make([]string, len(keys))
	for i, k := range keys {
		prefixedKeys[i] = c.makeKey(k)
		placeholders[i] = "?"
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE cache_key IN (%s)", c.tableName, strings.Join(placeholders, ","))
	_, err = db.ExecContext(ctx, query, prefixedKeys...)
	return err
}

func (c *DatabaseCache) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	// Simple non-atomic implementation for now
	valAny, err := c.Get(ctx, key)
	if err != nil {
		if err == ErrCacheMiss {
			// Set initial value
			if err := c.Set(ctx, key, delta, 0); err != nil {
				return 0, err
			}
			return delta, nil
		}
		return 0, err
	}

	var num int64
	switch v := valAny.(type) {
	case float64:
		num = int64(v)
	case int64:
		num = v
	default:
		return 0, fmt.Errorf("value is not an integer")
	}

	num += delta
	if err := c.Set(ctx, key, num, 0); err != nil { // Note: reset timeout
		return 0, err
	}

	return num, nil
}

func (c *DatabaseCache) Decr(ctx context.Context, key string, delta int64) (int64, error) {
	return c.Incr(ctx, key, -delta)
}

func (c *DatabaseCache) GetOrSet(ctx context.Context, key string, defaultFunc func() any, timeout time.Duration) (any, error) {
	val, err := c.Get(ctx, key)
	if err == nil {
		return val, nil
	}
	if err != ErrCacheMiss {
		return nil, err
	}

	newVal := defaultFunc()
	if err := c.Set(ctx, key, newVal, timeout); err != nil {
		return nil, err
	}
	return newVal, nil
}

func (c *DatabaseCache) Has(ctx context.Context, key string) (bool, error) {
	db, err := c.getDB()
	if err != nil {
		return false, err
	}

	query := fmt.Sprintf("SELECT expires FROM %s WHERE cache_key = ?", c.tableName)
	row := db.QueryRowContext(ctx, query, c.makeKey(key))

	var expires time.Time
	err = row.Scan(&expires)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if !expires.IsZero() && time.Now().UTC().After(expires) {
		return false, nil
	}

	return true, nil
}
