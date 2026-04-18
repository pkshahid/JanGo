package queryset

import (
	"fmt"

	"github.com/pkshahid/JanGo/orm"
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

// All executes the raw query and returns a slice of models.
func (rqs RawQuerySet[T]) All() ([]T, error) {
	// In a real framework, this executes the SQL via database driver.
	// db.Query(rqs.SQL, rqs.Params...)

	// We'll just return an empty slice since we have no DB connection.
	return []T{}, nil
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
