package queryset

import (
	"testing"
	"github.com/pkshahid/JanGo/orm"
)

func TestAggregate(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs := NewQuerySet[DummyModel]().Filter(Lookup{"Age__gt": 18})

	res, err := qs.Aggregate(Sum("age"), Count("id"))
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}

	expectedSQL := "SELECT SUM(age), COUNT(id) FROM dummymodel WHERE (age > ?)"
	if res["_sql"] != expectedSQL {
		t.Errorf("Expected SQL %q, got %q", expectedSQL, res["_sql"])
	}
}

func TestAnnotate(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs := NewQuerySet[DummyModel]().Annotate(Count("id"))

	if len(qs.query.Annotations) != 1 {
		t.Errorf("Expected 1 annotation")
	}

	if qs.query.Annotations[0].ToSQL() != "COUNT(id)" {
		t.Errorf("Annotation SQL mismatch")
	}
}
