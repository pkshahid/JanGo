package queryset

import (
	"fmt"

	"github.com/pkshahid/JanGo/orm"
)

// QuerySet represents a lazy database query for a specific model type T.
type QuerySet[T any] struct {
	query *Query
}

// NewQuerySet initializes a new QuerySet for type T.
func NewQuerySet[T any]() QuerySet[T] {
	var m T
	info, err := orm.GetModelInfo(m)
	if err != nil {
		// In a real framework, this might panic or return an error during initialization.
		// For chaining ergonomics, we store the error in the query or panic if the model isn't registered.
		panic(fmt.Sprintf("orm: failed to get model info for %T: %v", m, err))
	}
	return QuerySet[T]{
		query: NewQuery(info),
	}
}

// Filter adds an AND WHERE clause to the queryset.
func (qs QuerySet[T]) Filter(lookups ...Lookup) QuerySet[T] {
	c := qs.clone()
	for _, l := range lookups {
		c.query.Where = c.query.Where.And(Q(l))
	}
	return c
}

// FilterQ adds a Q object to the WHERE clause using AND.
func (qs QuerySet[T]) FilterQ(q *QNode) QuerySet[T] {
	c := qs.clone()
	c.query.Where = c.query.Where.And(q)
	return c
}

// Exclude adds a NOT WHERE clause to the queryset.
func (qs QuerySet[T]) Exclude(lookups ...Lookup) QuerySet[T] {
	c := qs.clone()
	for _, l := range lookups {
		c.query.Exclude = c.query.Exclude.And(Q(l))
	}
	return c
}

// ExcludeQ adds a Q object to the Exclude clause using AND.
func (qs QuerySet[T]) ExcludeQ(q *QNode) QuerySet[T] {
	c := qs.clone()
	c.query.Exclude = c.query.Exclude.And(q)
	return c
}

// OrderBy orders the queryset by the specified fields.
// Prefix with '-' for descending order.
func (qs QuerySet[T]) OrderBy(fields ...string) QuerySet[T] {
	c := qs.clone()
	c.query.OrderBy = fields
	return c
}

// Limit limits the number of results returned.
func (qs QuerySet[T]) Limit(n int) QuerySet[T] {
	c := qs.clone()
	c.query.Limit = n
	return c
}

// Offset sets the offset for the query results.
func (qs QuerySet[T]) Offset(n int) QuerySet[T] {
	c := qs.clone()
	c.query.Offset = n
	return c
}

// Distinct ensures the results are distinct.
func (qs QuerySet[T]) Distinct() QuerySet[T] {
	c := qs.clone()
	c.query.Distinct = true
	return c
}

// Only restricts the fetched columns to the specified fields.
func (qs QuerySet[T]) Only(fields ...string) QuerySet[T] {
	c := qs.clone()
	c.query.OnlyFields = append(c.query.OnlyFields, fields...)
	return c
}

// Defer prevents the specified fields from being fetched initially.
func (qs QuerySet[T]) Defer(fields ...string) QuerySet[T] {
	c := qs.clone()
	c.query.DeferFields = append(c.query.DeferFields, fields...)
	return c
}

// SelectRelated eager-loads foreign key relationships using JOINs.
func (qs QuerySet[T]) SelectRelated(depth int) QuerySet[T] {
	c := qs.clone()
	c.query.SelectRelated = depth
	return c
}

// PrefetchRelated eager-loads ManyToMany and reverse ForeignKey relationships using separate queries.
func (qs QuerySet[T]) PrefetchRelated(lookups ...string) QuerySet[T] {
	c := qs.clone()
	c.query.PrefetchRelated = append(c.query.PrefetchRelated, lookups...)
	return c
}

// Using selects the database to run the query against.
func (qs QuerySet[T]) Using(db string) QuerySet[T] {
	c := qs.clone()
	c.query.Database = db
	return c
}

// Annotate adds aggregate expressions to the query.
func (qs QuerySet[T]) Annotate(exprs ...AggExpr) QuerySet[T] {
	c := qs.clone()
	c.query.Annotations = append(c.query.Annotations, exprs...)
	return c
}

// clone creates a deep copy of the QuerySet, ensuring immutability of the chain.
func (qs QuerySet[T]) clone() QuerySet[T] {
	return QuerySet[T]{
		query: qs.query.clone(),
	}
}
