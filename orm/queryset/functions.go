package queryset

import (
	"fmt"
	"strings"

	"github.com/pkshahid/JanGo/orm"
)

// ---------------------------------------------------------------------------
// Func — the base type for all SQL database functions
// ---------------------------------------------------------------------------

// Func represents a SQL function call with zero or more arguments.
// Each argument may be an Expression (e.g. F, another Func, Case) or a
// literal Go value that becomes a bind parameter.
//
// Func implements both Expression and AggExpr, so it can be used in
// Filter, Update, Annotate, and Aggregate just like F.
type Func struct {
	name string
	args []any
}

// NewFunc creates a generic SQL function expression.
func NewFunc(name string, args ...any) *Func {
	return &Func{name: name, args: args}
}

// resolveArg renders a single function argument to SQL.
func resolveArg(arg any, info *orm.ModelInfo) (string, []any) {
	if expr, ok := arg.(Expression); ok {
		return expr.ResolveSQL(info)
	}
	return "?", []any{arg}
}

// ResolveSQL renders the function call to a SQL fragment with bind parameters.
func (f *Func) ResolveSQL(info *orm.ModelInfo) (string, []any) {
	var argStrs []string
	var params []any
	for _, arg := range f.args {
		sql, p := resolveArg(arg, info)
		argStrs = append(argStrs, sql)
		params = append(params, p...)
	}
	return fmt.Sprintf("%s(%s)", f.name, strings.Join(argStrs, ", ")), params
}

// ToSQL renders the function call without column resolution,
// satisfying the AggExpr interface for use in Annotate.
func (f *Func) ToSQL() string {
	sql, _ := f.ResolveSQL(nil)
	return sql
}

// ---------------------------------------------------------------------------
// Text functions
// ---------------------------------------------------------------------------

// Lower returns the lowercase version of the given expression.
func Lower(expr any) *Func { return &Func{name: "LOWER", args: []any{expr}} }

// Upper returns the uppercase version of the given expression.
func Upper(expr any) *Func { return &Func{name: "UPPER", args: []any{expr}} }

// Length returns the number of characters in the given expression.
func Length(expr any) *Func { return &Func{name: "LENGTH", args: []any{expr}} }

// Concat concatenates two or more expressions into a single string.
func Concat(exprs ...any) *Func { return &Func{name: "CONCAT", args: exprs} }

// Substr returns a substring starting at pos (1-indexed) with optional length.
func Substr(expr any, pos any, length ...any) *Func {
	args := []any{expr, pos}
	args = append(args, length...)
	return &Func{name: "SUBSTR", args: args}
}

// Replace replaces all occurrences of find with replacement in expr.
func Replace(expr, find, replacement any) *Func {
	return &Func{name: "REPLACE", args: []any{expr, find, replacement}}
}

// Trim removes leading and trailing whitespace from expr.
func Trim(expr any) *Func { return &Func{name: "TRIM", args: []any{expr}} }

// LTrim removes leading whitespace from expr.
func LTrim(expr any) *Func { return &Func{name: "LTRIM", args: []any{expr}} }

// RTrim removes trailing whitespace from expr.
func RTrim(expr any) *Func { return &Func{name: "RTRIM", args: []any{expr}} }

// Left returns the first length characters of expr.
func Left(expr any, length any) *Func {
	return &Func{name: "LEFT", args: []any{expr, length}}
}

// Right returns the last length characters of expr.
func Right(expr any, length any) *Func {
	return &Func{name: "RIGHT", args: []any{expr, length}}
}

// Reverse reverses the characters in expr.
func Reverse(expr any) *Func { return &Func{name: "REVERSE", args: []any{expr}} }

// StrIndex returns the 1-indexed position of substring within string.
func StrIndex(str, sub any) *Func {
	return &Func{name: "INSTR", args: []any{str, sub}}
}

// Repeat repeats expr count times.
func Repeat(expr any, count any) *Func {
	return &Func{name: "REPEAT", args: []any{expr, count}}
}

// Chr returns the character for the given ASCII/Unicode code point.
func Chr(expr any) *Func { return &Func{name: "CHR", args: []any{expr}} }

// Ascii returns the ASCII code of the first character of expr.
func Ascii(expr any) *Func { return &Func{name: "ASCII", args: []any{expr}} }

// ---------------------------------------------------------------------------
// Math functions
// ---------------------------------------------------------------------------

// Abs returns the absolute value of expr.
func Abs(expr any) *Func { return &Func{name: "ABS", args: []any{expr}} }

// Round rounds expr to an optional precision (default 0).
func Round(expr any, precision ...any) *Func {
	args := []any{expr}
	args = append(args, precision...)
	return &Func{name: "ROUND", args: args}
}

// Power returns expr raised to the power of exponent.
func Power(expr, exponent any) *Func {
	return &Func{name: "POWER", args: []any{expr, exponent}}
}

// Sqrt returns the square root of expr.
func Sqrt(expr any) *Func { return &Func{name: "SQRT", args: []any{expr}} }

// ModFunc returns expr1 modulo expr2.
// Named ModFunc to avoid collision with F.Mod.
func ModFunc(expr1, expr2 any) *Func {
	return &Func{name: "MOD", args: []any{expr1, expr2}}
}

// Floor returns the largest integer <= expr.
func Floor(expr any) *Func { return &Func{name: "FLOOR", args: []any{expr}} }

// Ceil returns the smallest integer >= expr.
func Ceil(expr any) *Func { return &Func{name: "CEIL", args: []any{expr}} }

// Sign returns -1, 0, or 1 depending on whether expr is negative, zero, or positive.
func Sign(expr any) *Func { return &Func{name: "SIGN", args: []any{expr}} }

// Exp returns e raised to the power of expr.
func Exp(expr any) *Func { return &Func{name: "EXP", args: []any{expr}} }

// Ln returns the natural logarithm of expr.
func Ln(expr any) *Func { return &Func{name: "LN", args: []any{expr}} }

// Log returns the logarithm of expr. With one arg it is natural log on some
// DBs and base-10 on others; with two args it is log base expr2 of expr1.
func Log(args ...any) *Func { return &Func{name: "LOG", args: args} }

// Degrees converts expr from radians to degrees.
func Degrees(expr any) *Func { return &Func{name: "DEGREES", args: []any{expr}} }

// Radians converts expr from degrees to radians.
func Radians(expr any) *Func { return &Func{name: "RADIANS", args: []any{expr}} }

// Pi returns the value of pi.
func Pi() *Func { return &Func{name: "PI"} }

// Rand returns a random float. With an optional seed argument on some DBs.
func Rand(seed ...any) *Func {
	return &Func{name: "RAND", args: seed}
}

// Greatest returns the largest of the given expressions.
func Greatest(exprs ...any) *Func { return &Func{name: "GREATEST", args: exprs} }

// Least returns the smallest of the given expressions.
func Least(exprs ...any) *Func { return &Func{name: "LEAST", args: exprs} }

// Acos returns the arc cosine of expr.
func Acos(expr any) *Func { return &Func{name: "ACOS", args: []any{expr}} }

// Asin returns the arc sine of expr.
func Asin(expr any) *Func { return &Func{name: "ASIN", args: []any{expr}} }

// Atan returns the arc tangent of expr.
func Atan(expr any) *Func { return &Func{name: "ATAN", args: []any{expr}} }

// Atan2 returns the arc tangent of y/x.
func Atan2(y, x any) *Func { return &Func{name: "ATAN2", args: []any{y, x}} }

// Cos returns the cosine of expr.
func Cos(expr any) *Func { return &Func{name: "COS", args: []any{expr}} }

// Sin returns the sine of expr.
func Sin(expr any) *Func { return &Func{name: "SIN", args: []any{expr}} }

// Tan returns the tangent of expr.
func Tan(expr any) *Func { return &Func{name: "TAN", args: []any{expr}} }

// Cot returns the cotangent of expr.
func Cot(expr any) *Func { return &Func{name: "COT", args: []any{expr}} }

// ---------------------------------------------------------------------------
// Date / Time functions
// ---------------------------------------------------------------------------

// Now returns the current timestamp.
func Now() *Func { return &Func{name: "NOW"} }

// CurrentDate returns the current date.
func CurrentDate() *Func { return &Func{name: "CURRENT_DATE"} }

// CurrentTime returns the current time.
func CurrentTime() *Func { return &Func{name: "CURRENT_TIME"} }

// UnixTimestamp returns the Unix epoch seconds for expr (or now if no arg).
func UnixTimestamp(expr ...any) *Func {
	return &Func{name: "UNIX_TIMESTAMP", args: expr}
}

// DateFunc extracts the date portion of a datetime expression.
func DateFunc(expr any) *Func { return &Func{name: "DATE", args: []any{expr}} }

// TimeFunc extracts the time portion of a datetime expression.
func TimeFunc(expr any) *Func { return &Func{name: "TIME", args: []any{expr}} }

// Year extracts the year from a date/datetime expression.
func Year(expr any) *Func { return &Func{name: "YEAR", args: []any{expr}} }

// Month extracts the month (1-12) from a date/datetime expression.
func Month(expr any) *Func { return &Func{name: "MONTH", args: []any{expr}} }

// Day extracts the day of the month from a date/datetime expression.
func Day(expr any) *Func { return &Func{name: "DAY", args: []any{expr}} }

// Hour extracts the hour from a time/datetime expression.
func Hour(expr any) *Func { return &Func{name: "HOUR", args: []any{expr}} }

// Minute extracts the minute from a time/datetime expression.
func Minute(expr any) *Func { return &Func{name: "MINUTE", args: []any{expr}} }

// Second extracts the second from a time/datetime expression.
func Second(expr any) *Func { return &Func{name: "SECOND", args: []any{expr}} }

// Week extracts the ISO week number from a date expression.
func Week(expr any) *Func { return &Func{name: "WEEK", args: []any{expr}} }

// WeekDay extracts the day of the week (1=Sunday) from a date expression.
func WeekDay(expr any) *Func { return &Func{name: "WEEKDAY", args: []any{expr}} }

// Quarter extracts the quarter (1-4) from a date expression.
func Quarter(expr any) *Func { return &Func{name: "QUARTER", args: []any{expr}} }

// DayName returns the name of the day of week for a date expression.
func DayName(expr any) *Func { return &Func{name: "DAYNAME", args: []any{expr}} }

// MonthName returns the name of the month for a date expression.
func MonthName(expr any) *Func { return &Func{name: "MONTHNAME", args: []any{expr}} }

// DayOfWeek returns the day of week index (1=Sunday) for a date expression.
func DayOfWeek(expr any) *Func { return &Func{name: "DAYOFWEEK", args: []any{expr}} }

// DayOfYear returns the day of year (1-366) for a date expression.
func DayOfYear(expr any) *Func { return &Func{name: "DAYOFYEAR", args: []any{expr}} }

// DateAdd adds an interval to a date/datetime expression.
func DateAdd(expr, interval, unit any) *Func {
	return &Func{name: "DATE_ADD", args: []any{expr, fmt.Sprintf("INTERVAL %v %s", interval, unit)}}
}

// DateSub subtracts an interval from a date/datetime expression.
func DateSub(expr, interval, unit any) *Func {
	return &Func{name: "DATE_SUB", args: []any{expr, fmt.Sprintf("INTERVAL %v %s", interval, unit)}}
}

// DateDiff returns the number of days between two date expressions.
func DateDiff(expr1, expr2 any) *Func {
	return &Func{name: "DATEDIFF", args: []any{expr1, expr2}}
}

// TimeDiff returns the difference between two time/datetime expressions.
func TimeDiff(expr1, expr2 any) *Func {
	return &Func{name: "TIMEDIFF", args: []any{expr1, expr2}}
}

// LastDay returns the last day of the month for a date expression.
func LastDay(expr any) *Func { return &Func{name: "LAST_DAY", args: []any{expr}} }

// ---------------------------------------------------------------------------
// Extract — EXTRACT(field FROM source)
// ---------------------------------------------------------------------------

// Extract represents a SQL EXTRACT(field FROM source) expression.
type Extract struct {
	source any
	field  string
}

// NewExtract creates an EXTRACT expression.
// Common field values: YEAR, MONTH, DAY, HOUR, MINUTE, SECOND, WEEK, QUARTER.
func NewExtract(source any, field string) *Extract {
	return &Extract{source: source, field: field}
}

// ResolveSQL renders the EXTRACT expression to SQL.
func (e *Extract) ResolveSQL(info *orm.ModelInfo) (string, []any) {
	srcSQL, srcParams := resolveArg(e.source, info)
	sql := fmt.Sprintf("EXTRACT(%s FROM %s)", e.field, srcSQL)
	return sql, srcParams
}

// ToSQL renders the EXTRACT expression without column resolution.
func (e *Extract) ToSQL() string {
	sql, _ := e.ResolveSQL(nil)
	return sql
}

// ---------------------------------------------------------------------------
// Cast — CAST(expr AS type)
// ---------------------------------------------------------------------------

// Cast represents a SQL CAST(expression AS type) expression.
type Cast struct {
	expr     any
	typeName string
}

// NewCast creates a CAST expression that converts expr to the given SQL type.
// Common type names: INTEGER, TEXT, REAL, BLOB, VARCHAR(n), DECIMAL(p,s), etc.
func NewCast(expr any, typeName string) *Cast {
	return &Cast{expr: expr, typeName: typeName}
}

// ResolveSQL renders the CAST expression to SQL.
func (c *Cast) ResolveSQL(info *orm.ModelInfo) (string, []any) {
	exprSQL, exprParams := resolveArg(c.expr, info)
	sql := fmt.Sprintf("CAST(%s AS %s)", exprSQL, c.typeName)
	return sql, exprParams
}

// ToSQL renders the CAST expression without column resolution.
func (c *Cast) ToSQL() string {
	sql, _ := c.ResolveSQL(nil)
	return sql
}

// ---------------------------------------------------------------------------
// Coalesce — COALESCE(expr1, expr2, ...)
// ---------------------------------------------------------------------------

// Coalesce returns the first non-NULL value among the given expressions.
func Coalesce(exprs ...any) *Func {
	return &Func{name: "COALESCE", args: exprs}
}

// ---------------------------------------------------------------------------
// NullIf — NULLIF(expr1, expr2)
// ---------------------------------------------------------------------------

// NullIf returns NULL if expr1 equals expr2, otherwise returns expr1.
func NullIf(expr1, expr2 any) *Func {
	return &Func{name: "NULLIF", args: []any{expr1, expr2}}
}

// ---------------------------------------------------------------------------
// IfNull — IFNULL(expr, value) / NVL equivalent
// ---------------------------------------------------------------------------

// IfNull returns value if expr is NULL, otherwise returns expr.
func IfNull(expr, value any) *Func {
	return &Func{name: "IFNULL", args: []any{expr, value}}
}

// ---------------------------------------------------------------------------
// Comparison & conditional functions
// ---------------------------------------------------------------------------

// IIf is a simple inline conditional: IIF(condition, true_val, false_val).
// The condition is a QNode tree.
type IIf struct {
	cond     *QNode
	trueVal  any
	falseVal any
}

// NewIIf creates an IIF expression.
func NewIIf(cond *QNode, trueVal, falseVal any) *IIf {
	return &IIf{cond: cond, trueVal: trueVal, falseVal: falseVal}
}

// ResolveSQL renders the IIF expression to SQL.
func (i *IIf) ResolveSQL(info *orm.ModelInfo) (string, []any) {
	condSQL, condParams := i.cond.toSQL(info)
	trueSQL, trueParams := resolveArg(i.trueVal, info)
	falseSQL, falseParams := resolveArg(i.falseVal, info)
	sql := fmt.Sprintf("IIF(%s, %s, %s)", condSQL, trueSQL, falseSQL)
	params := append(condParams, trueParams...)
	params = append(params, falseParams...)
	return sql, params
}

// ToSQL renders the IIF expression without column resolution.
func (i *IIf) ToSQL() string {
	sql, _ := i.ResolveSQL(nil)
	return sql
}

// ---------------------------------------------------------------------------
// String aggregation
// ---------------------------------------------------------------------------

// StringAgg represents GROUP_CONCAT / STRING_AGG.
// It concatenates values of expr across grouped rows, separated by separator.
type StringAgg struct {
	expr      any
	separator any
	distinct  bool
}

// NewStringAgg creates a string aggregation expression.
func NewStringAgg(expr, separator any) *StringAgg {
	return &StringAgg{expr: expr, separator: separator}
}

// Distinct makes the aggregation only consider distinct values.
func (s *StringAgg) Distinct() *StringAgg {
	cp := *s
	cp.distinct = true
	return &cp
}

// ResolveSQL renders the GROUP_CONCAT expression to SQL.
func (s *StringAgg) ResolveSQL(info *orm.ModelInfo) (string, []any) {
	exprSQL, exprParams := resolveArg(s.expr, info)
	sepSQL, sepParams := resolveArg(s.separator, info)

	keyword := "GROUP_CONCAT"
	prefix := ""
	if s.distinct {
		prefix = "DISTINCT "
	}

	sql := fmt.Sprintf("%s(%s%s SEPARATOR %s)", keyword, prefix, exprSQL, sepSQL)
	params := append(exprParams, sepParams...)
	return sql, params
}

// ToSQL renders the GROUP_CONCAT expression without column resolution.
func (s *StringAgg) ToSQL() string {
	sql, _ := s.ResolveSQL(nil)
	return sql
}
