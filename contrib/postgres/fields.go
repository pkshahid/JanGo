// Package postgres provides PostgreSQL-specific field types and helpers
// inspired by Django's contrib.postgres package. It includes:
//
//   - ArrayField: stores Go slices/arrays as PostgreSQL ARRAY columns
//   - HStoreField: stores map[string]string as PostgreSQL HSTORE columns
//   - Range types: IntRange, BigIntRange, NumRange, TsTzRange, DateRange
//     for PostgreSQL's range column types
//   - CICharField / CITextField / CIEmailField: case-insensitive text fields
//     backed by PostgreSQL's citext extension
//
// Field types are declared via struct tags using the orm package's `gd` tag
// syntax. For example:
//
//	type Product struct {
//	    orm.Model
//	    Tags    []string       `gd:"ArrayField,base_field=CharField,max_length=50"`
//	    Attrs   map[string]string `gd:"HStoreField,null=true"`
//	    Prices  IntRange       `gd:"IntegerRangeField,null=true"`
//	    Title   string         `gd:"CICharField,max_length=200"`
//	}
package postgres

import (
	"fmt"
	"strings"
	"time"
)

// ─── Range Types ───────────────────────────────────────────────────────────

// Range represents a bounded or unbounded interval. Any of the bounds may be
// nil to indicate an open-ended range. This mirrors PostgreSQL's range types
// (int4range, int8range, numrange, tstzrange, daterange).
type Range[T any] struct {
	Lower     *T
	Upper     *T
	LowerIncl bool // true → "[" (inclusive), false → "(" (exclusive)
	UpperIncl bool // true → "]" (inclusive), false → ")" (exclusive)
}

// IntRange is a convenience alias for Range[int].
type IntRange = Range[int]

// BigIntRange is a convenience alias for Range[int64].
type BigIntRange = Range[int64]

// NumRange is a convenience alias for Range[float64].
type NumRange = Range[float64]

// TsTzRange is a convenience alias for Range[time.Time].
type TsTzRange = Range[time.Time]

// DateRange is a convenience alias for Range[time.Time] (date-only).
type DateRange = Range[time.Time]

// NewRange creates a Range with inclusive bounds.
func NewRange[T any](lower, upper T) Range[T] {
	return Range[T]{
		Lower:     &lower,
		Upper:     &upper,
		LowerIncl: true,
		UpperIncl: true,
	}
}

// NewRangeExclusive creates a Range with exclusive bounds.
func NewRangeExclusive[T any](lower, upper T) Range[T] {
	return Range[T]{
		Lower:     &lower,
		Upper:     &upper,
		LowerIncl: false,
		UpperIncl: false,
	}
}

// NewRangeFrom creates a Range with only a lower bound (upper is unbounded).
func NewRangeFrom[T any](lower T, inclusive bool) Range[T] {
	return Range[T]{
		Lower:     &lower,
		LowerIncl: inclusive,
	}
}

// NewRangeTo creates a Range with only an upper bound (lower is unbounded).
func NewRangeTo[T any](upper T, inclusive bool) Range[T] {
	return Range[T]{
		Upper:     &upper,
		UpperIncl: inclusive,
	}
}

// IsEmpty returns true if both bounds are nil (an empty/unbounded range).
func (r Range[T]) IsEmpty() bool {
	return r.Lower == nil && r.Upper == nil
}

// String returns the PostgreSQL range literal, e.g. "[1,10)", "(,100]", or
// "empty" when both bounds are nil.
func (r Range[T]) String() string {
	if r.IsEmpty() {
		return "empty"
	}
	var sb strings.Builder
	if r.LowerIncl {
		sb.WriteByte('[')
	} else {
		sb.WriteByte('(')
	}
	if r.Lower != nil {
		sb.WriteString(fmt.Sprintf("%v", *r.Lower))
	}
	sb.WriteByte(',')
	if r.Upper != nil {
		sb.WriteString(fmt.Sprintf("%v", *r.Upper))
	}
	if r.UpperIncl {
		sb.WriteByte(']')
	} else {
		sb.WriteByte(')')
	}
	return sb.String()
}

// ─── HStore Helpers ────────────────────────────────────────────────────────

// HStore is a convenience type for map[string]string, representing a
// PostgreSQL HSTORE column (key-value pairs where both keys and values are
// text).
type HStore map[string]string

// NewHStore creates an HStore from an even number of alternating key-value
// string arguments.
func NewHStore(kv ...string) HStore {
	h := HStore{}
	for i := 0; i+1 < len(kv); i += 2 {
		h[kv[i]] = kv[i+1]
	}
	return h
}

// Get returns the value for key, or "" if not present.
func (h HStore) Get(key string) string {
	return h[key]
}

// Set sets key to value.
func (h HStore) Set(key, value string) {
	h[key] = value
}

// ─── ArrayField Helpers ────────────────────────────────────────────────────

// ArrayInt is a convenience type for []int, representing a PostgreSQL
// INTEGER[] column.
type ArrayInt []int

// ArrayInt64 is a convenience type for []int64, representing a PostgreSQL
// BIGINT[] column.
type ArrayInt64 []int64

// ArrayFloat64 is a convenience type for []float64, representing a PostgreSQL
// DOUBLE PRECISION[] column.
type ArrayFloat64 []float64

// ArrayString is a convenience type for []string, representing a PostgreSQL
// VARCHAR[] or TEXT[] column.
type ArrayString []string

// ArrayBool is a convenience type for []bool, representing a PostgreSQL
// BOOLEAN[] column.
type ArrayBool []bool

// ─── Tag Helpers ───────────────────────────────────────────────────────────

// ArrayFieldTag builds a `gd` struct tag value for an ArrayField with the
// given base field type and options. For example:
//
//	tag := postgres.ArrayFieldTag("IntegerField", postgres.ArrayFieldOpts{Size: 3})
//	// → "ArrayField,base_field=IntegerField,size=3"
func ArrayFieldTag(baseField string, opts ArrayFieldOpts) string {
	parts := []string{"ArrayField", "base_field=" + baseField}
	if opts.MaxLength > 0 {
		parts = append(parts, fmt.Sprintf("max_length=%d", opts.MaxLength))
	}
	if opts.MaxDigits > 0 {
		parts = append(parts, fmt.Sprintf("max_digits=%d", opts.MaxDigits))
	}
	if opts.DecimalPlaces > 0 {
		parts = append(parts, fmt.Sprintf("decimal_places=%d", opts.DecimalPlaces))
	}
	if opts.Size > 0 {
		parts = append(parts, fmt.Sprintf("size=%d", opts.Size))
	}
	if opts.Null {
		parts = append(parts, "null=true")
	}
	if opts.Blank {
		parts = append(parts, "blank=true")
	}
	return strings.Join(parts, ",")
}

// ArrayFieldOpts holds options for building an ArrayField tag.
type ArrayFieldOpts struct {
	Size          int
	Null          bool
	Blank         bool
	MaxLength     int // for CharField base
	MaxDigits     int // for DecimalField base
	DecimalPlaces int // for DecimalField base
}

// RangeFieldTag builds a `gd` struct tag value for a range field type.
// fieldType must be one of "IntegerRangeField", "BigIntegerRangeField",
// "DecimalRangeField", "DateTimeRangeField", "DateRangeField".
func RangeFieldTag(fieldType string, null, blank bool) string {
	parts := []string{fieldType}
	if null {
		parts = append(parts, "null=true")
	}
	if blank {
		parts = append(parts, "blank=true")
	}
	return strings.Join(parts, ",")
}

// CICharFieldTag builds a `gd` struct tag value for a CICharField.
func CICharFieldTag(maxLength int, null, blank bool) string {
	parts := []string{"CICharField", fmt.Sprintf("max_length=%d", maxLength)}
	if null {
		parts = append(parts, "null=true")
	}
	if blank {
		parts = append(parts, "blank=true")
	}
	return strings.Join(parts, ",")
}

// CITextFieldTag builds a `gd` struct tag value for a CITextField.
func CITextFieldTag(null, blank bool) string {
	parts := []string{"CITextField"}
	if null {
		parts = append(parts, "null=true")
	}
	if blank {
		parts = append(parts, "blank=true")
	}
	return strings.Join(parts, ",")
}

// CIEmailFieldTag builds a `gd` struct tag value for a CIEmailField.
func CIEmailFieldTag(null, blank bool) string {
	parts := []string{"CIEmailField"}
	if null {
		parts = append(parts, "null=true")
	}
	if blank {
		parts = append(parts, "blank=true")
	}
	return strings.Join(parts, ",")
}

// HStoreFieldTag builds a `gd` struct tag value for an HStoreField.
func HStoreFieldTag(null, blank bool) string {
	parts := []string{"HStoreField"}
	if null {
		parts = append(parts, "null=true")
	}
	if blank {
		parts = append(parts, "blank=true")
	}
	return strings.Join(parts, ",")
}
