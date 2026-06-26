package queryset

import (
	"iter"
	"time"
)

// ManagerInterface is the interface that all managers implement.
// It allows the framework to identify manager fields on model structs
// via reflection and treat them uniformly.
type ManagerInterface interface {
	IsManager()
}

// Manager provides table-level operations for a model type T.
// It is the Go equivalent of Django's Manager, wrapping a QuerySet[T]
// and delegating query-building and execution methods to it.
//
// Custom managers can be created by embedding Manager[T] and either:
//   - Adding custom methods that call GetQuerySet(), or
//   - Using WithQuerySet() to override the base queryset (equivalent to
//     overriding get_queryset() in Django).
//
// Because Go method promotion is not polymorphic, WithQuerySet() is the
// recommended way to customise the base queryset for a custom manager.
// Overriding GetQuerySet() on a custom manager type will NOT affect the
// delegated methods (Filter, Exclude, etc.) because they call the embedded
// Manager[T].GetQuerySet(), not the outer type's method.
//
// Example — custom manager with a pre-filtered queryset:
//
//	type PublishedManager struct {
//	    queryset.Manager[Post]
//	}
//
//	func NewPublishedManager() PublishedManager {
//	    return PublishedManager{
//	        Manager: queryset.NewManager[Post]().WithQuerySet(func() queryset.QuerySet[Post] {
//	            return queryset.NewQuerySet[Post]().Filter(queryset.Lookup{"Published": true})
//	        }),
//	    }
//	}
//
//	func (m PublishedManager) Published() queryset.QuerySet[Post] {
//	    return m.GetQuerySet()
//	}
//
// Example — custom manager with extra methods only:
//
//	type PostManager struct {
//	    queryset.Manager[Post]
//	}
//
//	func (m PostManager) WithTag(tag string) queryset.QuerySet[Post] {
//	    return m.GetQuerySet().Filter(queryset.Lookup{"Tags__contains": tag})
//	}
type Manager[T any] struct {
	getQuerySetFn func() QuerySet[T]
}

// NewManager creates a new Manager for type T.
// The model T must be registered with the ORM before the manager's
// methods are called (GetQuerySet panics if the model is unregistered,
// mirroring NewQuerySet's behaviour).
func NewManager[T any]() Manager[T] {
	return Manager[T]{
		getQuerySetFn: func() QuerySet[T] {
			return NewQuerySet[T]()
		},
	}
}

// GetQuerySet returns the base QuerySet for this manager.
// Custom managers can override the queryset by using WithQuerySet().
func (m Manager[T]) GetQuerySet() QuerySet[T] {
	if m.getQuerySetFn != nil {
		return m.getQuerySetFn()
	}
	return NewQuerySet[T]()
}

// WithQuerySet sets a custom queryset factory for this manager,
// equivalent to overriding get_queryset() in Django.
// Returns a new Manager with the custom factory so the original
// manager is not mutated (consistent with QuerySet immutability).
func (m Manager[T]) WithQuerySet(fn func() QuerySet[T]) Manager[T] {
	m.getQuerySetFn = fn
	return m
}

// IsManager marks Manager[T] as implementing ManagerInterface.
func (m Manager[T]) IsManager() {}

// ---------------------------------------------------------------------------
// QuerySet delegation — query-building methods (return new QuerySet[T])
// ---------------------------------------------------------------------------

// Filter adds an AND WHERE clause, delegated to the base QuerySet.
func (m Manager[T]) Filter(lookups ...Lookup) QuerySet[T] {
	return m.GetQuerySet().Filter(lookups...)
}

// FilterQ adds a Q object to the WHERE clause using AND.
func (m Manager[T]) FilterQ(q *QNode) QuerySet[T] {
	return m.GetQuerySet().FilterQ(q)
}

// Exclude adds a NOT WHERE clause.
func (m Manager[T]) Exclude(lookups ...Lookup) QuerySet[T] {
	return m.GetQuerySet().Exclude(lookups...)
}

// ExcludeQ adds a Q object to the Exclude clause using AND.
func (m Manager[T]) ExcludeQ(q *QNode) QuerySet[T] {
	return m.GetQuerySet().ExcludeQ(q)
}

// OrderBy orders the results by the specified fields.
func (m Manager[T]) OrderBy(fields ...string) QuerySet[T] {
	return m.GetQuerySet().OrderBy(fields...)
}

// Limit limits the number of results.
func (m Manager[T]) Limit(n int) QuerySet[T] {
	return m.GetQuerySet().Limit(n)
}

// Offset sets the offset for the query results.
func (m Manager[T]) Offset(n int) QuerySet[T] {
	return m.GetQuerySet().Offset(n)
}

// Distinct ensures the results are distinct.
func (m Manager[T]) Distinct() QuerySet[T] {
	return m.GetQuerySet().Distinct()
}

// Only restricts the fetched columns to the specified fields.
func (m Manager[T]) Only(fields ...string) QuerySet[T] {
	return m.GetQuerySet().Only(fields...)
}

// Defer prevents the specified fields from being fetched initially.
func (m Manager[T]) Defer(fields ...string) QuerySet[T] {
	return m.GetQuerySet().Defer(fields...)
}

// SelectRelated eager-loads foreign key relationships using JOINs.
func (m Manager[T]) SelectRelated(depth int) QuerySet[T] {
	return m.GetQuerySet().SelectRelated(depth)
}

// PrefetchRelated eager-loads ManyToMany and reverse FK relationships.
func (m Manager[T]) PrefetchRelated(lookups ...string) QuerySet[T] {
	return m.GetQuerySet().PrefetchRelated(lookups...)
}

// Using selects the database to run the query against.
func (m Manager[T]) Using(db string) QuerySet[T] {
	return m.GetQuerySet().Using(db)
}

// Annotate adds aggregate expressions to the query.
func (m Manager[T]) Annotate(exprs ...AggExpr) QuerySet[T] {
	return m.GetQuerySet().Annotate(exprs...)
}

// Extra injects raw SQL fragments into the query.
func (m Manager[T]) Extra(opts ExtraData) QuerySet[T] {
	return m.GetQuerySet().Extra(opts)
}

// ---------------------------------------------------------------------------
// QuerySet delegation — set operations (return new QuerySet[T])
// ---------------------------------------------------------------------------

// Union combines results with the given QuerySets using SQL UNION.
func (m Manager[T]) Union(others ...QuerySet[T]) QuerySet[T] {
	return m.GetQuerySet().Union(others...)
}

// UnionAll combines results using SQL UNION ALL.
func (m Manager[T]) UnionAll(others ...QuerySet[T]) QuerySet[T] {
	return m.GetQuerySet().UnionAll(others...)
}

// Intersection combines results using SQL INTERSECT.
func (m Manager[T]) Intersection(others ...QuerySet[T]) QuerySet[T] {
	return m.GetQuerySet().Intersection(others...)
}

// Difference combines results using SQL EXCEPT.
func (m Manager[T]) Difference(others ...QuerySet[T]) QuerySet[T] {
	return m.GetQuerySet().Difference(others...)
}

// ---------------------------------------------------------------------------
// QuerySet delegation — terminal methods (execute the query)
// ---------------------------------------------------------------------------

// All executes the query and returns all matched records.
func (m Manager[T]) All() ([]T, error) {
	return m.GetQuerySet().All()
}

// Get executes the query and returns exactly one record.
func (m Manager[T]) Get(lookups ...Lookup) (T, error) {
	return m.GetQuerySet().Get(lookups...)
}

// First returns the first object matched by the query.
func (m Manager[T]) First() (*T, error) {
	return m.GetQuerySet().First()
}

// Last returns the last object matched by the query.
func (m Manager[T]) Last() (*T, error) {
	return m.GetQuerySet().Last()
}

// Exists returns true if the QuerySet contains any results.
func (m Manager[T]) Exists() (bool, error) {
	return m.GetQuerySet().Exists()
}

// Count returns the number of records as an int64.
func (m Manager[T]) Count() (int64, error) {
	return m.GetQuerySet().Count()
}

// Values returns dictionaries (maps) rather than model instances.
func (m Manager[T]) Values(fields ...string) ([]map[string]any, error) {
	return m.GetQuerySet().Values(fields...)
}

// ValuesList returns tuples (slices) rather than model instances.
func (m Manager[T]) ValuesList(fields ...string) ([][]any, error) {
	return m.GetQuerySet().ValuesList(fields...)
}

// Create inserts a new record into the database.
func (m Manager[T]) Create(obj *T) error {
	return m.GetQuerySet().Create(obj)
}

// Update updates the matched records with the given fields.
func (m Manager[T]) Update(fields map[string]any) (int64, error) {
	return m.GetQuerySet().Update(fields)
}

// Delete deletes the matched records.
func (m Manager[T]) Delete() (int64, error) {
	return m.GetQuerySet().Delete()
}

// GetOrCreate gets an object, or creates it if it doesn't exist.
func (m Manager[T]) GetOrCreate(kwargs Lookup, defaults map[string]any) (T, bool, error) {
	return m.GetQuerySet().GetOrCreate(kwargs, defaults)
}

// UpdateOrCreate updates an object, or creates it if it doesn't exist.
func (m Manager[T]) UpdateOrCreate(kwargs Lookup, defaults map[string]any) (T, bool, error) {
	return m.GetQuerySet().UpdateOrCreate(kwargs, defaults)
}

// BulkCreate inserts multiple records efficiently.
func (m Manager[T]) BulkCreate(objs []T) error {
	return m.GetQuerySet().BulkCreate(objs)
}

// BulkUpdate updates multiple specific fields on a slice of objects.
func (m Manager[T]) BulkUpdate(objs []T, fields []string) error {
	return m.GetQuerySet().BulkUpdate(objs, fields)
}

// Aggregate executes the query and returns the requested aggregations.
func (m Manager[T]) Aggregate(exprs ...AggExpr) (map[string]any, error) {
	return m.GetQuerySet().Aggregate(exprs...)
}

// Dates returns a list of distinct date values truncated to the given kind.
func (m Manager[T]) Dates(field string, kind DateKind, order ...OrderDirection) ([]time.Time, error) {
	return m.GetQuerySet().Dates(field, kind, order...)
}

// Datetimes returns a list of distinct datetime values truncated to the given kind.
func (m Manager[T]) Datetimes(field string, kind DateTimeKind, order ...OrderDirection) ([]time.Time, error) {
	return m.GetQuerySet().Datetimes(field, kind, order...)
}

// Raw creates a raw SQL query.
func (m Manager[T]) Raw(sql string, params ...any) RawQuerySet[T] {
	return m.GetQuerySet().Raw(sql, params...)
}

// Iterator returns a range-over-func iterator that streams model instances.
func (m Manager[T]) Iterator() iter.Seq2[*T, error] {
	return m.GetQuerySet().Iterator()
}

// IteratorChunked is like Iterator but with a custom chunk size.
func (m Manager[T]) IteratorChunked(chunkSize int) iter.Seq2[*T, error] {
	return m.GetQuerySet().IteratorChunked(chunkSize)
}
