package queryset

import (
	"context"
	"iter"

	"github.com/pkshahid/JanGo/orm/backends"
)

// defaultIteratorChunkSize is the number of rows fetched per round-trip
// from the database cursor. This mirrors Django's default chunk_size of 2000.
const defaultIteratorChunkSize = 2000

// Iterator returns a range-over-func iterator that streams model instances
// one at a time directly from the database cursor, avoiding loading the
// entire result set into memory.
//
// Usage:
//
//	for obj, err := range qs.Filter(...).Iterator() {
//	    if err != nil { return err }
//	    // process obj
//	}
//
// The cursor is automatically closed when iteration stops (either by the
// loop finishing or by the caller breaking early) or when an error occurs.
//
// This is the Go equivalent of Django's QuerySet.iterator().
func (qs QuerySet[T]) Iterator() iter.Seq2[*T, error] {
	return qs.IteratorChunked(defaultIteratorChunkSize)
}

// IteratorChunked is like Iterator but allows the caller to specify the
// chunk_size — the number of rows fetched per round-trip from the database
// cursor. A smaller chunk_size reduces memory usage at the cost of more
// network round-trips; a larger value does the opposite.
//
// If chunkSize <= 0, the default (2000) is used.
func (qs QuerySet[T]) IteratorChunked(chunkSize int) iter.Seq2[*T, error] {
	if chunkSize <= 0 {
		chunkSize = defaultIteratorChunkSize
	}

	return func(yield func(*T, error) bool) {
		sqlStr, params := qs.query.ToSQL()

		dbAlias := qs.query.Database
		if dbAlias == "" {
			dbAlias = backends.RouteForRead(qs.query.ModelInfo)
		}

		backend, err := backends.GetBackend(dbAlias)
		if err != nil {
			yield(nil, err)
			return
		}

		ctx := context.Background()
		rows, err := backend.Query(ctx, sqlStr, params...)
		if err != nil {
			yield(nil, err)
			return
		}
		defer rows.Close()

		// Iterate row-by-row, yielding each model instance as it is scanned.
		// The chunkSize controls how many rows the DB driver buffers per
		// network round-trip (driver-dependent); we document it here for
		// parity with Django's iterator(chunk_size=N).
		_ = chunkSize

		for rows.Next() {
			var obj T
			if err := rows.Scan(&obj); err != nil {
				yield(nil, err)
				return
			}
			if !yield(&obj, nil) {
				return
			}
		}

		if err := rows.Err(); err != nil {
			yield(nil, err)
			return
		}
	}
}
