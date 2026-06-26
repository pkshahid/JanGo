package queryset

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkshahid/JanGo/orm"
)

// SetOperation represents a SQL set operation type.
type SetOperation string

const (
	SetOpNone         SetOperation = ""
	SetOpUnion        SetOperation = "UNION"
	SetOpIntersection SetOperation = "INTERSECT"
	SetOpDifference   SetOperation = "EXCEPT"
)

// ExtraData holds raw SQL fragments injected via QuerySet.Extra().
// This is an escape hatch for cases the ORM cannot express natively.
//
// Deprecated: prefer Annotate, F expressions, or Func-based expressions
// whenever possible. Extra is retained for parity with Django and for
// one-off raw fragments that don't warrant a full expression node.
type ExtraData struct {
	Select  map[string]string // alias -> raw SQL expression appended to SELECT
	Where   []string          // raw SQL fragments AND-ed into the WHERE clause
	Params  []any             // bind parameters for Where fragments
	Tables  []string          // extra table names appended to the FROM clause
	OrderBy []string          // raw ORDER BY fragments appended after ORM-generated ones
}

// Query holds the AST for a SQL query.
type Query struct {
	ModelInfo       *orm.ModelInfo
	Where           *QNode
	Exclude         *QNode
	Limit           int
	Offset          int
	OrderBy         []string
	SelectRelated   int
	PrefetchRelated []string
	Annotations     []AggExpr
	OnlyFields      []string
	DeferFields     []string
	Distinct        bool
	Database        string
	SetOp           SetOperation
	SetQueries      []*Query
	SetAll          bool
	Extra           *ExtraData
}

// NewQuery creates a new query object.
func NewQuery(modelInfo *orm.ModelInfo) *Query {
	return &Query{
		ModelInfo: modelInfo,
		Limit:     -1,
		Offset:    0,
	}
}

// clone creates a deep copy of the Query.
func (q *Query) clone() *Query {
	c := &Query{
		ModelInfo:     q.ModelInfo,
		Limit:         q.Limit,
		Offset:        q.Offset,
		SelectRelated: q.SelectRelated,
		Distinct:      q.Distinct,
		Database:      q.Database,
	}

	if q.Where != nil {
		c.Where = q.Where.clone()
	}
	if q.Exclude != nil {
		c.Exclude = q.Exclude.clone()
	}
	c.OrderBy = append([]string(nil), q.OrderBy...)
	c.PrefetchRelated = append([]string(nil), q.PrefetchRelated...)
	c.Annotations = append([]AggExpr(nil), q.Annotations...)
	c.OnlyFields = append([]string(nil), q.OnlyFields...)
	c.DeferFields = append([]string(nil), q.DeferFields...)
	c.SetOp = q.SetOp
	c.SetAll = q.SetAll
	for _, sq := range q.SetQueries {
		c.SetQueries = append(c.SetQueries, sq.clone())
	}

	if q.Extra != nil {
		c.Extra = &ExtraData{
			Where:   append([]string(nil), q.Extra.Where...),
			Params:  append([]any(nil), q.Extra.Params...),
			Tables:  append([]string(nil), q.Extra.Tables...),
			OrderBy: append([]string(nil), q.Extra.OrderBy...),
		}
		if q.Extra.Select != nil {
			c.Extra.Select = make(map[string]string, len(q.Extra.Select))
			for k, v := range q.Extra.Select {
				c.Extra.Select[k] = v
			}
		}
	}

	return c
}

// ToSQL generates the SQL string and parameters for the query.
func (q *Query) ToSQL() (string, []any) {
	if q.SetOp != SetOpNone {
		return q.toSetOpSQL()
	}

	sql, params := q.toCoreSelectSQL()
	sql += q.toOrderBySQL()

	if q.Limit > -1 {
		sql += fmt.Sprintf(" LIMIT %d", q.Limit)
	}

	if q.Offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", q.Offset)
	}

	return sql, params
}

// toCoreSelectSQL generates the SELECT...FROM...WHERE portion of the query
// without ORDER BY, LIMIT, or OFFSET. Used by both ToSQL and set operation queries.
func (q *Query) toCoreSelectSQL() (string, []any) {
	tableName := q.ModelInfo.Meta.DbTable
	selectFields := "*"
	if len(q.OnlyFields) > 0 {
		selectFields = strings.Join(q.OnlyFields, ", ")
	}

	// Append extra select expressions (alias -> raw SQL).
	if q.Extra != nil && len(q.Extra.Select) > 0 {
		var aliases []string
		for alias := range q.Extra.Select {
			aliases = append(aliases, alias)
		}
		// Sort by alias for deterministic output (map iteration is randomised in Go).
		sort.Strings(aliases)
		var extraSelects []string
		for _, alias := range aliases {
			extraSelects = append(extraSelects, fmt.Sprintf("(%s) AS %s", q.Extra.Select[alias], alias))
		}
		selectFields = selectFields + ", " + strings.Join(extraSelects, ", ")
	}

	fromClause := tableName
	if q.Extra != nil && len(q.Extra.Tables) > 0 {
		fromClause = fromClause + ", " + strings.Join(q.Extra.Tables, ", ")
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", selectFields, fromClause)
	params := []any{}

	var whereClauses []string

	if q.Where != nil {
		clause, p := q.Where.toSQL(q.ModelInfo)
		if clause != "" {
			whereClauses = append(whereClauses, "("+clause+")")
			params = append(params, p...)
		}
	}

	if q.Exclude != nil {
		clause, p := q.Exclude.toSQL(q.ModelInfo)
		if clause != "" {
			whereClauses = append(whereClauses, "NOT ("+clause+")")
			params = append(params, p...)
		}
	}

	if q.Extra != nil && len(q.Extra.Where) > 0 {
		whereClauses = append(whereClauses, q.Extra.Where...)
		params = append(params, q.Extra.Params...)
	}

	if len(whereClauses) > 0 {
		sql += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	return sql, params
}

// toOrderBySQL generates the ORDER BY clause string (including leading space) or empty string.
func (q *Query) toOrderBySQL() string {
	var orderStrs []string
	for _, order := range q.OrderBy {
		desc := false
		field := order
		if strings.HasPrefix(order, "-") {
			desc = true
			field = order[1:]
		}

		colName := field
		if f, ok := q.ModelInfo.FieldByName[field]; ok {
			colName = f.Column
		}

		if desc {
			orderStrs = append(orderStrs, colName+" DESC")
		} else {
			orderStrs = append(orderStrs, colName+" ASC")
		}
	}

	// Append raw extra order-by fragments.
	if q.Extra != nil {
		orderStrs = append(orderStrs, q.Extra.OrderBy...)
	}

	if len(orderStrs) == 0 {
		return ""
	}
	return " ORDER BY " + strings.Join(orderStrs, ", ")
}

// toSetOpSQL generates SQL for set operation queries (UNION, INTERSECT, EXCEPT).
// Each sub-query is wrapped in parentheses. ORDER BY, LIMIT, and OFFSET
// from the outer query are applied to the combined result.
func (q *Query) toSetOpSQL() (string, []any) {
	var parts []string
	var params []any

	// First operand: this query's core SELECT (without ORDER BY/LIMIT)
	coreSQL, coreParams := q.toCoreSelectSQL()
	parts = append(parts, "("+coreSQL+")")
	params = append(params, coreParams...)

	op := string(q.SetOp)
	if q.SetAll && q.SetOp == SetOpUnion {
		op = "UNION ALL"
	}

	for _, sub := range q.SetQueries {
		subSQL, subParams := sub.toCoreSelectSQL()
		parts = append(parts, "("+subSQL+")")
		params = append(params, subParams...)
	}

	sql := strings.Join(parts, " "+op+" ")

	// Apply ORDER BY, LIMIT, OFFSET to the combined result
	sql += q.toOrderBySQL()

	if q.Limit > -1 {
		sql += fmt.Sprintf(" LIMIT %d", q.Limit)
	}

	if q.Offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", q.Offset)
	}

	return sql, params
}

// toSQL converts a QNode tree into a SQL WHERE clause.
func (q *QNode) toSQL(info *orm.ModelInfo) (string, []any) {
	if q == nil {
		return "", nil
	}

	var clauses []string
	var params []any

	// Process internal lookups
	for k, v := range q.Filters {
		clause, p := parseLookup(k, v, info)
		clauses = append(clauses, clause)
		params = append(params, p...)
	}

	// Process children
	for _, child := range q.Children {
		childClause, childParams := child.toSQL(info)
		if childClause != "" {
			clauses = append(clauses, "("+childClause+")")
			params = append(params, childParams...)
		}
	}

	if len(clauses) == 0 {
		return "", nil
	}

	connector := " AND "
	if q.Connector == OR {
		connector = " OR "
	}

	sql := strings.Join(clauses, connector)
	if q.Negated {
		sql = "NOT (" + sql + ")"
	}

	return sql, params
}

// resolveVal renders a lookup value to its SQL placeholder and bind params.
// If the value implements Expression (e.g. F), it is resolved to a SQL fragment
// referencing database columns instead of a bind parameter.
func resolveVal(val any, info *orm.ModelInfo) (string, []any) {
	if expr, ok := val.(Expression); ok {
		return expr.ResolveSQL(info)
	}
	return "?", []any{val}
}

// parseLookup converts a Django-style lookup (e.g., "title__icontains") into a SQL fragment.
func parseLookup(key string, val any, info *orm.ModelInfo) (string, []any) {
	parts := strings.Split(key, "__")
	field := parts[0]
	lookup := "exact"

	if len(parts) > 1 {
		lookup = parts[len(parts)-1]
		// In a real ORM, we'd handle JOINs for multiple `__` (e.g., `author__user__name__icontains`)
		// Here, we just take the last part as the lookup and the rest as the field path.
		field = strings.Join(parts[:len(parts)-1], ".")
	}

	// Map field name to column name
	colName := field // Default to field name
	if info != nil {
		if f, ok := info.FieldByName[field]; ok {
			colName = f.Column
		}
	}

	// Simplistic lookup mapping
	switch lookup {
	case "exact":
		ph, p := resolveVal(val, info)
		return fmt.Sprintf("%s = %s", colName, ph), p
	case "iexact":
		return fmt.Sprintf("LOWER(%s) = LOWER(?)", colName), []any{val}
	case "contains":
		return fmt.Sprintf("%s LIKE ?", colName), []any{fmt.Sprintf("%%%v%%", val)}
	case "icontains":
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", colName), []any{fmt.Sprintf("%%%v%%", val)}
	case "startswith":
		return fmt.Sprintf("%s LIKE ?", colName), []any{fmt.Sprintf("%v%%", val)}
	case "istartswith":
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", colName), []any{fmt.Sprintf("%v%%", val)}
	case "endswith":
		return fmt.Sprintf("%s LIKE ?", colName), []any{fmt.Sprintf("%%%v", val)}
	case "iendswith":
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", colName), []any{fmt.Sprintf("%%%v", val)}
	case "gt":
		ph, p := resolveVal(val, info)
		return fmt.Sprintf("%s > %s", colName, ph), p
	case "gte":
		ph, p := resolveVal(val, info)
		return fmt.Sprintf("%s >= %s", colName, ph), p
	case "lt":
		ph, p := resolveVal(val, info)
		return fmt.Sprintf("%s < %s", colName, ph), p
	case "lte":
		ph, p := resolveVal(val, info)
		return fmt.Sprintf("%s <= %s", colName, ph), p
	case "in":
		// Handle slice/array
		// A real implementation expands `?` based on length
		return fmt.Sprintf("%s IN (?)", colName), []any{val}
	case "isnull":
		if b, ok := val.(bool); ok && b {
			return fmt.Sprintf("%s IS NULL", colName), nil
		}
		return fmt.Sprintf("%s IS NOT NULL", colName), nil
	default:
		// Fallback to exact if unknown
		ph, p := resolveVal(val, info)
		return fmt.Sprintf("%s = %s", colName, ph), p
	}
}

// AggExpr represents an aggregate expression.
type AggExpr interface {
	ToSQL() string
}

// ToUpdateSQL generates an UPDATE statement for the given field-value pairs.
// Values that implement Expression (e.g. F) are resolved to SQL column references.
func (q *Query) ToUpdateSQL(fields map[string]any) (string, []any) {
	tableName := q.ModelInfo.Meta.DbTable

	var setClauses []string
	var params []any

	for name, val := range fields {
		colName := name
		if f, ok := q.ModelInfo.FieldByName[name]; ok {
			colName = f.Column
		}

		if expr, ok := val.(Expression); ok {
			exprSQL, exprParams := expr.ResolveSQL(q.ModelInfo)
			setClauses = append(setClauses, fmt.Sprintf("%s = %s", colName, exprSQL))
			params = append(params, exprParams...)
		} else {
			setClauses = append(setClauses, fmt.Sprintf("%s = ?", colName))
			params = append(params, val)
		}
	}

	sql := fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(setClauses, ", "))

	var whereClauses []string

	if q.Where != nil {
		clause, p := q.Where.toSQL(q.ModelInfo)
		if clause != "" {
			whereClauses = append(whereClauses, "("+clause+")")
			params = append(params, p...)
		}
	}

	if q.Exclude != nil {
		clause, p := q.Exclude.toSQL(q.ModelInfo)
		if clause != "" {
			whereClauses = append(whereClauses, "NOT ("+clause+")")
			params = append(params, p...)
		}
	}

	if len(whereClauses) > 0 {
		sql += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	return sql, params
}
