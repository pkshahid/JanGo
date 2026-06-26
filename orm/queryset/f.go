package queryset

import (
	"fmt"

	"github.com/pkshahid/JanGo/orm"
)

// Expression is the interface for SQL expressions that can be resolved
// against a ModelInfo to produce a SQL fragment and bind parameters.
type Expression interface {
	ResolveSQL(info *orm.ModelInfo) (string, []any)
}

// F represents a reference to a model field value within a query expression.
// It allows referencing database columns directly in filters, updates, and
// annotations without pulling values into application memory.
//
// Usage:
//
//	qs.Filter(Lookup{"Views": F("Likes")})           // WHERE views = likes
//	qs.Update(map[string]any{"Views": F("Views").Add(1)})  // SET views = views + 1
//	qs.Filter(Lookup{"Views__gt": F("Likes").Add(10)})     // WHERE views > likes + 10
type F struct {
	field string // field name for leaf nodes
	op    string // arithmetic operator (empty for leaf)
	left  *F     // left operand for combined expressions
	right any    // right operand: *F or literal value
}

// NewF creates a new F expression referencing the given field name.
func NewF(field string) *F {
	return &F{field: field}
}

// F is a convenience constructor alias for NewF.
// It mirrors Django's F("field") syntax.
func FRef(field string) *F {
	return &F{field: field}
}

// Add returns a new F expression computing left + right.
func (f *F) Add(other any) *F {
	return &F{op: "+", left: f, right: other}
}

// Sub returns a new F expression computing left - right.
func (f *F) Sub(other any) *F {
	return &F{op: "-", left: f, right: other}
}

// Mul returns a new F expression computing left * right.
func (f *F) Mul(other any) *F {
	return &F{op: "*", left: f, right: other}
}

// Div returns a new F expression computing left / right.
func (f *F) Div(other any) *F {
	return &F{op: "/", left: f, right: other}
}

// Mod returns a new F expression computing left % right.
func (f *F) Mod(other any) *F {
	return &F{op: "%", left: f, right: other}
}

// resolveRight renders the right operand to SQL.
func (f *F) resolveRight(info *orm.ModelInfo) (string, []any) {
	if rightF, ok := f.right.(*F); ok {
		return rightF.ResolveSQL(info)
	}
	return "?", []any{f.right}
}

// ResolveSQL renders the F expression to a SQL fragment with bind parameters.
func (f *F) ResolveSQL(info *orm.ModelInfo) (string, []any) {
	if f.op == "" {
		colName := f.field
		if info != nil {
			if field, ok := info.FieldByName[f.field]; ok {
				colName = field.Column
			}
		}
		return colName, nil
	}

	leftSQL, leftParams := f.left.ResolveSQL(info)
	rightSQL, rightParams := f.resolveRight(info)

	sql := fmt.Sprintf("(%s %s %s)", leftSQL, f.op, rightSQL)
	return sql, append(leftParams, rightParams...)
}

// ToSQL renders the F expression to a SQL string without column resolution.
// This satisfies the AggExpr interface so F can be used in Annotate.
func (f *F) ToSQL() string {
	sql, _ := f.ResolveSQL(nil)
	return sql
}
