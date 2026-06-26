package queryset

// Union returns a QuerySet that combines results with the given QuerySets
// using SQL UNION (deduplicates rows).
//
//	qs1.Union(qs2, qs3)
//
// The resulting QuerySet is immutable. ORDER BY, LIMIT, and OFFSET applied
// to the result affect the combined set, not the individual sub-queries.
func (qs QuerySet[T]) Union(others ...QuerySet[T]) QuerySet[T] {
	c := qs.clone()
	c.query.SetOp = SetOpUnion
	c.query.SetAll = false
	for _, o := range others {
		c.query.SetQueries = append(c.query.SetQueries, o.query.clone())
	}
	return c
}

// UnionAll returns a QuerySet that combines results with the given QuerySets
// using SQL UNION ALL (preserves duplicate rows).
//
//	qs1.UnionAll(qs2, qs3)
func (qs QuerySet[T]) UnionAll(others ...QuerySet[T]) QuerySet[T] {
	c := qs.clone()
	c.query.SetOp = SetOpUnion
	c.query.SetAll = true
	for _, o := range others {
		c.query.SetQueries = append(c.query.SetQueries, o.query.clone())
	}
	return c
}

// Intersection returns a QuerySet that combines results with the given
// QuerySets using SQL INTERSECT (returns rows present in all QuerySets).
//
//	qs1.Intersection(qs2, qs3)
func (qs QuerySet[T]) Intersection(others ...QuerySet[T]) QuerySet[T] {
	c := qs.clone()
	c.query.SetOp = SetOpIntersection
	for _, o := range others {
		c.query.SetQueries = append(c.query.SetQueries, o.query.clone())
	}
	return c
}

// Difference returns a QuerySet that combines results with the given
// QuerySets using SQL EXCEPT (returns rows in the first QuerySet that are
// not in any of the others).
//
//	qs1.Difference(qs2, qs3)
func (qs QuerySet[T]) Difference(others ...QuerySet[T]) QuerySet[T] {
	c := qs.clone()
	c.query.SetOp = SetOpDifference
	for _, o := range others {
		c.query.SetQueries = append(c.query.SetQueries, o.query.clone())
	}
	return c
}
