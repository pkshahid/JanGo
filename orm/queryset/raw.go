package queryset

import (
	"context"
	"fmt"

	"github.com/pkshahid/JanGo/orm"
	"github.com/pkshahid/JanGo/orm/backends"
)

// RawQuerySet represents a raw SQL query.
type RawQuerySet[T any] struct {
	ModelInfo *orm.ModelInfo
	SQL       string
	Params    []any
	Database  string
}

// Raw creates a raw SQL query.
func (qs QuerySet[T]) Raw(sql string, params ...any) RawQuerySet[T] {
	return RawQuerySet[T]{
		ModelInfo: qs.query.ModelInfo,
		SQL:       sql,
		Params:    params,
		Database:  qs.query.Database,
	}
}

// Using specifies the database to use.
func (rqs RawQuerySet[T]) Using(db string) RawQuerySet[T] {
	rqs.Database = db
	return rqs
}

// getBackend resolves the database backend for this raw query.
func (rqs RawQuerySet[T]) getBackend() (backends.Backend, error) {
	dbAlias := rqs.Database
	if dbAlias == "" {
		dbAlias = "default"
	}
	return backends.GetBackend(dbAlias)
}

// All executes the raw query and returns a slice of models.
func (rqs RawQuerySet[T]) All() ([]T, error) {
	backend, err := rqs.getBackend()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	rows, err := backend.Query(ctx, rqs.SQL, rqs.Params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := []T{}
	for rows.Next() {
		var obj T
		scanDest := buildScanDest(&obj, rqs.ModelInfo, columns)
		if err := rows.Scan(scanDest...); err != nil {
			return nil, err
		}
		results = append(results, obj)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// Get executes the raw query and expects exactly one result.
func (rqs RawQuerySet[T]) Get() (T, error) {
	var zero T
	results, err := rqs.All()
	if err != nil {
		return zero, err
	}

	if len(results) == 0 {
		return zero, fmt.Errorf("orm: RawQuerySet.Get() returned 0 results")
	}
	if len(results) > 1 {
		return zero, fmt.Errorf("orm: RawQuerySet.Get() returned %d results", len(results))
	}

	return results[0], nil
}
