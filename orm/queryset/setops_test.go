package queryset

import (
	"reflect"
	"testing"

	"github.com/pkshahid/JanGo/orm"
)

func TestUnion(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs1 := NewQuerySet[DummyModel]().Filter(Lookup{"Name__exact": "Alice"})
	qs2 := NewQuerySet[DummyModel]().Filter(Lookup{"Name__exact": "Bob"})

	combined := qs1.Union(qs2)

	sql, params := combined.query.ToSQL()
	expectedSQL := "(SELECT * FROM dummymodel WHERE (name = ?)) UNION (SELECT * FROM dummymodel WHERE (name = ?))"
	if sql != expectedSQL {
		t.Errorf("Union SQL mismatch:\n  expected: %s\n  got:      %s", expectedSQL, sql)
	}
	if !reflect.DeepEqual(params, []any{"Alice", "Bob"}) {
		t.Errorf("Union params mismatch: got %v", params)
	}
}

func TestUnionAll(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs1 := NewQuerySet[DummyModel]().Filter(Lookup{"Name__exact": "Alice"})
	qs2 := NewQuerySet[DummyModel]().Filter(Lookup{"Name__exact": "Bob"})

	combined := qs1.UnionAll(qs2)

	sql, params := combined.query.ToSQL()
	expectedSQL := "(SELECT * FROM dummymodel WHERE (name = ?)) UNION ALL (SELECT * FROM dummymodel WHERE (name = ?))"
	if sql != expectedSQL {
		t.Errorf("UnionAll SQL mismatch:\n  expected: %s\n  got:      %s", expectedSQL, sql)
	}
	if !reflect.DeepEqual(params, []any{"Alice", "Bob"}) {
		t.Errorf("UnionAll params mismatch: got %v", params)
	}
}

func TestUnionMultiple(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs1 := NewQuerySet[DummyModel]().Filter(Lookup{"Name__exact": "Alice"})
	qs2 := NewQuerySet[DummyModel]().Filter(Lookup{"Age__gt": 20})
	qs3 := NewQuerySet[DummyModel]().Filter(Lookup{"Age__lt": 10})

	combined := qs1.Union(qs2, qs3)

	sql, params := combined.query.ToSQL()
	expectedSQL := "(SELECT * FROM dummymodel WHERE (name = ?)) UNION (SELECT * FROM dummymodel WHERE (age > ?)) UNION (SELECT * FROM dummymodel WHERE (age < ?))"
	if sql != expectedSQL {
		t.Errorf("Union multiple SQL mismatch:\n  expected: %s\n  got:      %s", expectedSQL, sql)
	}
	if !reflect.DeepEqual(params, []any{"Alice", 20, 10}) {
		t.Errorf("Union multiple params mismatch: got %v", params)
	}
}

func TestUnionWithOrderByAndLimit(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs1 := NewQuerySet[DummyModel]().Filter(Lookup{"Name__exact": "Alice"})
	qs2 := NewQuerySet[DummyModel]().Filter(Lookup{"Name__exact": "Bob"})

	combined := qs1.Union(qs2).OrderBy("-Name").Limit(10)

	sql, _ := combined.query.ToSQL()
	expectedSQL := "(SELECT * FROM dummymodel WHERE (name = ?)) UNION (SELECT * FROM dummymodel WHERE (name = ?)) ORDER BY name DESC LIMIT 10"
	if sql != expectedSQL {
		t.Errorf("Union with ORDER BY/LIMIT SQL mismatch:\n  expected: %s\n  got:      %s", expectedSQL, sql)
	}
}

func TestIntersection(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs1 := NewQuerySet[DummyModel]().Filter(Lookup{"Name__exact": "Alice"})
	qs2 := NewQuerySet[DummyModel]().Filter(Lookup{"Age__gt": 20})

	combined := qs1.Intersection(qs2)

	sql, params := combined.query.ToSQL()
	expectedSQL := "(SELECT * FROM dummymodel WHERE (name = ?)) INTERSECT (SELECT * FROM dummymodel WHERE (age > ?))"
	if sql != expectedSQL {
		t.Errorf("Intersection SQL mismatch:\n  expected: %s\n  got:      %s", expectedSQL, sql)
	}
	if !reflect.DeepEqual(params, []any{"Alice", 20}) {
		t.Errorf("Intersection params mismatch: got %v", params)
	}
}

func TestIntersectionMultiple(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs1 := NewQuerySet[DummyModel]()
	qs2 := NewQuerySet[DummyModel]().Filter(Lookup{"Age__gt": 20})
	qs3 := NewQuerySet[DummyModel]().Filter(Lookup{"Age__lt": 50})

	combined := qs1.Intersection(qs2, qs3)

	sql, _ := combined.query.ToSQL()
	expectedSQL := "(SELECT * FROM dummymodel) INTERSECT (SELECT * FROM dummymodel WHERE (age > ?)) INTERSECT (SELECT * FROM dummymodel WHERE (age < ?))"
	if sql != expectedSQL {
		t.Errorf("Intersection multiple SQL mismatch:\n  expected: %s\n  got:      %s", expectedSQL, sql)
	}
}

func TestDifference(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs1 := NewQuerySet[DummyModel]().Filter(Lookup{"Name__exact": "Alice"})
	qs2 := NewQuerySet[DummyModel]().Filter(Lookup{"Age__gt": 20})

	combined := qs1.Difference(qs2)

	sql, params := combined.query.ToSQL()
	expectedSQL := "(SELECT * FROM dummymodel WHERE (name = ?)) EXCEPT (SELECT * FROM dummymodel WHERE (age > ?))"
	if sql != expectedSQL {
		t.Errorf("Difference SQL mismatch:\n  expected: %s\n  got:      %s", expectedSQL, sql)
	}
	if !reflect.DeepEqual(params, []any{"Alice", 20}) {
		t.Errorf("Difference params mismatch: got %v", params)
	}
}

func TestDifferenceMultiple(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs1 := NewQuerySet[DummyModel]()
	qs2 := NewQuerySet[DummyModel]().Filter(Lookup{"Age__gt": 20})
	qs3 := NewQuerySet[DummyModel]().Filter(Lookup{"Age__lt": 10})

	combined := qs1.Difference(qs2, qs3)

	sql, _ := combined.query.ToSQL()
	expectedSQL := "(SELECT * FROM dummymodel) EXCEPT (SELECT * FROM dummymodel WHERE (age > ?)) EXCEPT (SELECT * FROM dummymodel WHERE (age < ?))"
	if sql != expectedSQL {
		t.Errorf("Difference multiple SQL mismatch:\n  expected: %s\n  got:      %s", expectedSQL, sql)
	}
}

func TestSetOpsImmutability(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs1 := NewQuerySet[DummyModel]().Filter(Lookup{"Name__exact": "Alice"})
	qs2 := NewQuerySet[DummyModel]().Filter(Lookup{"Name__exact": "Bob"})

	combined := qs1.Union(qs2)

	// Original qs1 should be unmodified
	sql1, _ := qs1.query.ToSQL()
	expectedSQL1 := "SELECT * FROM dummymodel WHERE (name = ?)"
	if sql1 != expectedSQL1 {
		t.Errorf("qs1 was modified by Union:\n  expected: %s\n  got:      %s", expectedSQL1, sql1)
	}

	// Original qs2 should be unmodified
	sql2, _ := qs2.query.ToSQL()
	expectedSQL2 := "SELECT * FROM dummymodel WHERE (name = ?)"
	if sql2 != expectedSQL2 {
		t.Errorf("qs2 was modified by Union:\n  expected: %s\n  got:      %s", expectedSQL2, sql2)
	}

	// Combined should have the set op
	if combined.query.SetOp != SetOpUnion {
		t.Errorf("Combined query should have SetOpUnion, got %s", combined.query.SetOp)
	}
}

func TestSetOpsWithOnlyFields(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs1 := NewQuerySet[DummyModel]().Only("Name").Filter(Lookup{"Name__exact": "Alice"})
	qs2 := NewQuerySet[DummyModel]().Only("Name").Filter(Lookup{"Name__exact": "Bob"})

	combined := qs1.Union(qs2)

	sql, _ := combined.query.ToSQL()
	expectedSQL := "(SELECT Name FROM dummymodel WHERE (name = ?)) UNION (SELECT Name FROM dummymodel WHERE (name = ?))"
	if sql != expectedSQL {
		t.Errorf("Union with OnlyFields SQL mismatch:\n  expected: %s\n  got:      %s", expectedSQL, sql)
	}
}

func TestSetOpsWithOffset(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs1 := NewQuerySet[DummyModel]()
	qs2 := NewQuerySet[DummyModel]().Filter(Lookup{"Name__exact": "Bob"})

	combined := qs1.Union(qs2).OrderBy("Name").Limit(5).Offset(10)

	sql, _ := combined.query.ToSQL()
	expectedSQL := "(SELECT * FROM dummymodel) UNION (SELECT * FROM dummymodel WHERE (name = ?)) ORDER BY name ASC LIMIT 5 OFFSET 10"
	if sql != expectedSQL {
		t.Errorf("Union with offset SQL mismatch:\n  expected: %s\n  got:      %s", expectedSQL, sql)
	}
}
