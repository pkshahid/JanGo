package queryset

import (
	"fmt"
	"strings"
)

// Aggregate functions
type Sum string
type Count string
type Avg string
type Max string
type Min string
type StdDev string
type Variance string

func (a Sum) ToSQL() string      { return fmt.Sprintf("SUM(%s)", string(a)) }
func (a Count) ToSQL() string    { return fmt.Sprintf("COUNT(%s)", string(a)) }
func (a Avg) ToSQL() string      { return fmt.Sprintf("AVG(%s)", string(a)) }
func (a Max) ToSQL() string      { return fmt.Sprintf("MAX(%s)", string(a)) }
func (a Min) ToSQL() string      { return fmt.Sprintf("MIN(%s)", string(a)) }
func (a StdDev) ToSQL() string   { return fmt.Sprintf("STDDEV(%s)", string(a)) }
func (a Variance) ToSQL() string { return fmt.Sprintf("VARIANCE(%s)", string(a)) }

// Aggregate executes the query and returns the requested aggregations.
// Unlike Annotate, Aggregate is a terminal method that fires the query.
func (qs QuerySet[T]) Aggregate(exprs ...AggExpr) (map[string]any, error) {
	// 1. Build the base query string
	// A full implementation executes SQL here. For prototype, we generate the SQL string to verify logic.

	tableName := qs.query.ModelInfo.Meta.DbTable
	var aggStrs []string

	// Create keys for the output map based on standard Django naming, e.g., "views__sum"
	// We'll just append them to the SELECT.
	for _, expr := range exprs {
		aggStrs = append(aggStrs, expr.ToSQL())
	}

	selectClause := strings.Join(aggStrs, ", ")
	sql := fmt.Sprintf("SELECT %s FROM %s", selectClause, tableName)

	var whereClauses []string
	var params []any

	if qs.query.Where != nil {
		clause, p := qs.query.Where.toSQL(qs.query.ModelInfo, nil)
		if clause != "" {
			whereClauses = append(whereClauses, "("+clause+")")
			params = append(params, p...)
		}
	}

	if qs.query.Exclude != nil {
		clause, p := qs.query.Exclude.toSQL(qs.query.ModelInfo, nil)
		if clause != "" {
			whereClauses = append(whereClauses, "NOT ("+clause+")")
			params = append(params, p...)
		}
	}

	if len(whereClauses) > 0 {
		sql += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// We don't append ORDER BY or LIMIT for pure aggregate, as they don't apply to total aggregates unless grouped.

	// Mock database execution
	// Executor.QueryRow(sql, params...) -> scan into map

	// Return a dummy map for compilation matching signature
	res := make(map[string]any)
	res["_sql"] = sql // Injecting the generated SQL for test verification

	return res, nil
}
