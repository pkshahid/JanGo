package queryset

import (
	"testing"

	"github.com/pkshahid/JanGo/orm"
)

func TestExecutionMethods(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})
	setupTestDB(t, &DummyModel{})

	qs := NewQuerySet[DummyModel]()

	// All on empty table
	results, err := qs.All()
	if err != nil {
		t.Errorf("All() error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results on empty table, got %d", len(results))
	}

	// Create a record
	obj := &DummyModel{Name: "Alice", Age: 25}
	if err := qs.Create(obj); err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if obj.ID == 0 {
		t.Error("Expected non-zero ID after Create")
	}

	// Create another
	obj2 := &DummyModel{Name: "Bob", Age: 30}
	if err := qs.Create(obj2); err != nil {
		t.Fatalf("Create error: %v", err)
	}

	// All should return 2 records
	results, err = qs.All()
	if err != nil {
		t.Errorf("All() error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// PrefetchRelated should not error
	qsPrefetch := qs.PrefetchRelated("author", "tags")
	_, err = qsPrefetch.All()
	if err != nil {
		t.Errorf("Prefetch All() error: %v", err)
	}

	// Get with filter
	alice, err := qs.Filter(Lookup{"Name__exact": "Alice"}).Get()
	if err != nil {
		t.Errorf("Get error: %v", err)
	}
	if alice.Name != "Alice" {
		t.Errorf("Expected Alice, got %s", alice.Name)
	}

	// Get with no match
	_, err = qs.Filter(Lookup{"Name__exact": "Nobody"}).Get()
	if err == nil || err.Error() != "orm: DoesNotExist" {
		t.Errorf("Expected DoesNotExist error, got %v", err)
	}

	// Exists
	exists, err := qs.Filter(Lookup{"Name__exact": "Alice"}).Exists()
	if err != nil {
		t.Errorf("Exists error: %v", err)
	}
	if !exists {
		t.Error("Expected exists=true for Alice")
	}

	exists, err = qs.Filter(Lookup{"Name__exact": "Nobody"}).Exists()
	if err != nil {
		t.Errorf("Exists error: %v", err)
	}
	if exists {
		t.Error("Expected exists=false for Nobody")
	}

	// Count
	count, err := qs.Count()
	if err != nil {
		t.Errorf("Count error: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count=2, got %d", count)
	}

	// Update
	rows, err := qs.Filter(Lookup{"Name__exact": "Alice"}).Update(map[string]any{"Age": 26})
	if err != nil {
		t.Errorf("Update error: %v", err)
	}
	if rows != 1 {
		t.Errorf("Expected 1 row updated, got %d", rows)
	}

	// Verify update
	alice, err = qs.Filter(Lookup{"Name__exact": "Alice"}).Get()
	if err != nil {
		t.Fatalf("Get after update error: %v", err)
	}
	if alice.Age != 26 {
		t.Errorf("Expected Age=26 after update, got %d", alice.Age)
	}

	// Delete
	rows, err = qs.Filter(Lookup{"Name__exact": "Bob"}).Delete()
	if err != nil {
		t.Errorf("Delete error: %v", err)
	}
	if rows != 1 {
		t.Errorf("Expected 1 row deleted, got %d", rows)
	}

	// Verify deletion
	count, err = qs.Count()
	if err != nil {
		t.Errorf("Count after delete error: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count=1 after delete, got %d", count)
	}
}

func TestFirstLastOrdering(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})
	setupTestDB(t, &DummyModel{})

	qs := NewQuerySet[DummyModel]()

	// Create records
	qs.Create(&DummyModel{Name: "Charlie", Age: 20})
	qs.Create(&DummyModel{Name: "Alice", Age: 30})
	qs.Create(&DummyModel{Name: "Bob", Age: 25})

	// First without order should order by PK ASC
	first, err := qs.First()
	if err != nil {
		t.Fatalf("First() error: %v", err)
	}
	if first == nil {
		t.Fatal("Expected non-nil first result")
	}
	if first.Name != "Charlie" {
		t.Errorf("Expected Charlie (first PK), got %s", first.Name)
	}

	// First with ordering by Age
	first, err = qs.OrderBy("Age").First()
	if err != nil {
		t.Fatalf("First() error: %v", err)
	}
	if first.Age != 20 {
		t.Errorf("Expected Age=20 (youngest), got %d", first.Age)
	}

	// Last with ordering by Age should return oldest
	last, err := qs.OrderBy("Age").Last()
	if err != nil {
		t.Fatalf("Last() error: %v", err)
	}
	if last == nil {
		t.Fatal("Expected non-nil last result")
	}
	if last.Age != 30 {
		t.Errorf("Expected Age=30 (oldest), got %d", last.Age)
	}

	// Last without ordering should use -PK
	last, err = qs.Last()
	if err != nil {
		t.Fatalf("Last() error: %v", err)
	}
	if last.Name != "Bob" {
		t.Errorf("Expected Bob (last PK), got %s", last.Name)
	}
}

func TestValuesAndValuesList(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})
	setupTestDB(t, &DummyModel{})

	qs := NewQuerySet[DummyModel]()
	qs.Create(&DummyModel{Name: "Alice", Age: 25})
	qs.Create(&DummyModel{Name: "Bob", Age: 30})

	// Values
	results, err := qs.Values("name", "age")
	if err != nil {
		t.Fatalf("Values error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 value rows, got %d", len(results))
	}

	// ValuesList
	listResults, err := qs.ValuesList("name")
	if err != nil {
		t.Fatalf("ValuesList error: %v", err)
	}
	if len(listResults) != 2 {
		t.Fatalf("Expected 2 valueslist rows, got %d", len(listResults))
	}
}

func TestGetOrCreate(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})
	setupTestDB(t, &DummyModel{})

	qs := NewQuerySet[DummyModel]()

	// Should create
	obj, created, err := qs.GetOrCreate(Lookup{"Name": "Carol"}, map[string]any{"Age": 28})
	if err != nil {
		t.Fatalf("GetOrCreate error: %v", err)
	}
	if !created {
		t.Error("Expected created=true")
	}
	if obj.Name != "Carol" {
		t.Errorf("Expected Name=Carol, got %s", obj.Name)
	}
	if obj.Age != 28 {
		t.Errorf("Expected Age=28, got %d", obj.Age)
	}

	// Should find existing
	obj, created, err = qs.GetOrCreate(Lookup{"Name": "Carol"}, map[string]any{"Age": 99})
	if err != nil {
		t.Fatalf("GetOrCreate error: %v", err)
	}
	if created {
		t.Error("Expected created=false for existing record")
	}
	if obj.Age != 28 {
		t.Errorf("Expected Age=28 (not updated), got %d", obj.Age)
	}
}

func TestUpdateOrCreate(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})
	setupTestDB(t, &DummyModel{})

	qs := NewQuerySet[DummyModel]()

	// Should create
	obj, created, err := qs.UpdateOrCreate(Lookup{"Name": "Dave"}, map[string]any{"Age": 40})
	if err != nil {
		t.Fatalf("UpdateOrCreate error: %v", err)
	}
	if !created {
		t.Error("Expected created=true")
	}
	if obj.Age != 40 {
		t.Errorf("Expected Age=40, got %d", obj.Age)
	}

	// Should update existing
	obj, created, err = qs.UpdateOrCreate(Lookup{"Name": "Dave"}, map[string]any{"Age": 41})
	if err != nil {
		t.Fatalf("UpdateOrCreate error: %v", err)
	}
	if created {
		t.Error("Expected created=false for existing record")
	}
	if obj.Age != 41 {
		t.Errorf("Expected Age=41 (updated), got %d", obj.Age)
	}
}

func TestBulkCreate(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})
	setupTestDB(t, &DummyModel{})

	qs := NewQuerySet[DummyModel]()

	objs := []DummyModel{
		{Name: "Bulk1", Age: 10},
		{Name: "Bulk2", Age: 20},
		{Name: "Bulk3", Age: 30},
	}
	if err := qs.BulkCreate(objs); err != nil {
		t.Fatalf("BulkCreate error: %v", err)
	}

	count, err := qs.Count()
	if err != nil {
		t.Fatalf("Count error: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected count=3, got %d", count)
	}
}

func TestBulkUpdate(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&DummyModel{})
	setupTestDB(t, &DummyModel{})

	qs := NewQuerySet[DummyModel]()

	// Create records
	obj1 := &DummyModel{Name: "Original1", Age: 10}
	obj2 := &DummyModel{Name: "Original2", Age: 20}
	qs.Create(obj1)
	qs.Create(obj2)

	// Modify and bulk update
	obj1.Age = 100
	obj2.Age = 200
	if err := qs.BulkUpdate([]DummyModel{*obj1, *obj2}, []string{"Age"}); err != nil {
		t.Fatalf("BulkUpdate error: %v", err)
	}

	// Verify
	updated1, _ := qs.Filter(Lookup{"Name__exact": "Original1"}).Get()
	if updated1.Age != 100 {
		t.Errorf("Expected Age=100, got %d", updated1.Age)
	}
	updated2, _ := qs.Filter(Lookup{"Name__exact": "Original2"}).Get()
	if updated2.Age != 200 {
		t.Errorf("Expected Age=200, got %d", updated2.Age)
	}
}
