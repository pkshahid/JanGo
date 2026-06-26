package queryset

import (
	"fmt"
	"strings"

	"github.com/pkshahid/JanGo/orm"
)

// When represents a single CASE WHEN ... THEN ... clause.
// The Condition is a QNode tree that produces a SQL boolean expression,
// and Then is the result value (literal or Expression) returned when the
// condition is true.
//
// Usage:
//
//	w := NewWhen(Q(Lookup{"Age__gte": 18}), "adult")
//	w2 := NewWhen(Q(Lookup{"Age__lt": 18}), FRef("Discount"))
type When struct {
	Condition *QNode
	Then      any
}

// NewWhen creates a When clause with the given condition and result.
func NewWhen(condition *QNode, then any) *When {
	return &When{
		Condition: condition,
		Then:      then,
	}
}

// resolveThen renders the THEN value to SQL, handling both Expression
// and literal values.
func (w *When) resolveThen(info *orm.ModelInfo) (string, []any) {
	if expr, ok := w.Then.(Expression); ok {
		return expr.ResolveSQL(info)
	}
	return "?", []any{w.Then}
}

// ResolveSQL renders this When clause as a SQL fragment for use within a
// CASE expression. It does NOT include the CASE keyword itself.
func (w *When) ResolveSQL(info *orm.ModelInfo) (string, []any) {
	condSQL, condParams := w.Condition.toSQL(info)
	thenSQL, thenParams := w.resolveThen(info)

	sql := fmt.Sprintf("WHEN %s THEN %s", condSQL, thenSQL)
	return sql, append(condParams, thenParams...)
}

// ToSQL renders the When clause without column resolution (for AggExpr).
func (w *When) ToSQL() string {
	sql, _ := w.ResolveSQL(nil)
	return sql
}

// Case represents a SQL CASE expression composed of one or more When
// clauses and an optional default (ELSE) value.
//
// Usage:
//
//	c := NewCase(
//	    NewWhen(Q(Lookup{"Age__gte": 18}), "adult"),
//	    NewWhen(Q(Lookup{"Age__lt": 18}), "minor"),
//	).Default("unknown")
//
//	qs = qs.Annotate(c)                     // use in annotation
//	qs = qs.Filter(Lookup{"Category": c})   // use in filter
//	qs = qs.Update(map[string]any{"Label": c}) // use in update
type Case struct {
	cases   []*When
	def     any
	hasDef  bool
	output  string // optional output field type hint (e.g., "TEXT", "INTEGER")
	hasOut  bool
}

// NewCase creates a Case expression from one or more When clauses.
func NewCase(whens ...*When) *Case {
	return &Case{cases: whens}
}

// Default sets the ELSE value for the Case expression.
func (c *Case) Default(value any) *Case {
	clone := c.clone()
	clone.def = value
	clone.hasDef = true
	return clone
}

// OutputField sets the type hint for the Case expression's result.
// This is informational and may be used by backends that require
// explicit casting (e.g., PostgreSQL).
func (c *Case) OutputField(typeName string) *Case {
	clone := c.clone()
	clone.output = typeName
	clone.hasOut = true
	return clone
}

// clone returns a shallow copy of the Case expression.
func (c *Case) clone() *Case {
	cp := &Case{
		def:    c.def,
		hasDef: c.hasDef,
		output: c.output,
		hasOut: c.hasOut,
	}
	cp.cases = append([]*When(nil), c.cases...)
	return cp
}

// resolveDefault renders the default (ELSE) value to SQL.
func (c *Case) resolveDefault(info *orm.ModelInfo) (string, []any) {
	if expr, ok := c.def.(Expression); ok {
		return expr.ResolveSQL(info)
	}
	return "?", []any{c.def}
}

// ResolveSQL renders the full CASE expression to a SQL fragment with bind
// parameters, satisfying the Expression interface.
func (c *Case) ResolveSQL(info *orm.ModelInfo) (string, []any) {
	var parts []string
	var params []any

	for _, w := range c.cases {
		wSQL, wParams := w.ResolveSQL(info)
		parts = append(parts, wSQL)
		params = append(params, wParams...)
	}

	if c.hasDef {
		defSQL, defParams := c.resolveDefault(info)
		parts = append(parts, "ELSE "+defSQL)
		params = append(params, defParams...)
	}

	sql := "CASE " + strings.Join(parts, " ") + " END"

	if c.hasOut && c.output != "" {
		sql = fmt.Sprintf("CAST(%s AS %s)", sql, c.output)
	}

	return sql, params
}

// ToSQL renders the Case expression without column resolution,
// satisfying the AggExpr interface for use in Annotate.
func (c *Case) ToSQL() string {
	sql, _ := c.ResolveSQL(nil)
	return sql
}
