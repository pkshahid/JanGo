package queryset

import (
	"testing"
	"github.com/godjango/godjango/orm"
)

type DummyModel struct {
	orm.Model
	Name string `gd:"CharField"`
	Age  int    `gd:"IntegerField"`
}

func TestQuerySetImmutability(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs := NewQuerySet[DummyModel]()

	// Chain 1
	qs1 := qs.Filter(Lookup{"Name__exact": "Alice"})

	// Chain 2
	qs2 := qs1.Filter(Lookup{"Age__gt": 20}).Limit(5)

	// Chain 3 (diverging)
	qs3 := qs1.Exclude(Lookup{"Age__lt": 10}).OrderBy("-Name")

	// Check qs1
	sql1, _ := qs1.query.ToSQL()
	if sql1 != "SELECT * FROM dummymodel WHERE (name = ?)" {
		t.Errorf("qs1 modified improperly: %s", sql1)
	}

	// Check qs2
	sql2, _ := qs2.query.ToSQL()
	if sql2 != "SELECT * FROM dummymodel WHERE ((name = ?) AND (age > ?)) LIMIT 5" {
		t.Errorf("qs2 built improperly: %s", sql2)
	}

	// Check qs3
	sql3, _ := qs3.query.ToSQL()
	if sql3 != "SELECT * FROM dummymodel WHERE (name = ?) AND NOT (age < ?) ORDER BY name DESC" {
		t.Errorf("qs3 built improperly: %s", sql3)
	}
}
