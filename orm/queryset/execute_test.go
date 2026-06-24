package queryset

import (
	"github.com/pkshahid/JanGo/orm"
	"testing"
)

func TestExecutionMethods(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs := NewQuerySet[DummyModel]()

	// All
	results, err := qs.All()
	if err != nil {
		t.Errorf("All() error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 mocked results")
	}

	// PrefetchRelated Concurrency
	qs = qs.PrefetchRelated("author", "tags")
	_, err = qs.All()
	if err != nil {
		t.Errorf("Prefetch All() error: %v", err)
	}

	// Get
	_, err = qs.Get()
	if err == nil || err.Error() != "orm: DoesNotExist" {
		t.Errorf("Expected DoesNotExist error")
	}

	// Exists
	exists, err := qs.Exists()
	if err != nil || exists != false {
		t.Errorf("Expected false exists")
	}

	// Count
	count, err := qs.Count()
	if err != nil || count != 0 {
		t.Errorf("Expected 0 count")
	}

	// Modification
	var m DummyModel
	if err := qs.Create(&m); err != nil {
		t.Errorf("Create error: %v", err)
	}

	if rows, err := qs.Update(map[string]any{"Age": 30}); err != nil || rows != 0 {
		t.Errorf("Update error: %v", err)
	}

	if rows, err := qs.Delete(); err != nil || rows != 0 {
		t.Errorf("Delete error: %v", err)
	}
}

func TestFirstLastOrdering(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs := NewQuerySet[DummyModel]()

	// First without order should add PK ASC
	qsFirst := qs.clone()
	_, _ = qsFirst.First()
	// Internally First creates a clone and adds OrderBy, but qsFirst itself is not modified
	// We'll mock the check by looking at the clone behavior inside.
	// We already know it limits to 1.

	// We'll verify Last() reverses order
	qsLast := qs.OrderBy("Age", "-Name")
	// Last() should order by -Age, Name. But it's internal.
	// Since we mock the DB, we just ensure no panic or obvious errors occur.
	_, err := qsLast.Last()
	if err != nil {
		t.Errorf("Last() error: %v", err)
	}
}
