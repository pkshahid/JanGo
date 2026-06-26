package queryset

import (
	"testing"
	"time"

	"github.com/pkshahid/JanGo/orm"
)

type DateTestModel struct {
	orm.Model
	Title     string    `gd:"CharField"`
	Published time.Time `gd:"DateTimeField"`
	Birthday  time.Time `gd:"DateField"`
}

func setupDateTestModel(t *testing.T) QuerySet[DateTestModel] {
	orm.ClearRegistry()
	orm.Register(&DateTestModel{})
	return NewQuerySet[DateTestModel]()
}

// ---------------------------------------------------------------------------
// ToDatesSQL — SQL generation
// ---------------------------------------------------------------------------

func TestToDatesSQLYear(t *testing.T) {
	qs := setupDateTestModel(t)
	sql, params := qs.query.ToDatesSQL(DateKindYear, "Published", OrderASC)

	expected := "SELECT DISTINCT DATE_TRUNC('year', published) AS date_field FROM datetestmodel ORDER BY date_field ASC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
	if len(params) != 0 {
		t.Errorf("Expected 0 params, got %v", params)
	}
}

func TestToDatesSQLMonth(t *testing.T) {
	qs := setupDateTestModel(t)
	sql, _ := qs.query.ToDatesSQL(DateKindMonth, "Published", OrderASC)

	expected := "SELECT DISTINCT DATE_TRUNC('month', published) AS date_field FROM datetestmodel ORDER BY date_field ASC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

func TestToDatesSQLDay(t *testing.T) {
	qs := setupDateTestModel(t)
	sql, _ := qs.query.ToDatesSQL(DateKindDay, "Birthday", OrderASC)

	expected := "SELECT DISTINCT DATE_TRUNC('day', birthday) AS date_field FROM datetestmodel ORDER BY date_field ASC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

func TestToDatesSQLWeek(t *testing.T) {
	qs := setupDateTestModel(t)
	sql, _ := qs.query.ToDatesSQL(DateKindWeek, "Published", OrderASC)

	expected := "SELECT DISTINCT DATE_TRUNC('week', published) AS date_field FROM datetestmodel ORDER BY date_field ASC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

func TestToDatesSQLQuarter(t *testing.T) {
	qs := setupDateTestModel(t)
	sql, _ := qs.query.ToDatesSQL(DateKindQuarter, "Published", OrderASC)

	expected := "SELECT DISTINCT DATE_TRUNC('quarter', published) AS date_field FROM datetestmodel ORDER BY date_field ASC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

func TestToDatesSQLDescOrder(t *testing.T) {
	qs := setupDateTestModel(t)
	sql, _ := qs.query.ToDatesSQL(DateKindYear, "Published", OrderDESC)

	expected := "SELECT DISTINCT DATE_TRUNC('year', published) AS date_field FROM datetestmodel ORDER BY date_field DESC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

func TestToDatesSQLDefaultOrder(t *testing.T) {
	qs := setupDateTestModel(t)
	sql, _ := qs.query.ToDatesSQL(DateKindYear, "Published", "")

	expected := "SELECT DISTINCT DATE_TRUNC('year', published) AS date_field FROM datetestmodel ORDER BY date_field ASC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

func TestToDatesSQLWithFilter(t *testing.T) {
	qs := setupDateTestModel(t)
	qs = qs.Filter(Lookup{"Title__exact": "hello"})
	sql, params := qs.query.ToDatesSQL(DateKindYear, "Published", OrderASC)

	expected := "SELECT DISTINCT DATE_TRUNC('year', published) AS date_field FROM datetestmodel WHERE (title = ?) ORDER BY date_field ASC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
	if len(params) != 1 || params[0] != "hello" {
		t.Errorf("Expected params [hello], got %v", params)
	}
}

func TestToDatesSQLWithFilterAndExclude(t *testing.T) {
	qs := setupDateTestModel(t)
	qs = qs.Filter(Lookup{"Title__exact": "hello"}).Exclude(Lookup{"Title__exact": "world"})
	sql, params := qs.query.ToDatesSQL(DateKindMonth, "Published", OrderDESC)

	expected := "SELECT DISTINCT DATE_TRUNC('month', published) AS date_field FROM datetestmodel WHERE (title = ?) AND NOT (title = ?) ORDER BY date_field DESC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
	if len(params) != 2 || params[0] != "hello" || params[1] != "world" {
		t.Errorf("Expected params [hello, world], got %v", params)
	}
}

func TestToDatesSQLInvalidKind(t *testing.T) {
	qs := setupDateTestModel(t)
	sql, _ := qs.query.ToDatesSQL(DateKind("hour"), "Published", OrderASC)
	if sql != "" {
		t.Errorf("Expected empty SQL for invalid kind, got %q", sql)
	}
}

// ---------------------------------------------------------------------------
// ToDatetimesSQL — SQL generation
// ---------------------------------------------------------------------------

func TestToDatetimesSQLYear(t *testing.T) {
	qs := setupDateTestModel(t)
	sql, params := qs.query.ToDatetimesSQL(DateTimeKindYear, "Published", OrderASC)

	expected := "SELECT DISTINCT DATETIME_TRUNC('year', published) AS datetime_field FROM datetestmodel ORDER BY datetime_field ASC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
	if len(params) != 0 {
		t.Errorf("Expected 0 params, got %v", params)
	}
}

func TestToDatetimesSQLHour(t *testing.T) {
	qs := setupDateTestModel(t)
	sql, _ := qs.query.ToDatetimesSQL(DateTimeKindHour, "Published", OrderASC)

	expected := "SELECT DISTINCT DATETIME_TRUNC('hour', published) AS datetime_field FROM datetestmodel ORDER BY datetime_field ASC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

func TestToDatetimesSQLMinute(t *testing.T) {
	qs := setupDateTestModel(t)
	sql, _ := qs.query.ToDatetimesSQL(DateTimeKindMinute, "Published", OrderASC)

	expected := "SELECT DISTINCT DATETIME_TRUNC('minute', published) AS datetime_field FROM datetestmodel ORDER BY datetime_field ASC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

func TestToDatetimesSQLSecond(t *testing.T) {
	qs := setupDateTestModel(t)
	sql, _ := qs.query.ToDatetimesSQL(DateTimeKindSecond, "Published", OrderDESC)

	expected := "SELECT DISTINCT DATETIME_TRUNC('second', published) AS datetime_field FROM datetestmodel ORDER BY datetime_field DESC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

func TestToDatetimesSQLDay(t *testing.T) {
	qs := setupDateTestModel(t)
	sql, _ := qs.query.ToDatetimesSQL(DateTimeKindDay, "Published", OrderASC)

	expected := "SELECT DISTINCT DATETIME_TRUNC('day', published) AS datetime_field FROM datetestmodel ORDER BY datetime_field ASC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

func TestToDatetimesSQLDefaultOrder(t *testing.T) {
	qs := setupDateTestModel(t)
	sql, _ := qs.query.ToDatetimesSQL(DateTimeKindYear, "Published", "")

	expected := "SELECT DISTINCT DATETIME_TRUNC('year', published) AS datetime_field FROM datetestmodel ORDER BY datetime_field ASC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

func TestToDatetimesSQLWithFilter(t *testing.T) {
	qs := setupDateTestModel(t)
	qs = qs.Filter(Lookup{"Title__exact": "test"})
	sql, params := qs.query.ToDatetimesSQL(DateTimeKindHour, "Published", OrderASC)

	expected := "SELECT DISTINCT DATETIME_TRUNC('hour', published) AS datetime_field FROM datetestmodel WHERE (title = ?) ORDER BY datetime_field ASC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
	if len(params) != 1 || params[0] != "test" {
		t.Errorf("Expected params [test], got %v", params)
	}
}

func TestToDatetimesSQLInvalidKind(t *testing.T) {
	qs := setupDateTestModel(t)
	sql, _ := qs.query.ToDatetimesSQL(DateTimeKind("week"), "Published", OrderASC)
	if sql != "" {
		t.Errorf("Expected empty SQL for invalid kind, got %q", sql)
	}
}

// ---------------------------------------------------------------------------
// Dates / Datetimes — terminal method execution (mocked)
// ---------------------------------------------------------------------------

func TestDatesExecution(t *testing.T) {
	qs := setupDateTestModel(t)

	results, err := qs.Dates("Published", DateKindYear)
	if err != nil {
		t.Fatalf("Dates() error: %v", err)
	}
	if results == nil {
		t.Errorf("Expected non-nil results slice")
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 mocked results, got %d", len(results))
	}
}

func TestDatesExecutionWithOrder(t *testing.T) {
	qs := setupDateTestModel(t)

	results, err := qs.Dates("Published", DateKindMonth, OrderDESC)
	if err != nil {
		t.Fatalf("Dates() error: %v", err)
	}
	if results == nil {
		t.Errorf("Expected non-nil results slice")
	}
}

func TestDatesInvalidKind(t *testing.T) {
	qs := setupDateTestModel(t)

	_, err := qs.Dates("Published", DateKind("hour"))
	if err == nil {
		t.Errorf("Expected error for invalid date kind")
	}
}

func TestDatetimesExecution(t *testing.T) {
	qs := setupDateTestModel(t)

	results, err := qs.Datetimes("Published", DateTimeKindHour)
	if err != nil {
		t.Fatalf("Datetimes() error: %v", err)
	}
	if results == nil {
		t.Errorf("Expected non-nil results slice")
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 mocked results, got %d", len(results))
	}
}

func TestDatetimesExecutionWithOrder(t *testing.T) {
	qs := setupDateTestModel(t)

	results, err := qs.Datetimes("Published", DateTimeKindMinute, OrderDESC)
	if err != nil {
		t.Fatalf("Datetimes() error: %v", err)
	}
	if results == nil {
		t.Errorf("Expected non-nil results slice")
	}
}

func TestDatetimesInvalidKind(t *testing.T) {
	qs := setupDateTestModel(t)

	_, err := qs.Datetimes("Published", DateTimeKind("week"))
	if err == nil {
		t.Errorf("Expected error for invalid datetime kind")
	}
}

// ---------------------------------------------------------------------------
// Immutability — Dates/Datetimes should not mutate the original QuerySet
// ---------------------------------------------------------------------------

func TestDatesImmutability(t *testing.T) {
	qs := setupDateTestModel(t)
	originalSQL, _ := qs.query.ToSQL()

	_, _ = qs.Dates("Published", DateKindYear)
	_, _ = qs.Datetimes("Published", DateTimeKindHour, OrderDESC)

	afterSQL, _ := qs.query.ToSQL()
	if originalSQL != afterSQL {
		t.Errorf("Dates/Datetimes mutated the original queryset: before=%q, after=%q", originalSQL, afterSQL)
	}
}
