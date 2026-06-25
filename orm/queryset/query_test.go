package queryset

import (
	"github.com/pkshahid/JanGo/orm"
	"reflect"
	"testing"
)

type TestArticle struct {
	orm.Model
	Title string `gd:"CharField"`
	Views int    `gd:"IntegerField"`
}

func (a *TestArticle) ModelMeta() *orm.Meta {
	return &orm.Meta{DbTable: "test_article"}
}

func TestQuerySQL(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&TestArticle{})

	info, _ := orm.GetModelInfo(&TestArticle{})
	query := NewQuery(info)

	// Filter
	query.Where = Q(Lookup{"Title__icontains": "Go"})
	sql, params := query.ToSQL()

	expectedSQL := "SELECT * FROM test_article WHERE (LOWER(title) LIKE LOWER(?))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if !reflect.DeepEqual(params, []any{"%Go%"}) {
		t.Errorf("Params mismatch")
	}

	// Exclude + Limit + OrderBy
	query.Exclude = Q(Lookup{"Views__lt": 10})
	query.Limit = 5
	query.OrderBy = []string{"-Views", "Title"}

	sql, params = query.ToSQL()
	expectedSQL2 := "SELECT * FROM test_article WHERE (LOWER(title) LIKE LOWER(?)) AND NOT (views < ?) ORDER BY views DESC, title ASC LIMIT 5"
	if sql != expectedSQL2 {
		t.Errorf("Expected %q, got %q", expectedSQL2, sql)
	}
	if len(params) != 2 {
		t.Errorf("Expected 2 params")
	}

	// Q Chaining
	query2 := NewQuery(info)
	qNode := Q(Lookup{"Title__exact": "A"}).Or(Q(Lookup{"Views__gt": 100}))
	query2.Where = qNode

	sql2, _ := query2.ToSQL()
	// Depending on map iteration order, clauses inside Q(Lookup{}) might vary, but since our Lookups only have 1 key here, it's deterministic.
	// Wait, the Q constructor creates a node. `Or` wraps them.
	// `Or` wraps two nodes.
	expectedSQL3 := "SELECT * FROM test_article WHERE ((title = ?) OR (views > ?))"
	if sql2 != expectedSQL3 {
		t.Errorf("Expected %q, got %q", expectedSQL3, sql2)
	}
}
