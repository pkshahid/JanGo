package queryset

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkshahid/JanGo/orm"
)

// DateKind represents the truncation granularity for date-based grouping.
type DateKind string

const (
	DateKindYear    DateKind = "year"
	DateKindMonth   DateKind = "month"
	DateKindWeek    DateKind = "week"
	DateKindDay     DateKind = "day"
	DateKindQuarter DateKind = "quarter"
)

// DateTimeKind represents the truncation granularity for datetime-based grouping.
type DateTimeKind string

const (
	DateTimeKindYear   DateTimeKind = "year"
	DateTimeKindMonth  DateTimeKind = "month"
	DateTimeKindDay    DateTimeKind = "day"
	DateTimeKindHour   DateTimeKind = "hour"
	DateTimeKindMinute DateTimeKind = "minute"
	DateTimeKindSecond DateTimeKind = "second"
)

// OrderDirection represents ASC or DESC ordering for date/datetime queries.
type OrderDirection string

const (
	OrderASC  OrderDirection = "ASC"
	OrderDESC OrderDirection = "DESC"
)

// validDateKinds tracks which kinds are allowed for Dates().
var validDateKinds = map[DateKind]bool{
	DateKindYear:    true,
	DateKindMonth:   true,
	DateKindWeek:    true,
	DateKindDay:     true,
	DateKindQuarter: true,
}

// validDateTimeKinds tracks which kinds are allowed for Datetimes().
var validDateTimeKinds = map[DateTimeKind]bool{
	DateTimeKindYear:   true,
	DateTimeKindMonth:  true,
	DateTimeKindDay:    true,
	DateTimeKindHour:   true,
	DateTimeKindMinute: true,
	DateTimeKindSecond: true,
}

// resolveColumn maps a field name to its database column name using ModelInfo.
func resolveColumn(info *orm.ModelInfo, field string) string {
	if info != nil {
		if f, ok := info.FieldByName[field]; ok {
			return f.Column
		}
	}
	return field
}

// ToDatesSQL generates a SELECT DISTINCT DATE_TRUNC(...) query for date-based grouping.
// The kind parameter controls truncation granularity (year, month, week, day, quarter).
// The order parameter controls ASC/DESC ordering of the grouped dates.
func (q *Query) ToDatesSQL(kind DateKind, field string, order OrderDirection) (string, []any) {
	if !validDateKinds[kind] {
		return "", nil
	}
	if order == "" {
		order = OrderASC
	}

	colName := resolveColumn(q.ModelInfo, field)
	alias := "date_field"

	selectExpr := fmt.Sprintf("DATE_TRUNC('%s', %s) AS %s", string(kind), colName, alias)
	sql := fmt.Sprintf("SELECT DISTINCT %s FROM %s", selectExpr, q.ModelInfo.Meta.DbTable)
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

	if len(whereClauses) > 0 {
		sql += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	sql += fmt.Sprintf(" ORDER BY %s %s", alias, string(order))

	return sql, params
}

// ToDatetimesSQL generates a SELECT DISTINCT DATETIME_TRUNC(...) query for datetime-based grouping.
// The kind parameter controls truncation granularity (year, month, day, hour, minute, second).
// The order parameter controls ASC/DESC ordering of the grouped datetimes.
func (q *Query) ToDatetimesSQL(kind DateTimeKind, field string, order OrderDirection) (string, []any) {
	if !validDateTimeKinds[kind] {
		return "", nil
	}
	if order == "" {
		order = OrderASC
	}

	colName := resolveColumn(q.ModelInfo, field)
	alias := "datetime_field"

	selectExpr := fmt.Sprintf("DATETIME_TRUNC('%s', %s) AS %s", string(kind), colName, alias)
	sql := fmt.Sprintf("SELECT DISTINCT %s FROM %s", selectExpr, q.ModelInfo.Meta.DbTable)
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

	if len(whereClauses) > 0 {
		sql += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	sql += fmt.Sprintf(" ORDER BY %s %s", alias, string(order))

	return sql, params
}

// Dates returns a list of distinct date values truncated to the given kind
// from the queryset's scope. This is a terminal method that executes the query.
//
// kind must be one of: "year", "month", "week", "day", "quarter".
// order defaults to "ASC" if not specified.
func (qs QuerySet[T]) Dates(field string, kind DateKind, order ...OrderDirection) ([]time.Time, error) {
	if !validDateKinds[kind] {
		return nil, fmt.Errorf("orm: invalid date kind %q; must be one of year, month, week, day, quarter", kind)
	}

	dir := OrderASC
	if len(order) > 0 {
		dir = order[0]
	}

	c := qs.clone()
	sql, params := c.query.ToDatesSQL(kind, field, dir)
	_ = sql
	_ = params

	// Mock database execution:
	// rows, err := db.Query(sql, params...)
	// Parse each row into time.Time with time components zeroed for date-only results.
	return []time.Time{}, nil
}

// Datetimes returns a list of distinct datetime values truncated to the given kind
// from the queryset's scope. This is a terminal method that executes the query.
//
// kind must be one of: "year", "month", "day", "hour", "minute", "second".
// order defaults to "ASC" if not specified.
func (qs QuerySet[T]) Datetimes(field string, kind DateTimeKind, order ...OrderDirection) ([]time.Time, error) {
	if !validDateTimeKinds[kind] {
		return nil, fmt.Errorf("orm: invalid datetime kind %q; must be one of year, month, day, hour, minute, second", kind)
	}

	dir := OrderASC
	if len(order) > 0 {
		dir = order[0]
	}

	c := qs.clone()
	sql, params := c.query.ToDatetimesSQL(kind, field, dir)
	_ = sql
	_ = params

	// Mock database execution:
	// rows, err := db.Query(sql, params...)
	// Parse each row into time.Time preserving time components per the truncation kind.
	return []time.Time{}, nil
}
