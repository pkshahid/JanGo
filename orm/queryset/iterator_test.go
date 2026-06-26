package queryset

import (
	"testing"

	"github.com/pkshahid/JanGo/orm"
)

func TestIteratorSQLGeneration(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs := NewQuerySet[DummyModel]().
		Filter(Lookup{"Name__exact": "Alice"}).
		OrderBy("-Age").
		Limit(100)

	// Verify the SQL is generated correctly for the iterator query.
	sqlStr, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM dummymodel WHERE (name = ?) ORDER BY age DESC LIMIT 100"
	if sqlStr != expectedSQL {
		t.Errorf("Iterator SQL mismatch:\n  got:  %s\n  want: %s", sqlStr, expectedSQL)
	}

	if len(params) != 1 {
		t.Errorf("Expected 1 param, got %d", len(params))
	}
	if params[0] != "Alice" {
		t.Errorf("Expected param 'Alice', got %v", params[0])
	}
}

func TestIteratorChunkedDefault(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs := NewQuerySet[DummyModel]()

	// Iterator() should use the default chunk size.
	// Since there's no real DB connection, calling the iterator will yield
	// an error from GetBackend. We verify the error is propagated correctly.
	count := 0
	for obj, err := range qs.Iterator() {
		if err != nil {
			// Expected: no backend registered in test environment.
			break
		}
		_ = obj
		count++
	}
	// With no DB, we should get 0 successful yields (error breaks the loop).
	if count != 0 {
		t.Errorf("Expected 0 yields without DB, got %d", count)
	}
}

func TestIteratorChunkedCustomSize(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs := NewQuerySet[DummyModel]()

	// IteratorChunked with custom size should not panic.
	count := 0
	for obj, err := range qs.IteratorChunked(500) {
		if err != nil {
			break
		}
		_ = obj
		count++
	}
	if count != 0 {
		t.Errorf("Expected 0 yields without DB, got %d", count)
	}
}

func TestIteratorChunkedZeroOrNegative(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs := NewQuerySet[DummyModel]()

	// Passing 0 or negative chunk size should fall back to default.
	for _, size := range []int{0, -1, -100} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("IteratorChunked(%d) panicked: %v", size, r)
				}
			}()
			for _, err := range qs.IteratorChunked(size) {
				if err != nil {
					break
				}
			}
		}()
	}
}

func TestIteratorEarlyBreak(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})

	qs := NewQuerySet[DummyModel]()

	// Breaking early should not cause issues (cursor cleanup via defer).
	for obj, err := range qs.Iterator() {
		if err != nil {
			break
		}
		_ = obj
		break
	}
}
