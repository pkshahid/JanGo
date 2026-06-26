package queryset

import (
	"testing"

	"github.com/pkshahid/JanGo/orm"
)

func setupCaseTestModel(t *testing.T) QuerySet[CaseTestModel] {
	orm.ClearRegistry()
	orm.Register(&CaseTestModel{})
	return NewQuerySet[CaseTestModel]()
}

type CaseTestModel struct {
	orm.Model
	Name   string `gd:"CharField"`
	Age    int    `gd:"IntegerField"`
	Status string `gd:"CharField"`
}

func TestCaseBasicLiteralThen(t *testing.T) {
	qs := setupCaseTestModel(t)

	c := NewCase(
		NewWhen(Q(Lookup{"Age__gte": 18}), "adult"),
		NewWhen(Q(Lookup{"Age__lt": 18}), "minor"),
	)

	qs = qs.Annotate(c)
	sql, _ := qs.query.ToSQL()

	// Annotate stores in Annotations but ToSQL for SELECT doesn't render them yet.
	// Verify via the expression's own ResolveSQL.
	exprSQL, exprParams := c.ResolveSQL(qs.query.ModelInfo)

	expectedSQL := "CASE WHEN age >= ? THEN ? WHEN age < ? THEN ? END"
	if exprSQL != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, exprSQL)
	}

	expectedParams := []any{18, "adult", 18, "minor"}
	if len(exprParams) != len(expectedParams) {
		t.Errorf("Expected %d params, got %d: %v", len(expectedParams), len(exprParams), exprParams)
	} else {
		for i, p := range expectedParams {
			if exprParams[i] != p {
				t.Errorf("Param %d: expected %v, got %v", i, p, exprParams[i])
			}
		}
	}

	// Ensure ToSQL doesn't panic and returns non-empty
	if c.ToSQL() == "" {
		t.Error("ToSQL returned empty string")
	}

	_ = sql
}

func TestCaseWithDefault(t *testing.T) {
	qs := setupCaseTestModel(t)

	c := NewCase(
		NewWhen(Q(Lookup{"Age__gte": 65}), "senior"),
	).Default("non-senior")

	exprSQL, exprParams := c.ResolveSQL(qs.query.ModelInfo)

	expectedSQL := "CASE WHEN age >= ? THEN ? ELSE ? END"
	if exprSQL != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, exprSQL)
	}

	expectedParams := []any{65, "senior", "non-senior"}
	if len(exprParams) != len(expectedParams) {
		t.Errorf("Expected %d params, got %d: %v", len(expectedParams), len(exprParams), exprParams)
	}
}

func TestCaseWithFExpressionThen(t *testing.T) {
	qs := setupCaseTestModel(t)

	c := NewCase(
		NewWhen(Q(Lookup{"Age__gt": 21}), FRef("Age").Add(10)),
		NewWhen(Q(Lookup{"Age__lte": 21}), FRef("Age")),
	).Default(0)

	exprSQL, exprParams := c.ResolveSQL(qs.query.ModelInfo)

	expectedSQL := "CASE WHEN age > ? THEN (age + ?) WHEN age <= ? THEN age ELSE ? END"
	if exprSQL != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, exprSQL)
	}

	expectedParams := []any{21, 10, 21, 0}
	if len(exprParams) != len(expectedParams) {
		t.Errorf("Expected %d params, got %d: %v", len(expectedParams), len(exprParams), exprParams)
	}
}

func TestCaseInFilter(t *testing.T) {
	qs := setupCaseTestModel(t)

	c := NewCase(
		NewWhen(Q(Lookup{"Age__gte": 18}), "adult"),
		NewWhen(Q(Lookup{"Age__lt": 18}), "minor"),
	).Default("unknown")

	qs = qs.Filter(Lookup{"Status__exact": c})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM casetestmodel WHERE (status = CASE WHEN age >= ? THEN ? WHEN age < ? THEN ? ELSE ? END)"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}

	expectedParams := []any{18, "adult", 18, "minor", "unknown"}
	if len(params) != len(expectedParams) {
		t.Errorf("Expected %d params, got %d: %v", len(expectedParams), len(params), params)
	}
}

func TestCaseInUpdate(t *testing.T) {
	qs := setupCaseTestModel(t)

	c := NewCase(
		NewWhen(Q(Lookup{"Age__gte": 18}), "adult"),
		NewWhen(Q(Lookup{"Age__lt": 18}), "minor"),
	).Default("unknown")

	qs = qs.Filter(Lookup{"Name__exact": "Alice"})
	sql, params := qs.query.ToUpdateSQL(map[string]any{"Status": c})

	expectedSQL := "UPDATE casetestmodel SET status = CASE WHEN age >= ? THEN ? WHEN age < ? THEN ? ELSE ? END WHERE (name = ?)"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}

	expectedParams := []any{18, "adult", 18, "minor", "unknown", "Alice"}
	if len(params) != len(expectedParams) {
		t.Errorf("Expected %d params, got %d: %v", len(expectedParams), len(params), params)
	}
}

func TestCaseWithQOrCondition(t *testing.T) {
	qs := setupCaseTestModel(t)

	c := NewCase(
		NewWhen(
			Q(Lookup{"Age__lt": 13}).Or(Q(Lookup{"Age__gt": 65})),
			"special",
		),
	).Default("standard")

	exprSQL, exprParams := c.ResolveSQL(qs.query.ModelInfo)

	expectedSQL := "CASE WHEN (age < ?) OR (age > ?) THEN ? ELSE ? END"
	if exprSQL != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, exprSQL)
	}

	expectedParams := []any{13, 65, "special", "standard"}
	if len(exprParams) != len(expectedParams) {
		t.Errorf("Expected %d params, got %d: %v", len(expectedParams), len(exprParams), exprParams)
	}
}

func TestCaseOutputField(t *testing.T) {
	qs := setupCaseTestModel(t)

	c := NewCase(
		NewWhen(Q(Lookup{"Age__gte": 18}), 1),
		NewWhen(Q(Lookup{"Age__lt": 18}), 0),
	).OutputField("INTEGER")

	exprSQL, _ := c.ResolveSQL(qs.query.ModelInfo)

	expectedSQL := "CAST(CASE WHEN age >= ? THEN ? WHEN age < ? THEN ? END AS INTEGER)"
	if exprSQL != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, exprSQL)
	}
}

func TestCaseNoDefault(t *testing.T) {
	qs := setupCaseTestModel(t)

	c := NewCase(
		NewWhen(Q(Lookup{"Age__gte": 18}), "adult"),
	)

	exprSQL, _ := c.ResolveSQL(qs.query.ModelInfo)

	expectedSQL := "CASE WHEN age >= ? THEN ? END"
	if exprSQL != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, exprSQL)
	}
}

func TestCaseImmutability(t *testing.T) {
	c := NewCase(
		NewWhen(Q(Lookup{"Age__gte": 18}), "adult"),
	)

	cWithDefault := c.Default("minor")

	// Original should NOT have a default
	originalSQL, _ := c.ResolveSQL(nil)
	if contains(originalSQL, "ELSE") {
		t.Errorf("Original Case should not have ELSE, got: %s", originalSQL)
	}

	// Clone should have default
	cloneSQL, _ := cWithDefault.ResolveSQL(nil)
	if !contains(cloneSQL, "ELSE") {
		t.Errorf("Clone should have ELSE, got: %s", cloneSQL)
	}
}

func TestCaseAggExprInterface(t *testing.T) {
	c := NewCase(
		NewWhen(Q(Lookup{"Age__gte": 18}), "adult"),
		NewWhen(Q(Lookup{"Age__lt": 18}), "minor"),
	)

	// Verify it satisfies AggExpr
	var _ AggExpr = c

	// Verify it satisfies Expression
	var _ Expression = c

	// ToSQL should work
	sql := c.ToSQL()
	if sql == "" {
		t.Error("ToSQL returned empty string")
	}
	if !contains(sql, "CASE") || !contains(sql, "WHEN") || !contains(sql, "THEN") {
		t.Errorf("ToSQL should contain CASE/WHEN/THEN, got: %s", sql)
	}
}

func TestCaseEmpty(t *testing.T) {
	c := NewCase()

	exprSQL, _ := c.ResolveSQL(nil)
	expectedSQL := "CASE  END"
	if exprSQL != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, exprSQL)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
