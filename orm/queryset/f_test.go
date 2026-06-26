package queryset

import (
	"testing"

	"github.com/pkshahid/JanGo/orm"
)

type FTestModel struct {
	orm.Model
	Title string `gd:"CharField"`
	Views int    `gd:"IntegerField"`
	Likes int    `gd:"IntegerField"`
}

func setupFTestModel(t *testing.T) QuerySet[FTestModel] {
	orm.ClearRegistry()
	orm.Register(&FTestModel{})
	return NewQuerySet[FTestModel]()
}

func TestFSimpleFieldReference(t *testing.T) {
	qs := setupFTestModel(t)

	qs = qs.Filter(Lookup{"Views": FRef("Likes")})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM ftestmodel WHERE (views = likes)"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 0 {
		t.Errorf("Expected 0 params, got %v", params)
	}
}

func TestFInGtLookup(t *testing.T) {
	qs := setupFTestModel(t)

	qs = qs.Filter(Lookup{"Views__gt": FRef("Likes")})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM ftestmodel WHERE (views > likes)"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 0 {
		t.Errorf("Expected 0 params, got %v", params)
	}
}

func TestFAddLiteral(t *testing.T) {
	qs := setupFTestModel(t)

	qs = qs.Filter(Lookup{"Views__gt": FRef("Likes").Add(10)})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM ftestmodel WHERE (views > (likes + ?))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 1 || params[0] != 10 {
		t.Errorf("Expected params [10], got %v", params)
	}
}

func TestFSubLiteral(t *testing.T) {
	qs := setupFTestModel(t)

	qs = qs.Filter(Lookup{"Views__gte": FRef("Likes").Sub(5)})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM ftestmodel WHERE (views >= (likes - ?))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 1 || params[0] != 5 {
		t.Errorf("Expected params [5], got %v", params)
	}
}

func TestFMulLiteral(t *testing.T) {
	qs := setupFTestModel(t)

	qs = qs.Filter(Lookup{"Views__lt": FRef("Likes").Mul(2)})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM ftestmodel WHERE (views < (likes * ?))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 1 || params[0] != 2 {
		t.Errorf("Expected params [2], got %v", params)
	}
}

func TestFDivLiteral(t *testing.T) {
	qs := setupFTestModel(t)

	qs = qs.Filter(Lookup{"Views__lte": FRef("Likes").Div(2)})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM ftestmodel WHERE (views <= (likes / ?))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 1 || params[0] != 2 {
		t.Errorf("Expected params [2], got %v", params)
	}
}

func TestFModLiteral(t *testing.T) {
	qs := setupFTestModel(t)

	qs = qs.Filter(Lookup{"Views__exact": FRef("Likes").Mod(3)})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM ftestmodel WHERE (views = (likes % ?))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 1 || params[0] != 3 {
		t.Errorf("Expected params [3], got %v", params)
	}
}

func TestFAddF(t *testing.T) {
	qs := setupFTestModel(t)

	qs = qs.Filter(Lookup{"Views__gt": FRef("Likes").Add(FRef("Views"))})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM ftestmodel WHERE (views > (likes + views))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 0 {
		t.Errorf("Expected 0 params, got %v", params)
	}
}

func TestFChainedArithmetic(t *testing.T) {
	qs := setupFTestModel(t)

	qs = qs.Filter(Lookup{"Views__gt": FRef("Likes").Add(10).Mul(2)})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM ftestmodel WHERE (views > ((likes + ?) * ?))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 2 || params[0] != 10 || params[1] != 2 {
		t.Errorf("Expected params [10, 2], got %v", params)
	}
}

func TestFUpdateSimple(t *testing.T) {
	qs := setupFTestModel(t)

	qs = qs.Filter(Lookup{"Title__exact": "hello"})
	sql, params := qs.query.ToUpdateSQL(map[string]any{"Views": FRef("Views").Add(1)})

	expectedSQL := "UPDATE ftestmodel SET views = (views + ?) WHERE (title = ?)"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 2 || params[0] != 1 || params[1] != "hello" {
		t.Errorf("Expected params [1, hello], got %v", params)
	}
}

func TestFUpdateMultipleFields(t *testing.T) {
	qs := setupFTestModel(t)

	sql, params := qs.query.ToUpdateSQL(map[string]any{
		"Views": FRef("Likes"),
		"Likes": FRef("Views").Add(5),
	})

	if sql != "UPDATE ftestmodel SET views = likes, likes = (views + ?)" {
		t.Errorf("Unexpected SQL: %q", sql)
	}
	if len(params) != 1 || params[0] != 5 {
		t.Errorf("Expected params [5], got %v", params)
	}
}

func TestFUpdateWithLiteral(t *testing.T) {
	qs := setupFTestModel(t)

	sql, params := qs.query.ToUpdateSQL(map[string]any{"Views": 42})

	expectedSQL := "UPDATE ftestmodel SET views = ?"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 1 || params[0] != 42 {
		t.Errorf("Expected params [42], got %v", params)
	}
}

func TestFUpdateWithWhereAndF(t *testing.T) {
	qs := setupFTestModel(t)

	qs = qs.Filter(Lookup{"Views__gt": FRef("Likes")})
	sql, params := qs.query.ToUpdateSQL(map[string]any{"Views": FRef("Views").Sub(1)})

	expectedSQL := "UPDATE ftestmodel SET views = (views - ?) WHERE (views > likes)"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if len(params) != 1 || params[0] != 1 {
		t.Errorf("Expected params [1], got %v", params)
	}
}

func TestFAnnotate(t *testing.T) {
	qs := setupFTestModel(t)

	qs = qs.Annotate(FRef("Views").Add(FRef("Likes")))

	if len(qs.query.Annotations) != 1 {
		t.Fatalf("Expected 1 annotation, got %d", len(qs.query.Annotations))
	}

	expectedSQL := "(Views + Likes)"
	if qs.query.Annotations[0].ToSQL() != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, qs.query.Annotations[0].ToSQL())
	}
}

func TestFResolvesColumnName(t *testing.T) {
	info, err := orm.GetModelInfo(&FTestModel{})
	if err != nil {
		t.Fatalf("Failed to get model info: %v", err)
	}

	expr := FRef("Views").Add(1)
	sql, params := expr.ResolveSQL(info)

	if sql != "(views + ?)" {
		t.Errorf("Expected '(views + ?)', got %q", sql)
	}
	if len(params) != 1 || params[0] != 1 {
		t.Errorf("Expected params [1], got %v", params)
	}
}

func TestFToSQLWithoutInfo(t *testing.T) {
	expr := FRef("Views").Add(FRef("Likes"))

	sql := expr.ToSQL()
	if sql != "(Views + Likes)" {
		t.Errorf("Expected '(Views + Likes)', got %q", sql)
	}
}
