package queryset

import (
	"testing"
	"github.com/pkshahid/JanGo/orm"
)

func TestRawQuerySet(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs := NewQuerySet[DummyModel]()

	raw := qs.Raw("SELECT * FROM test WHERE name = ?", "Alice").Using("replica")

	if raw.SQL != "SELECT * FROM test WHERE name = ?" {
		t.Errorf("Expected raw SQL, got %s", raw.SQL)
	}
	if len(raw.Params) != 1 || raw.Params[0] != "Alice" {
		t.Errorf("Expected raw params, got %v", raw.Params)
	}
	if raw.Database != "replica" {
		t.Errorf("Expected database replica, got %s", raw.Database)
	}

	// Execution stubs
	results, err := raw.All()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results")
	}

	_, err = raw.Get()
	if err == nil {
		t.Errorf("Expected error from Get with no results")
	}
}
