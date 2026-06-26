package queryset

import (
	"github.com/pkshahid/JanGo/orm"
	"testing"
)

func TestRawQuerySet(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})
	setupTestDB(t, &DummyModel{})

	qs := NewQuerySet[DummyModel]()

	// Insert a test record
	qs.Create(&DummyModel{Name: "Alice", Age: 25})

	raw := qs.Raw("SELECT * FROM dummymodel WHERE name = ?", "Alice")

	if raw.SQL != "SELECT * FROM dummymodel WHERE name = ?" {
		t.Errorf("Expected raw SQL, got %s", raw.SQL)
	}
	if len(raw.Params) != 1 || raw.Params[0] != "Alice" {
		t.Errorf("Expected raw params, got %v", raw.Params)
	}

	// Execute raw query
	results, err := raw.All()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Name != "Alice" {
		t.Errorf("Expected Alice, got %s", results[0].Name)
	}

	// Get should return the single result
	obj, err := raw.Get()
	if err != nil {
		t.Errorf("Unexpected error from Get: %v", err)
	}
	if obj.Name != "Alice" {
		t.Errorf("Expected Alice, got %s", obj.Name)
	}

	// Get with no results should error
	rawEmpty := qs.Raw("SELECT * FROM dummymodel WHERE name = ?", "Nobody")
	_, err = rawEmpty.Get()
	if err == nil {
		t.Errorf("Expected error from Get with no results")
	}
}
