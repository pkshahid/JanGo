package queryset

import (
	"fmt"
	ormsignals "github.com/pkshahid/JanGo/orm/signals"
	"strings"
	"sync"
)

// All executes the query and returns all matched records.
func (qs QuerySet[T]) All() ([]T, error) {
	// sql, params := qs.query.ToSQL()
	// db.Query(sql, params...)

	// Mock executing the main query
	results := []T{}

	// Handle PrefetchRelated concurrently via goroutines
	if len(qs.query.PrefetchRelated) > 0 {
		var wg sync.WaitGroup
		errChan := make(chan error, len(qs.query.PrefetchRelated))

		for _, prefetch := range qs.query.PrefetchRelated {
			wg.Add(1)
			go func(lookup string) {
				defer wg.Done()
				// Mock batch query execution (e.g. SELECT * FROM related WHERE parent_id IN (...))
				// err := db.Query(...)
				_ = lookup

			}(prefetch)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			if err != nil {
				return nil, fmt.Errorf("prefetch error: %v", err)
			}
		}
	}

	return results, nil
}

// Get executes the query and returns exactly one record.
// Errors if 0 or >1 records match.
func (qs QuerySet[T]) Get(lookups ...Lookup) (T, error) {
	var zero T
	c := qs.Filter(lookups...)
	c = c.Limit(2) // We limit to 2 to check for MultipleObjectsReturned

	results, err := c.All()
	if err != nil {
		return zero, err
	}

	if len(results) == 0 {
		return zero, fmt.Errorf("orm: DoesNotExist")
	}
	if len(results) > 1 {
		return zero, fmt.Errorf("orm: MultipleObjectsReturned")
	}

	return results[0], nil
}

// First returns the first object matched by the query.
func (qs QuerySet[T]) First() (*T, error) {
	c := qs.clone()
	// If no ordering is specified, Django defaults to PK ASC.
	if len(c.query.OrderBy) == 0 && c.query.ModelInfo.PrimaryKey != nil {
		c = c.OrderBy(c.query.ModelInfo.PrimaryKey.Name)
	}
	c = c.Limit(1)

	results, err := c.All()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil // Return nil, nil for First/Last
	}
	return &results[0], nil
}

// Last returns the last object matched by the query.
func (qs QuerySet[T]) Last() (*T, error) {
	c := qs.clone()
	if len(c.query.OrderBy) == 0 && c.query.ModelInfo.PrimaryKey != nil {
		c = c.OrderBy("-" + c.query.ModelInfo.PrimaryKey.Name)
	} else {
		// Reverse the existing ordering
		var reversed []string
		for _, o := range c.query.OrderBy {
			if strings.HasPrefix(o, "-") {
				reversed = append(reversed, o[1:])
			} else {
				reversed = append(reversed, "-"+o)
			}
		}
		c.query.OrderBy = reversed
	}
	c = c.Limit(1)

	results, err := c.All()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}
	return &results[0], nil
}

// Exists returns true if the QuerySet contains any results.
func (qs QuerySet[T]) Exists() (bool, error) {
	c := qs.clone()
	c.query.OnlyFields = []string{"1"}
	c = c.Limit(1)
	c.query.OrderBy = nil // Order by doesn't matter for exists

	// sql, params := c.query.ToSQL()
	// db.QueryRow(sql, params...)

	// Mock
	return false, nil
}

// Count returns the number of records as an int64.
func (qs QuerySet[T]) Count() (int64, error) {
	c := qs.clone()
	c.query.OrderBy = nil // Remove order by for count

	// sql, params := c.query.ToSQL()
	// sql = "SELECT COUNT(*) FROM (" + sql + ") AS subquery"

	// Mock
	return 0, nil
}

// Values returns dictionaries (maps) rather than model instances.
func (qs QuerySet[T]) Values(fields ...string) ([]map[string]any, error) {
	_ = qs.Only(fields...)

	// Execute SQL and scan into maps instead of structs
	return []map[string]any{}, nil
}

// ValuesList returns tuples (slices) rather than model instances.
func (qs QuerySet[T]) ValuesList(fields ...string) ([][]any, error) {
	_ = qs.Only(fields...)

	// Execute SQL and scan into slices
	return [][]any{}, nil
}

// Create inserts a new record into the database.
func (qs QuerySet[T]) Create(obj *T) error {
	// Trigger PreSave
	ormsignals.PreSave.Send(obj, map[string]any{"instance": obj, "created": true})

	// INSERT INTO table (fields) VALUES (values)

	// Trigger PostSave
	ormsignals.PostSave.Send(obj, map[string]any{"instance": obj, "created": true})
	return nil
}

// Update updates the records matched by the query with the given fields.
func (qs QuerySet[T]) Update(fields map[string]any) (int64, error) {
	// UPDATE table SET col=val WHERE (query conditions)
	return 0, nil
}

// Delete deletes the records matched by the query.
func (qs QuerySet[T]) Delete() (int64, error) {
	var zero T
	ormsignals.PreDelete.Send(zero, map[string]any{"instance": nil}) // In a real ORM we would iterate deleted instances
	// DELETE FROM table WHERE (query conditions)
	ormsignals.PostDelete.Send(zero, map[string]any{"instance": nil})
	return 0, nil
}

// GetOrCreate gets an object, or creates it if it doesn't exist.
func (qs QuerySet[T]) GetOrCreate(kwargs Lookup, defaults map[string]any) (T, bool, error) {
	var zero T
	obj, err := qs.Get(kwargs)
	if err == nil {
		return obj, false, nil // Found
	}

	if err.Error() == "orm: DoesNotExist" {
		// Mock create
		// Create logic merges kwargs and defaults
		return zero, true, nil
	}

	return zero, false, err
}

// UpdateOrCreate updates an object, or creates it if it doesn't exist.
func (qs QuerySet[T]) UpdateOrCreate(kwargs Lookup, defaults map[string]any) (T, bool, error) {
	var zero T
	obj, err := qs.Get(kwargs)
	if err == nil {
		// Update existing
		// _, err = qs.Filter(kwargs).Update(defaults)
		return obj, false, nil
	}

	if err.Error() == "orm: DoesNotExist" {
		// Mock create
		return zero, true, nil
	}

	return zero, false, err
}

// BulkCreate inserts multiple records efficiently.
func (qs QuerySet[T]) BulkCreate(objs []T) error {
	// INSERT INTO table (...) VALUES (...), (...), (...)
	return nil
}

// BulkUpdate updates multiple specific fields on a slice of objects efficiently.
func (qs QuerySet[T]) BulkUpdate(objs []T, fields []string) error {
	// Using CASE statements or DB-specific batch update syntax
	return nil
}
