package queryset

import (
	"reflect"
	"testing"

	"github.com/pkshahid/JanGo/orm"
)

func TestExtraSelect(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&TestArticle{})

	qs := NewQuerySet[TestArticle]()
	qs = qs.Extra(ExtraData{
		Select: map[string]string{
			"view_count": "SELECT COUNT(*) FROM test_article",
		},
	})

	sql, params := qs.query.ToSQL()
	expected := "SELECT *, (SELECT COUNT(*) FROM test_article) AS view_count FROM test_article"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
	if len(params) != 0 {
		t.Errorf("Expected 0 params, got %d", len(params))
	}
}

func TestExtraWhereWithParams(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&TestArticle{})

	qs := NewQuerySet[TestArticle]()
	qs = qs.Extra(ExtraData{
		Where:  []string{"title LIKE ?", "views > ?"},
		Params: []any{"%Go%", 50},
	})

	sql, params := qs.query.ToSQL()
	expected := "SELECT * FROM test_article WHERE title LIKE ? AND views > ?"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
	if !reflect.DeepEqual(params, []any{"%Go%", 50}) {
		t.Errorf("Params mismatch: got %v", params)
	}
}

func TestExtraTables(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&TestArticle{})

	qs := NewQuerySet[TestArticle]()
	qs = qs.Extra(ExtraData{
		Tables: []string{"other_table"},
	})

	sql, _ := qs.query.ToSQL()
	expected := "SELECT * FROM test_article, other_table"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

func TestExtraOrderBy(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&TestArticle{})

	qs := NewQuerySet[TestArticle]()
	qs = qs.OrderBy("-Views").Extra(ExtraData{
		OrderBy: []string{"LENGTH(title) DESC"},
	})

	sql, _ := qs.query.ToSQL()
	expected := "SELECT * FROM test_article ORDER BY views DESC, LENGTH(title) DESC"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

func TestExtraCombinedWithFilter(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&TestArticle{})

	qs := NewQuerySet[TestArticle]()
	qs = qs.Filter(Lookup{"Views__gt": 10}).Extra(ExtraData{
		Select: map[string]string{"is_popular": "views > 1000"},
		Where:  []string{"title != ?"},
		Params: []any{"spam"},
	}).Limit(5)

	sql, params := qs.query.ToSQL()
	expected := "SELECT *, (views > 1000) AS is_popular FROM test_article WHERE (views > ?) AND title != ? LIMIT 5"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
	if !reflect.DeepEqual(params, []any{10, "spam"}) {
		t.Errorf("Params mismatch: got %v", params)
	}
}

func TestExtraImmutability(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&TestArticle{})

	qs := NewQuerySet[TestArticle]()
	qs1 := qs.Extra(ExtraData{
		Where: []string{"views > 100"},
	})
	qs2 := qs.Extra(ExtraData{
		Where: []string{"views < 50"},
	})

	sql1, _ := qs1.query.ToSQL()
	sql2, _ := qs2.query.ToSQL()
	sql0, _ := qs.query.ToSQL()

	expected0 := "SELECT * FROM test_article"
	expected1 := "SELECT * FROM test_article WHERE views > 100"
	expected2 := "SELECT * FROM test_article WHERE views < 50"

	if sql0 != expected0 {
		t.Errorf("Original modified: expected %q, got %q", expected0, sql0)
	}
	if sql1 != expected1 {
		t.Errorf("qs1 wrong: expected %q, got %q", expected1, sql1)
	}
	if sql2 != expected2 {
		t.Errorf("qs2 wrong: expected %q, got %q", expected2, sql2)
	}
}

func TestExtraChainingAccumulates(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&TestArticle{})

	qs := NewQuerySet[TestArticle]()
	qs = qs.Extra(ExtraData{
		Where: []string{"a = 1"},
	}).Extra(ExtraData{
		Where: []string{"b = 2"},
	})

	sql, _ := qs.query.ToSQL()
	expected := "SELECT * FROM test_article WHERE a = 1 AND b = 2"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

func TestExtraSelectMultipleSorted(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&TestArticle{})

	qs := NewQuerySet[TestArticle]()
	qs = qs.Extra(ExtraData{
		Select: map[string]string{
			"zeta":  "1",
			"alpha": "2",
			"mid":   "3",
		},
	})

	sql, _ := qs.query.ToSQL()
	// Extra select entries are sorted by alias for deterministic output.
	expected := "SELECT *, (2) AS alpha, (3) AS mid, (1) AS zeta FROM test_article"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}
