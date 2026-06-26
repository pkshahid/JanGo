package queryset

import (
	"testing"

	"github.com/pkshahid/JanGo/orm"
)

type FuncTestModel struct {
	orm.Model
	Title   string  `gd:"CharField"`
	Views   int     `gd:"IntegerField"`
	Likes   int     `gd:"IntegerField"`
	Price   float64 `gd:"FloatField"`
	Content string  `gd:"TextField"`
}

func (m *FuncTestModel) ModelMeta() *orm.Meta {
	return &orm.Meta{DbTable: "functestmodel"}
}

func setupFuncTestModel(t *testing.T) QuerySet[FuncTestModel] {
	orm.ClearRegistry()
	orm.Register(&FuncTestModel{})
	return NewQuerySet[FuncTestModel]()
}

// ---------------------------------------------------------------------------
// Text functions
// ---------------------------------------------------------------------------

func TestLower(t *testing.T) {
	qs := setupFuncTestModel(t)

	qs = qs.Filter(Lookup{"Title__exact": Lower(FRef("Title"))})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM functestmodel WHERE (title = LOWER(title))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 0 {
		t.Errorf("Expected 0 params, got %v", params)
	}
}

func TestUpper(t *testing.T) {
	qs := setupFuncTestModel(t)

	qs = qs.Filter(Lookup{"Title__exact": Upper(FRef("Content"))})
	sql, _ := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM functestmodel WHERE (title = UPPER(content))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
}

func TestLowerWithLiteral(t *testing.T) {
	qs := setupFuncTestModel(t)

	qs = qs.Filter(Lookup{"Title__exact": Lower("HELLO")})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM functestmodel WHERE (title = LOWER(?))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 1 || params[0] != "HELLO" {
		t.Errorf("Expected params [HELLO], got %v", params)
	}
}

func TestLength(t *testing.T) {
	qs := setupFuncTestModel(t)

	qs = qs.Filter(Lookup{"Views__gt": Length(FRef("Title"))})
	sql, _ := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM functestmodel WHERE (views > LENGTH(title))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
}

func TestConcat(t *testing.T) {
	qs := setupFuncTestModel(t)

	qs = qs.Filter(Lookup{"Title__exact": Concat(FRef("Title"), " - ", FRef("Content"))})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM functestmodel WHERE (title = CONCAT(title, ?, content))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 1 || params[0] != " - " {
		t.Errorf("Expected params [ - ], got %v", params)
	}
}

func TestSubstr(t *testing.T) {
	qs := setupFuncTestModel(t)

	qs = qs.Filter(Lookup{"Title__exact": Substr(FRef("Content"), 1, 10)})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM functestmodel WHERE (title = SUBSTR(content, ?, ?))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 2 || params[0] != 1 || params[1] != 10 {
		t.Errorf("Expected params [1, 10], got %v", params)
	}
}

func TestSubstrNoLength(t *testing.T) {
	f := Substr(FRef("Title"), 5)
	sql := f.ToSQL()
	expected := "SUBSTR(Title, ?)"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

func TestReplace(t *testing.T) {
	qs := setupFuncTestModel(t)

	qs = qs.Filter(Lookup{"Title__exact": Replace(FRef("Content"), "old", "new")})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM functestmodel WHERE (title = REPLACE(content, ?, ?))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 2 || params[0] != "old" || params[1] != "new" {
		t.Errorf("Expected params [old, new], got %v", params)
	}
}

func TestTrim(t *testing.T) {
	f := Trim(FRef("Title"))
	sql := f.ToSQL()
	if sql != "TRIM(Title)" {
		t.Errorf("Expected 'TRIM(Title)', got %q", sql)
	}
}

func TestLTrim(t *testing.T) {
	f := LTrim(FRef("Title"))
	sql := f.ToSQL()
	if sql != "LTRIM(Title)" {
		t.Errorf("Expected 'LTRIM(Title)', got %q", sql)
	}
}

func TestRTrim(t *testing.T) {
	f := RTrim(FRef("Title"))
	sql := f.ToSQL()
	if sql != "RTRIM(Title)" {
		t.Errorf("Expected 'RTRIM(Title)', got %q", sql)
	}
}

func TestLeft(t *testing.T) {
	f := Left(FRef("Title"), 5)
	sql, params := f.ResolveSQL(nil)
	if sql != "LEFT(Title, ?)" {
		t.Errorf("Expected 'LEFT(Title, ?)', got %q", sql)
	}
	if len(params) != 1 || params[0] != 5 {
		t.Errorf("Expected params [5], got %v", params)
	}
}

func TestRight(t *testing.T) {
	f := Right(FRef("Title"), 3)
	sql := f.ToSQL()
	if sql != "RIGHT(Title, ?)" {
		t.Errorf("Expected 'RIGHT(Title, ?)', got %q", sql)
	}
}

func TestReverse(t *testing.T) {
	f := Reverse(FRef("Title"))
	sql := f.ToSQL()
	if sql != "REVERSE(Title)" {
		t.Errorf("Expected 'REVERSE(Title)', got %q", sql)
	}
}

func TestStrIndex(t *testing.T) {
	f := StrIndex(FRef("Content"), "search")
	sql, params := f.ResolveSQL(nil)
	if sql != "INSTR(Content, ?)" {
		t.Errorf("Expected 'INSTR(Content, ?)', got %q", sql)
	}
	if len(params) != 1 || params[0] != "search" {
		t.Errorf("Expected params [search], got %v", params)
	}
}

func TestRepeat(t *testing.T) {
	f := Repeat(FRef("Title"), 3)
	sql := f.ToSQL()
	if sql != "REPEAT(Title, ?)" {
		t.Errorf("Expected 'REPEAT(Title, ?)', got %q", sql)
	}
}

func TestChr(t *testing.T) {
	f := Chr(65)
	sql := f.ToSQL()
	if sql != "CHR(?)" {
		t.Errorf("Expected 'CHR(?)', got %q", sql)
	}
}

func TestAscii(t *testing.T) {
	f := Ascii(FRef("Title"))
	sql := f.ToSQL()
	if sql != "ASCII(Title)" {
		t.Errorf("Expected 'ASCII(Title)', got %q", sql)
	}
}

// ---------------------------------------------------------------------------
// Math functions
// ---------------------------------------------------------------------------

func TestAbs(t *testing.T) {
	qs := setupFuncTestModel(t)

	qs = qs.Filter(Lookup{"Views__exact": Abs(FRef("Likes"))})
	sql, _ := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM functestmodel WHERE (views = ABS(likes))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
}

func TestRound(t *testing.T) {
	f := Round(FRef("Price"), 2)
	sql, params := f.ResolveSQL(nil)
	if sql != "ROUND(Price, ?)" {
		t.Errorf("Expected 'ROUND(Price, ?)', got %q", sql)
	}
	if len(params) != 1 || params[0] != 2 {
		t.Errorf("Expected params [2], got %v", params)
	}
}

func TestRoundNoPrecision(t *testing.T) {
	f := Round(FRef("Price"))
	sql := f.ToSQL()
	if sql != "ROUND(Price)" {
		t.Errorf("Expected 'ROUND(Price)', got %q", sql)
	}
}

func TestPower(t *testing.T) {
	f := Power(FRef("Views"), 2)
	sql, params := f.ResolveSQL(nil)
	if sql != "POWER(Views, ?)" {
		t.Errorf("Expected 'POWER(Views, ?)', got %q", sql)
	}
	if len(params) != 1 || params[0] != 2 {
		t.Errorf("Expected params [2], got %v", params)
	}
}

func TestSqrt(t *testing.T) {
	f := Sqrt(FRef("Views"))
	sql := f.ToSQL()
	if sql != "SQRT(Views)" {
		t.Errorf("Expected 'SQRT(Views)', got %q", sql)
	}
}

func TestModFunc(t *testing.T) {
	f := ModFunc(FRef("Views"), 10)
	sql, params := f.ResolveSQL(nil)
	if sql != "MOD(Views, ?)" {
		t.Errorf("Expected 'MOD(Views, ?)', got %q", sql)
	}
	if len(params) != 1 || params[0] != 10 {
		t.Errorf("Expected params [10], got %v", params)
	}
}

func TestFloor(t *testing.T) {
	f := Floor(FRef("Price"))
	sql := f.ToSQL()
	if sql != "FLOOR(Price)" {
		t.Errorf("Expected 'FLOOR(Price)', got %q", sql)
	}
}

func TestCeil(t *testing.T) {
	f := Ceil(FRef("Price"))
	sql := f.ToSQL()
	if sql != "CEIL(Price)" {
		t.Errorf("Expected 'CEIL(Price)', got %q", sql)
	}
}

func TestSign(t *testing.T) {
	f := Sign(FRef("Views"))
	sql := f.ToSQL()
	if sql != "SIGN(Views)" {
		t.Errorf("Expected 'SIGN(Views)', got %q", sql)
	}
}

func TestExp(t *testing.T) {
	f := Exp(FRef("Views"))
	sql := f.ToSQL()
	if sql != "EXP(Views)" {
		t.Errorf("Expected 'EXP(Views)', got %q", sql)
	}
}

func TestLn(t *testing.T) {
	f := Ln(FRef("Views"))
	sql := f.ToSQL()
	if sql != "LN(Views)" {
		t.Errorf("Expected 'LN(Views)', got %q", sql)
	}
}

func TestDegrees(t *testing.T) {
	f := Degrees(FRef("Views"))
	sql := f.ToSQL()
	if sql != "DEGREES(Views)" {
		t.Errorf("Expected 'DEGREES(Views)', got %q", sql)
	}
}

func TestRadians(t *testing.T) {
	f := Radians(FRef("Views"))
	sql := f.ToSQL()
	if sql != "RADIANS(Views)" {
		t.Errorf("Expected 'RADIANS(Views)', got %q", sql)
	}
}

func TestPi(t *testing.T) {
	f := Pi()
	sql := f.ToSQL()
	if sql != "PI()" {
		t.Errorf("Expected 'PI()', got %q", sql)
	}
}

func TestRand(t *testing.T) {
	f := Rand()
	sql := f.ToSQL()
	if sql != "RAND()" {
		t.Errorf("Expected 'RAND()', got %q", sql)
	}
}

func TestGreatest(t *testing.T) {
	f := Greatest(FRef("Views"), FRef("Likes"), 0)
	sql, params := f.ResolveSQL(nil)
	if sql != "GREATEST(Views, Likes, ?)" {
		t.Errorf("Expected 'GREATEST(Views, Likes, ?)', got %q", sql)
	}
	if len(params) != 1 || params[0] != 0 {
		t.Errorf("Expected params [0], got %v", params)
	}
}

func TestLeast(t *testing.T) {
	f := Least(FRef("Views"), FRef("Likes"))
	sql := f.ToSQL()
	if sql != "LEAST(Views, Likes)" {
		t.Errorf("Expected 'LEAST(Views, Likes)', got %q", sql)
	}
}

func TestTrigFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       *Func
		expected string
	}{
		{"Cos", Cos(FRef("Price")), "COS(Price)"},
		{"Sin", Sin(FRef("Price")), "SIN(Price)"},
		{"Tan", Tan(FRef("Price")), "TAN(Price)"},
		{"Cot", Cot(FRef("Price")), "COT(Price)"},
		{"Acos", Acos(FRef("Price")), "ACOS(Price)"},
		{"Asin", Asin(FRef("Price")), "ASIN(Price)"},
		{"Atan", Atan(FRef("Price")), "ATAN(Price)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql := tt.fn.ToSQL()
			if sql != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, sql)
			}
		})
	}
}

func TestAtan2(t *testing.T) {
	f := Atan2(FRef("Views"), FRef("Likes"))
	sql := f.ToSQL()
	if sql != "ATAN2(Views, Likes)" {
		t.Errorf("Expected 'ATAN2(Views, Likes)', got %q", sql)
	}
}

// ---------------------------------------------------------------------------
// Date / Time functions
// ---------------------------------------------------------------------------

func TestNow(t *testing.T) {
	f := Now()
	sql := f.ToSQL()
	if sql != "NOW()" {
		t.Errorf("Expected 'NOW()', got %q", sql)
	}
}

func TestCurrentDate(t *testing.T) {
	f := CurrentDate()
	sql := f.ToSQL()
	if sql != "CURRENT_DATE()" {
		t.Errorf("Expected 'CURRENT_DATE()', got %q", sql)
	}
}

func TestCurrentTime(t *testing.T) {
	f := CurrentTime()
	sql := f.ToSQL()
	if sql != "CURRENT_TIME()" {
		t.Errorf("Expected 'CURRENT_TIME()', got %q", sql)
	}
}

func TestYear(t *testing.T) {
	f := Year(FRef("CreatedAt"))
	sql := f.ToSQL()
	if sql != "YEAR(CreatedAt)" {
		t.Errorf("Expected 'YEAR(CreatedAt)', got %q", sql)
	}
}

func TestMonth(t *testing.T) {
	f := Month(FRef("CreatedAt"))
	sql := f.ToSQL()
	if sql != "MONTH(CreatedAt)" {
		t.Errorf("Expected 'MONTH(CreatedAt)', got %q", sql)
	}
}

func TestDay(t *testing.T) {
	f := Day(FRef("CreatedAt"))
	sql := f.ToSQL()
	if sql != "DAY(CreatedAt)" {
		t.Errorf("Expected 'DAY(CreatedAt)', got %q", sql)
	}
}

func TestHour(t *testing.T) {
	f := Hour(FRef("CreatedAt"))
	sql := f.ToSQL()
	if sql != "HOUR(CreatedAt)" {
		t.Errorf("Expected 'HOUR(CreatedAt)', got %q", sql)
	}
}

func TestMinute(t *testing.T) {
	f := Minute(FRef("CreatedAt"))
	sql := f.ToSQL()
	if sql != "MINUTE(CreatedAt)" {
		t.Errorf("Expected 'MINUTE(CreatedAt)', got %q", sql)
	}
}

func TestSecond(t *testing.T) {
	f := Second(FRef("CreatedAt"))
	sql := f.ToSQL()
	if sql != "SECOND(CreatedAt)" {
		t.Errorf("Expected 'SECOND(CreatedAt)', got %q", sql)
	}
}

func TestWeek(t *testing.T) {
	f := Week(FRef("CreatedAt"))
	sql := f.ToSQL()
	if sql != "WEEK(CreatedAt)" {
		t.Errorf("Expected 'WEEK(CreatedAt)', got %q", sql)
	}
}

func TestQuarter(t *testing.T) {
	f := Quarter(FRef("CreatedAt"))
	sql := f.ToSQL()
	if sql != "QUARTER(CreatedAt)" {
		t.Errorf("Expected 'QUARTER(CreatedAt)', got %q", sql)
	}
}

func TestDayName(t *testing.T) {
	f := DayName(FRef("CreatedAt"))
	sql := f.ToSQL()
	if sql != "DAYNAME(CreatedAt)" {
		t.Errorf("Expected 'DAYNAME(CreatedAt)', got %q", sql)
	}
}

func TestMonthName(t *testing.T) {
	f := MonthName(FRef("CreatedAt"))
	sql := f.ToSQL()
	if sql != "MONTHNAME(CreatedAt)" {
		t.Errorf("Expected 'MONTHNAME(CreatedAt)', got %q", sql)
	}
}

func TestDayOfYear(t *testing.T) {
	f := DayOfYear(FRef("CreatedAt"))
	sql := f.ToSQL()
	if sql != "DAYOFYEAR(CreatedAt)" {
		t.Errorf("Expected 'DAYOFYEAR(CreatedAt)', got %q", sql)
	}
}

func TestDateDiff(t *testing.T) {
	f := DateDiff(FRef("CreatedAt"), FRef("UpdatedAt"))
	sql := f.ToSQL()
	if sql != "DATEDIFF(CreatedAt, UpdatedAt)" {
		t.Errorf("Expected 'DATEDIFF(CreatedAt, UpdatedAt)', got %q", sql)
	}
}

func TestLastDay(t *testing.T) {
	f := LastDay(FRef("CreatedAt"))
	sql := f.ToSQL()
	if sql != "LAST_DAY(CreatedAt)" {
		t.Errorf("Expected 'LAST_DAY(CreatedAt)', got %q", sql)
	}
}

// ---------------------------------------------------------------------------
// Extract
// ---------------------------------------------------------------------------

func TestExtract(t *testing.T) {
	f := NewExtract(FRef("CreatedAt"), "YEAR")
	sql, params := f.ResolveSQL(nil)
	if sql != "EXTRACT(YEAR FROM CreatedAt)" {
		t.Errorf("Expected 'EXTRACT(YEAR FROM CreatedAt)', got %q", sql)
	}
	if len(params) != 0 {
		t.Errorf("Expected 0 params, got %v", params)
	}
}

func TestExtractWithLiteral(t *testing.T) {
	f := NewExtract("2024-01-15", "MONTH")
	sql, params := f.ResolveSQL(nil)
	if sql != "EXTRACT(MONTH FROM ?)" {
		t.Errorf("Expected 'EXTRACT(MONTH FROM ?)', got %q", sql)
	}
	if len(params) != 1 || params[0] != "2024-01-15" {
		t.Errorf("Expected params [2024-01-15], got %v", params)
	}
}

func TestExtractToSQL(t *testing.T) {
	f := NewExtract(FRef("CreatedAt"), "HOUR")
	sql := f.ToSQL()
	if sql != "EXTRACT(HOUR FROM CreatedAt)" {
		t.Errorf("Expected 'EXTRACT(HOUR FROM CreatedAt)', got %q", sql)
	}
}

// ---------------------------------------------------------------------------
// Cast
// ---------------------------------------------------------------------------

func TestCast(t *testing.T) {
	qs := setupFuncTestModel(t)

	qs = qs.Filter(Lookup{"Views__exact": NewCast(FRef("Price"), "INTEGER")})
	sql, _ := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM functestmodel WHERE (views = CAST(price AS INTEGER))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
}

func TestCastWithLiteral(t *testing.T) {
	f := NewCast("123", "INTEGER")
	sql, params := f.ResolveSQL(nil)
	if sql != "CAST(? AS INTEGER)" {
		t.Errorf("Expected 'CAST(? AS INTEGER)', got %q", sql)
	}
	if len(params) != 1 || params[0] != "123" {
		t.Errorf("Expected params [123], got %v", params)
	}
}

func TestCastToSQL(t *testing.T) {
	f := NewCast(FRef("Price"), "TEXT")
	sql := f.ToSQL()
	if sql != "CAST(Price AS TEXT)" {
		t.Errorf("Expected 'CAST(Price AS TEXT)', got %q", sql)
	}
}

// ---------------------------------------------------------------------------
// Coalesce
// ---------------------------------------------------------------------------

func TestCoalesce(t *testing.T) {
	qs := setupFuncTestModel(t)

	qs = qs.Filter(Lookup{"Title__exact": Coalesce(FRef("Content"), FRef("Title"), "default")})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM functestmodel WHERE (title = COALESCE(content, title, ?))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 1 || params[0] != "default" {
		t.Errorf("Expected params [default], got %v", params)
	}
}

func TestCoalesceAllExpressions(t *testing.T) {
	f := Coalesce(FRef("Title"), FRef("Content"))
	sql := f.ToSQL()
	if sql != "COALESCE(Title, Content)" {
		t.Errorf("Expected 'COALESCE(Title, Content)', got %q", sql)
	}
}

// ---------------------------------------------------------------------------
// NullIf & IfNull
// ---------------------------------------------------------------------------

func TestNullIf(t *testing.T) {
	f := NullIf(FRef("Views"), 0)
	sql, params := f.ResolveSQL(nil)
	if sql != "NULLIF(Views, ?)" {
		t.Errorf("Expected 'NULLIF(Views, ?)', got %q", sql)
	}
	if len(params) != 1 || params[0] != 0 {
		t.Errorf("Expected params [0], got %v", params)
	}
}

func TestIfNull(t *testing.T) {
	f := IfNull(FRef("Title"), "N/A")
	sql, params := f.ResolveSQL(nil)
	if sql != "IFNULL(Title, ?)" {
		t.Errorf("Expected 'IFNULL(Title, ?)', got %q", sql)
	}
	if len(params) != 1 || params[0] != "N/A" {
		t.Errorf("Expected params [N/A], got %v", params)
	}
}

// ---------------------------------------------------------------------------
// IIf
// ---------------------------------------------------------------------------

func TestIIf(t *testing.T) {
	f := NewIIf(Q(Lookup{"Views__gt": 100}), "popular", "normal")
	sql, params := f.ResolveSQL(nil)

	if sql != "IIF(Views > ?, ?, ?)" {
		t.Errorf("Expected 'IIF(Views > ?, ?, ?)', got %q", sql)
	}
	if len(params) != 3 {
		t.Errorf("Expected 3 params, got %v", params)
	}
	if params[0] != 100 || params[1] != "popular" || params[2] != "normal" {
		t.Errorf("Expected params [100, popular, normal], got %v", params)
	}
}

func TestIIfWithExpressions(t *testing.T) {
	f := NewIIf(Q(Lookup{"Views__gt": 0}), FRef("Title"), FRef("Content"))
	sql := f.ToSQL()
	if sql != "IIF(Views > ?, Title, Content)" {
		t.Errorf("Expected 'IIF(Views > ?, Title, Content)', got %q", sql)
	}
}

func TestIIfInAnnotate(t *testing.T) {
	qs := setupFuncTestModel(t)

	qs = qs.Annotate(NewIIf(Q(Lookup{"Views__gt": 100}), "popular", "normal"))
	if len(qs.query.Annotations) != 1 {
		t.Fatalf("Expected 1 annotation, got %d", len(qs.query.Annotations))
	}
	sql := qs.query.Annotations[0].ToSQL()
	if sql != "IIF(Views > ?, ?, ?)" {
		t.Errorf("Expected 'IIF(Views > ?, ?, ?)', got %q", sql)
	}
}

// ---------------------------------------------------------------------------
// StringAgg
// ---------------------------------------------------------------------------

func TestStringAgg(t *testing.T) {
	f := NewStringAgg(FRef("Title"), ", ")
	sql, params := f.ResolveSQL(nil)
	if sql != "GROUP_CONCAT(Title SEPARATOR ?)" {
		t.Errorf("Expected 'GROUP_CONCAT(Title SEPARATOR ?)', got %q", sql)
	}
	if len(params) != 1 || params[0] != ", " {
		t.Errorf("Expected params [, ], got %v", params)
	}
}

func TestStringAggDistinct(t *testing.T) {
	f := NewStringAgg(FRef("Title"), ", ").Distinct()
	sql := f.ToSQL()
	if sql != "GROUP_CONCAT(DISTINCT Title SEPARATOR ?)" {
		t.Errorf("Expected 'GROUP_CONCAT(DISTINCT Title SEPARATOR ?)', got %q", sql)
	}
}

// ---------------------------------------------------------------------------
// Integration: functions in Annotate, Update, Filter
// ---------------------------------------------------------------------------

func TestFunctionInAnnotate(t *testing.T) {
	qs := setupFuncTestModel(t)

	qs = qs.Annotate(Lower(FRef("Title")))
	if len(qs.query.Annotations) != 1 {
		t.Fatalf("Expected 1 annotation")
	}
	sql := qs.query.Annotations[0].ToSQL()
	if sql != "LOWER(Title)" {
		t.Errorf("Expected 'LOWER(Title)', got %q", sql)
	}
}

func TestFunctionInUpdate(t *testing.T) {
	qs := setupFuncTestModel(t)

	sql, params := qs.query.ToUpdateSQL(map[string]any{
		"Title": Upper(FRef("Content")),
	})

	if sql != "UPDATE functestmodel SET title = UPPER(content)" {
		t.Errorf("Expected 'UPDATE functestmodel SET title = UPPER(content)', got %q", sql)
	}
	if len(params) != 0 {
		t.Errorf("Expected 0 params, got %v", params)
	}
}

func TestNestedFunctions(t *testing.T) {
	qs := setupFuncTestModel(t)

	qs = qs.Filter(Lookup{"Title__exact": Upper(Lower(FRef("Content")))})
	sql, _ := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM functestmodel WHERE (title = UPPER(LOWER(content)))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
}

func TestFunctionWithFArithmetic(t *testing.T) {
	qs := setupFuncTestModel(t)

	qs = qs.Filter(Lookup{"Views__exact": Abs(FRef("Views").Sub(FRef("Likes")))})
	sql, _ := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM functestmodel WHERE (views = ABS((views - likes)))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
}

func TestFunctionInAggregate(t *testing.T) {
	qs := setupFuncTestModel(t)

	res, err := qs.Aggregate(Coalesce("a", "b"))
	_ = res
	_ = err
	// This just verifies Coalesce result can be used as AggExpr (it implements ToSQL)
}

func TestNewFuncGeneric(t *testing.T) {
	f := NewFunc("CUSTOM_FUNC", FRef("Title"), 42, "hello")
	sql, params := f.ResolveSQL(nil)
	if sql != "CUSTOM_FUNC(Title, ?, ?)" {
		t.Errorf("Expected 'CUSTOM_FUNC(Title, ?, ?)', got %q", sql)
	}
	if len(params) != 2 || params[0] != 42 || params[1] != "hello" {
		t.Errorf("Expected params [42, hello], got %v", params)
	}
}

func TestFunctionResolvesColumnName(t *testing.T) {
	info, err := orm.GetModelInfo(&FuncTestModel{})
	if err != nil {
		t.Fatalf("Failed to get model info: %v", err)
	}

	f := Lower(FRef("Title"))
	sql, params := f.ResolveSQL(info)
	if sql != "LOWER(title)" {
		t.Errorf("Expected 'LOWER(title)', got %q", sql)
	}
	if len(params) != 0 {
		t.Errorf("Expected 0 params, got %v", params)
	}
}
