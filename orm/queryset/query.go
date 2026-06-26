package queryset

import (
	"fmt"
	"reflect"
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

	sql, params, jm := q.toCoreSelectSQL()
	sql += q.toOrderBySQL(jm)

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
// Returns the SQL, bind params, and the joinManager (for use by toOrderBySQL).
func (q *Query) toCoreSelectSQL() (string, []any, *joinManager) {
	jm := newJoinManager(q.ModelInfo.Meta.DbTable)

	// Process SelectRelated → add joins and collect related columns.
	var relatedCols []selectCol
	if q.SelectRelated > 0 {
		relatedCols = jm.collectSelectRelated(q.ModelInfo, jm.baseAlias, "", q.SelectRelated)
	}

	// Build WHERE clauses (may add joins for __ path lookups).
	var whereClauses []string
	params := []any{}

	if q.Where != nil {
		clause, p := q.Where.toSQL(q.ModelInfo, jm)
		if clause != "" {
			whereClauses = append(whereClauses, "("+clause+")")
			params = append(params, p...)
		}
	}

	if q.Exclude != nil {
		clause, p := q.Exclude.toSQL(q.ModelInfo, jm)
		if clause != "" {
			whereClauses = append(whereClauses, "NOT ("+clause+")")
			params = append(params, p...)
		}
	}

	if q.Extra != nil && len(q.Extra.Where) > 0 {
		whereClauses = append(whereClauses, q.Extra.Where...)
		params = append(params, q.Extra.Params...)
	}

	// Pre-resolve ORDER BY fields so their joins are included before
	// the FROM clause is generated.
	for _, order := range q.OrderBy {
		field := order
		if strings.HasPrefix(field, "-") {
			field = field[1:]
		}
		jm.resolveOrderByField(field, q.ModelInfo)
	}

	// Build SELECT clause.
	selectClause := q.buildSelectClause(jm, relatedCols)

	// Build FROM clause with JOINs.
	fromClause := jm.fromClause()
	if q.Extra != nil && len(q.Extra.Tables) > 0 {
		fromClause = fromClause + ", " + strings.Join(q.Extra.Tables, ", ")
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", selectClause, fromClause)

	if len(whereClauses) > 0 {
		sql += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	return sql, params, jm
}

// buildSelectClause generates the SELECT column list, qualifying columns
// with table aliases when JOINs are present.
func (q *Query) buildSelectClause(jm *joinManager, relatedCols []selectCol) string {
	hasJoins := jm.hasJoins()
	var parts []string

	// Base table columns.
	if len(q.OnlyFields) > 0 {
		for _, fieldName := range q.OnlyFields {
			if hasJoins {
				colName := fieldName
				if f, ok := q.ModelInfo.FieldByName[fieldName]; ok {
					colName = f.Column
				}
				parts = append(parts, jm.baseAlias+"."+colName)
			} else {
				parts = append(parts, fieldName)
			}
		}
	} else {
		if hasJoins {
			parts = append(parts, jm.baseAlias+".*")
		} else {
			parts = append(parts, "*")
		}
	}

	// Related table columns from SelectRelated.
	for _, rc := range relatedCols {
		parts = append(parts, rc.expr)
	}

	// Extra select expressions (alias -> raw SQL).
	if q.Extra != nil && len(q.Extra.Select) > 0 {
		var aliases []string
		for alias := range q.Extra.Select {
			aliases = append(aliases, alias)
		}
		sort.Strings(aliases)
		for _, alias := range aliases {
			parts = append(parts, fmt.Sprintf("(%s) AS %s", q.Extra.Select[alias], alias))
		}
	}

	return strings.Join(parts, ", ")
}

// toOrderBySQL generates the ORDER BY clause string (including leading space) or empty string.
// When jm is non-nil, order fields are resolved through the join manager so that
// related-field ordering (e.g. "author__name") uses the correct table alias.
func (q *Query) toOrderBySQL(jm *joinManager) string {
	var orderStrs []string
	for _, order := range q.OrderBy {
		desc := false
		field := order
		if strings.HasPrefix(order, "-") {
			desc = true
			field = order[1:]
		}

		var colRef string
		if jm != nil {
			colRef = jm.resolveOrderByField(field, q.ModelInfo)
		} else {
			colName := field
			if f, ok := q.ModelInfo.FieldByName[field]; ok {
				colName = f.Column
			} else if field == "pk" && q.ModelInfo.PrimaryKey != nil {
				colName = q.ModelInfo.PrimaryKey.Column
			}
			colRef = colName
		}

		if desc {
			orderStrs = append(orderStrs, colRef+" DESC")
		} else {
			orderStrs = append(orderStrs, colRef+" ASC")
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
	coreSQL, coreParams, _ := q.toCoreSelectSQL()
	parts = append(parts, "("+coreSQL+")")
	params = append(params, coreParams...)

	op := string(q.SetOp)
	if q.SetAll && q.SetOp == SetOpUnion {
		op = "UNION ALL"
	}

	for _, sub := range q.SetQueries {
		subSQL, subParams, _ := sub.toCoreSelectSQL()
		parts = append(parts, "("+subSQL+")")
		params = append(params, subParams...)
	}

	sql := strings.Join(parts, " "+op+" ")

	// Apply ORDER BY, LIMIT, OFFSET to the combined result
	sql += q.toOrderBySQL(nil)

	if q.Limit > -1 {
		sql += fmt.Sprintf(" LIMIT %d", q.Limit)
	}

	if q.Offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", q.Offset)
	}

	return sql, params
}

// toSQL converts a QNode tree into a SQL WHERE clause.
// When jm is non-nil, __-separated field paths are resolved through the
// join manager, creating JOINs for ForeignKey traversals as needed.
func (q *QNode) toSQL(info *orm.ModelInfo, jm *joinManager) (string, []any) {
	if q == nil {
		return "", nil
	}

	var clauses []string
	var params []any

	// Process internal lookups
	for k, v := range q.Filters {
		clause, p := parseLookup(k, v, info, jm)
		clauses = append(clauses, clause)
		params = append(params, p...)
	}

	// Process children
	for _, child := range q.Children {
		childClause, childParams := child.toSQL(info, jm)
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

// parseLookup converts a Django-style lookup (e.g., "title__icontains" or
// "author__name__icontains") into a SQL fragment. When jm is non-nil,
// __-separated field paths are resolved through the join manager, creating
// JOINs for ForeignKey traversals as needed.
func parseLookup(key string, val any, info *orm.ModelInfo, jm *joinManager) (string, []any) {
	parts := strings.Split(key, "__")

	var colRef string
	var lookup string

	if jm != nil {
		colRef, lookup = jm.resolveFieldPath(parts, info)
	} else {
		// Fallback for nil jm (UPDATE, DELETE, Case expressions).
		field := parts[0]
		lookup = "exact"
		if len(parts) > 1 && knownLookupTypes[parts[len(parts)-1]] {
			lookup = parts[len(parts)-1]
			fieldParts := parts[:len(parts)-1]
			if len(fieldParts) == 1 {
				field = fieldParts[0]
			} else {
				field = strings.Join(fieldParts, ".")
			}
		}
		colRef = field
		if info != nil {
			if f, ok := info.FieldByName[field]; ok {
				colRef = f.Column
			} else if field == "pk" && info.PrimaryKey != nil {
				colRef = info.PrimaryKey.Column
			}
		}
	}

	// Extract PK from model instance values for direct FK lookups.
	val = extractPKIfNeeded(val, info, parts)

	switch lookup {
	case "exact":
		ph, p := resolveVal(val, info)
		return fmt.Sprintf("%s = %s", colRef, ph), p
	case "iexact":
		return fmt.Sprintf("LOWER(%s) = LOWER(?)", colRef), []any{val}
	case "contains":
		if _, ok := val.(orm.GeometryValue); ok {
			geom, srid := extractGeometry(val)
			return fmt.Sprintf("ST_Contains(%s, ST_GeomFromText(?, %d))", colRef, srid), []any{geom}
		}
		return fmt.Sprintf("%s LIKE ?", colRef), []any{fmt.Sprintf("%%%v%%", val)}
	case "icontains":
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", colRef), []any{fmt.Sprintf("%%%v%%", val)}
	case "startswith":
		return fmt.Sprintf("%s LIKE ?", colRef), []any{fmt.Sprintf("%v%%", val)}
	case "istartswith":
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", colRef), []any{fmt.Sprintf("%v%%", val)}
	case "endswith":
		return fmt.Sprintf("%s LIKE ?", colRef), []any{fmt.Sprintf("%%%v", val)}
	case "iendswith":
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", colRef), []any{fmt.Sprintf("%%%v", val)}
	case "gt":
		ph, p := resolveVal(val, info)
		return fmt.Sprintf("%s > %s", colRef, ph), p
	case "gte":
		ph, p := resolveVal(val, info)
		return fmt.Sprintf("%s >= %s", colRef, ph), p
	case "lt":
		ph, p := resolveVal(val, info)
		return fmt.Sprintf("%s < %s", colRef, ph), p
	case "lte":
		ph, p := resolveVal(val, info)
		return fmt.Sprintf("%s <= %s", colRef, ph), p
	case "in":
		return expandInLookup(colRef, val)
	case "isnull":
		if b, ok := val.(bool); ok && b {
			return fmt.Sprintf("%s IS NULL", colRef), nil
		}
		return fmt.Sprintf("%s IS NOT NULL", colRef), nil
	case "range":
		return expandRangeLookup(colRef, val)
	// ── Spatial lookups (PostGIS / SpatiaLite) ──
	case "intersects":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("%s && ST_GeomFromText(?, %d)", colRef, srid), []any{geom}
	case "equals":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("ST_Equals(%s, ST_GeomFromText(?, %d))", colRef, srid), []any{geom}
	case "disjoint":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("ST_Disjoint(%s, ST_GeomFromText(?, %d))", colRef, srid), []any{geom}
	case "touches":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("ST_Touches(%s, ST_GeomFromText(?, %d))", colRef, srid), []any{geom}
	case "crosses":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("ST_Crosses(%s, ST_GeomFromText(?, %d))", colRef, srid), []any{geom}
	case "within":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("ST_Within(%s, ST_GeomFromText(?, %d))", colRef, srid), []any{geom}
	case "overlaps":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("ST_Overlaps(%s, ST_GeomFromText(?, %d))", colRef, srid), []any{geom}
	case "contains_properly":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("ST_ContainsProperly(%s, ST_GeomFromText(?, %d))", colRef, srid), []any{geom}
	case "covers":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("ST_Covers(%s, ST_GeomFromText(?, %d))", colRef, srid), []any{geom}
	case "coveredby":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("ST_CoveredBy(%s, ST_GeomFromText(?, %d))", colRef, srid), []any{geom}
	case "same_as":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("ST_OrderingEquals(%s, ST_GeomFromText(?, %d))", colRef, srid), []any{geom}
	case "relate":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("ST_Relate(%s, ST_GeomFromText(?, %d))", colRef, srid), []any{geom}
	case "bbcontains":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("%s @ ST_GeomFromText(?, %d)", colRef, srid), []any{geom}
	case "bboverlaps":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("%s && ST_GeomFromText(?, %d)", colRef, srid), []any{geom}
	case "left":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("%s << ST_GeomFromText(?, %d)", colRef, srid), []any{geom}
	case "right":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("%s >> ST_GeomFromText(?, %d)", colRef, srid), []any{geom}
	case "above":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("%s |>> ST_GeomFromText(?, %d)", colRef, srid), []any{geom}
	case "below":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("%s <<| ST_GeomFromText(?, %d)", colRef, srid), []any{geom}
	case "strictly_above":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("%s |>> ST_GeomFromText(?, %d)", colRef, srid), []any{geom}
	case "strictly_below":
		geom, srid := extractGeometry(val)
		return fmt.Sprintf("%s <<| ST_GeomFromText(?, %d)", colRef, srid), []any{geom}
	case "distance_lte":
		geom, srid, dist := extractDistanceGeometry(val)
		return fmt.Sprintf("ST_DWithin(%s, ST_GeomFromText(?, %d), ?)", colRef, srid), []any{geom, dist}
	case "distance_lt":
		geom, srid, dist := extractDistanceGeometry(val)
		return fmt.Sprintf("ST_DWithin(%s, ST_GeomFromText(?, %d), ?) AND NOT ST_Equals(%s, ST_GeomFromText(?, %d))", colRef, srid, colRef, srid), []any{geom, dist, geom, srid}
	case "distance_gt":
		geom, srid, dist := extractDistanceGeometry(val)
		return fmt.Sprintf("NOT ST_DWithin(%s, ST_GeomFromText(?, %d), ?)", colRef, srid), []any{geom, dist}
	case "distance_gte":
		geom, srid, dist := extractDistanceGeometry(val)
		return fmt.Sprintf("ST_Distance(%s, ST_GeomFromText(?, %d)) >= ?", colRef, srid), []any{geom, dist}
	default:
		// Fallback to exact if unknown
		ph, p := resolveVal(val, info)
		return fmt.Sprintf("%s = %s", colRef, ph), p
	}
}

// expandInLookup expands a slice/array value into IN (?, ?, ...) syntax.
// A non-slice value is treated as a single-element IN.
func expandInLookup(colRef string, val any) (string, []any) {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		if v.Len() == 0 {
			return "1=0", nil
		}
		placeholders := make([]string, v.Len())
		params := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			placeholders[i] = "?"
			params[i] = v.Index(i).Interface()
		}
		return fmt.Sprintf("%s IN (%s)", colRef, strings.Join(placeholders, ", ")), params
	}
	return fmt.Sprintf("%s IN (?)", colRef), []any{val}
}

// expandRangeLookup handles the __range lookup with a 2-element slice.
func expandRangeLookup(colRef string, val any) (string, []any) {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Slice && v.Len() == 2 {
		return fmt.Sprintf("%s BETWEEN ? AND ?", colRef), []any{v.Index(0).Interface(), v.Index(1).Interface()}
	}
	return fmt.Sprintf("%s BETWEEN ? AND ?", colRef), []any{val}
}

// extractPKIfNeeded extracts the primary key value from a model instance
// when filtering on a ForeignKey field directly (e.g. Filter(Lookup{"author": userObj})).
// For scalar values (int, string, etc.) it returns the value unchanged.
func extractPKIfNeeded(val any, info *orm.ModelInfo, parts []string) any {
	fieldParts := parts
	if len(parts) > 1 && knownLookupTypes[parts[len(parts)-1]] {
		fieldParts = parts[:len(parts)-1]
	}
	if len(fieldParts) != 1 || info == nil {
		return val
	}
	field, ok := info.FieldByName[fieldParts[0]]
	if !ok || (field.Type != orm.ForeignKey && field.Type != orm.OneToOneField) {
		return val
	}
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return val
	}
	idField := v.FieldByName("ID")
	if idField.IsValid() && idField.CanInterface() {
		return idField.Interface()
	}
	return val
}

// extractGeometry extracts a WKT string and SRID from a value that implements
// orm.GeometryValue. If the value is a raw string it is treated as WKT with
// SRID 4326 (WGS84).
func extractGeometry(val any) (string, int) {
	if gv, ok := val.(orm.GeometryValue); ok {
		return gv.WKT(), gv.GetSRID()
	}
	if s, ok := val.(string); ok {
		return s, 4326
	}
	return fmt.Sprintf("%v", val), 4326
}

// extractDistanceGeometry extracts WKT, SRID, and distance in meters from a
// value that implements orm.DistanceGeometryValue.
func extractDistanceGeometry(val any) (string, int, float64) {
	if dv, ok := val.(orm.DistanceGeometryValue); ok {
		return dv.WKT(), dv.GetSRID(), dv.DistanceMeters()
	}
	if gv, ok := val.(orm.GeometryValue); ok {
		return gv.WKT(), gv.GetSRID(), 0
	}
	if s, ok := val.(string); ok {
		return s, 4326, 0
	}
	return fmt.Sprintf("%v", val), 4326, 0
}

// AggExpr represents an aggregate expression.
type AggExpr interface {
	ToSQL() string
}

// ToUpdateSQL generates an UPDATE statement for the given field-value pairs.
// Values that implement Expression (e.g. F) are resolved to SQL column references.
func (q *Query) ToUpdateSQL(fields map[string]any) (string, []any) {
	tableName := q.ModelInfo.Meta.DbTable

	// Sort field names for deterministic SQL output (Go maps iterate in random order).
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)

	var setClauses []string
	var params []any

	for _, name := range names {
		val := fields[name]
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
		clause, p := q.Where.toSQL(q.ModelInfo, nil)
		if clause != "" {
			whereClauses = append(whereClauses, "("+clause+")")
			params = append(params, p...)
		}
	}

	if q.Exclude != nil {
		clause, p := q.Exclude.toSQL(q.ModelInfo, nil)
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
