package queryset

import (
	"testing"

	"github.com/pkshahid/JanGo/orm"
)

type ManagerTestModel struct {
	orm.Model
	Title     string `gd:"CharField,max_length=200"`
	Published bool   `gd:"BooleanField,default=false"`
	Views     int    `gd:"IntegerField,default=0"`
}

// PublishedManager is a custom manager that pre-filters to published records.
type PublishedManager struct {
	Manager[ManagerTestModel]
}

func NewPublishedManager() PublishedManager {
	return PublishedManager{
		Manager: NewManager[ManagerTestModel]().WithQuerySet(func() QuerySet[ManagerTestModel] {
			return NewQuerySet[ManagerTestModel]().Filter(Lookup{"Published": true})
		}),
	}
}

// Published returns the pre-filtered queryset (custom method).
func (m PublishedManager) Published() QuerySet[ManagerTestModel] {
	return m.GetQuerySet()
}

// Popular returns records with more than 100 views (custom method).
func (m PublishedManager) Popular() QuerySet[ManagerTestModel] {
	return m.GetQuerySet().Filter(Lookup{"Views__gt": 100})
}

func setupManagerTest(t *testing.T) {
	t.Helper()
	orm.ClearRegistry()
	if err := orm.Register(&ManagerTestModel{}); err != nil {
		t.Fatalf("failed to register model: %v", err)
	}
	setupTestDB(t, &ManagerTestModel{})
}

func TestNewManager(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	qs := m.GetQuerySet()
	sql, _ := qs.query.ToSQL()
	expected := "SELECT * FROM managertestmodel"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestManagerImplementsManagerInterface(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()
	var _ ManagerInterface = m
}

func TestManagerDelegationFilter(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	qs := m.Filter(Lookup{"Title": "Hello"})
	sql, _ := qs.query.ToSQL()
	expected := "SELECT * FROM managertestmodel WHERE (title = ?)"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestManagerDelegationChained(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	qs := m.Filter(Lookup{"Title__contains": "Go"}).OrderBy("-Views").Limit(10)
	sql, _ := qs.query.ToSQL()
	expected := "SELECT * FROM managertestmodel WHERE (title LIKE ?) ORDER BY views DESC LIMIT 10"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestManagerDelegationExclude(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	qs := m.Exclude(Lookup{"Views__lt": 10})
	sql, _ := qs.query.ToSQL()
	expected := "SELECT * FROM managertestmodel WHERE NOT (views < ?)"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestManagerDelegationOrderBy(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	qs := m.OrderBy("Title")
	sql, _ := qs.query.ToSQL()
	expected := "SELECT * FROM managertestmodel ORDER BY title ASC"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestManagerDelegationUsing(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	qs := m.Using("replica")
	if qs.query.Database != "replica" {
		t.Errorf("expected database 'replica', got %q", qs.query.Database)
	}
}

func TestManagerDelegationAnnotate(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	qs := m.Annotate(NewF("Views"))
	if len(qs.query.Annotations) != 1 {
		t.Errorf("expected 1 annotation, got %d", len(qs.query.Annotations))
	}
}

func TestManagerDelegationExtra(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	qs := m.Extra(ExtraData{
		Where:  []string{"views > 0"},
		Params: []any{},
	})
	sql, _ := qs.query.ToSQL()
	expected := "SELECT * FROM managertestmodel WHERE views > 0"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestManagerDelegationRaw(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	rqs := m.Raw("SELECT * FROM managertestmodel WHERE title = ?", "test")
	if rqs.SQL != "SELECT * FROM managertestmodel WHERE title = ?" {
		t.Errorf("unexpected raw SQL: %q", rqs.SQL)
	}
	if len(rqs.Params) != 1 || rqs.Params[0] != "test" {
		t.Errorf("unexpected raw params: %v", rqs.Params)
	}
}

func TestManagerDelegationAll(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	results, err := m.All()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results on empty table, got %d", len(results))
	}
}

func TestManagerDelegationCount(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	count, err := m.Count()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 count on empty table, got %d", count)
	}
}

func TestManagerDelegationExists(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	exists, err := m.Exists()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false on empty table")
	}
}

func TestCustomManagerWithQuerySet(t *testing.T) {
	setupManagerTest(t)
	m := NewPublishedManager()

	// GetQuerySet should return the pre-filtered queryset.
	qs := m.GetQuerySet()
	sql, _ := qs.query.ToSQL()
	expected := "SELECT * FROM managertestmodel WHERE (published = ?)"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestCustomManagerPublishedMethod(t *testing.T) {
	setupManagerTest(t)
	m := NewPublishedManager()

	qs := m.Published()
	sql, _ := qs.query.ToSQL()
	expected := "SELECT * FROM managertestmodel WHERE (published = ?)"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestCustomManagerPopularMethod(t *testing.T) {
	setupManagerTest(t)
	m := NewPublishedManager()

	qs := m.Popular()
	sql, _ := qs.query.ToSQL()
	expected := "SELECT * FROM managertestmodel WHERE ((published = ?) AND (views > ?))"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestCustomManagerDelegationUsesCustomQuerySet(t *testing.T) {
	setupManagerTest(t)
	m := NewPublishedManager()

	// Delegated methods should use the custom queryset from WithQuerySet.
	qs := m.Filter(Lookup{"Title": "Test"})
	sql, _ := qs.query.ToSQL()
	expected := "SELECT * FROM managertestmodel WHERE ((published = ?) AND (title = ?))"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestCustomManagerChainedFromCustomQuerySet(t *testing.T) {
	setupManagerTest(t)
	m := NewPublishedManager()

	qs := m.OrderBy("-Views").Limit(5)
	sql, _ := qs.query.ToSQL()
	expected := "SELECT * FROM managertestmodel WHERE (published = ?) ORDER BY views DESC LIMIT 5"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestWithQuerySetDoesNotMutateOriginal(t *testing.T) {
	setupManagerTest(t)
	original := NewManager[ManagerTestModel]()

	custom := original.WithQuerySet(func() QuerySet[ManagerTestModel] {
		return NewQuerySet[ManagerTestModel]().Filter(Lookup{"Published": true})
	})

	// Original should have no filter.
	origSQL, _ := original.GetQuerySet().query.ToSQL()
	if origSQL != "SELECT * FROM managertestmodel" {
		t.Errorf("original manager was mutated: %q", origSQL)
	}

	// Custom should have the filter.
	customSQL, _ := custom.GetQuerySet().query.ToSQL()
	expected := "SELECT * FROM managertestmodel WHERE (published = ?)"
	if customSQL != expected {
		t.Errorf("custom manager missing filter: %q", customSQL)
	}
}

func TestManagerGetOrCreate(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	obj, created, err := m.GetOrCreate(Lookup{"Title": "Test"}, map[string]any{"Published": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = obj
	_ = created
}

func TestManagerUpdateOrCreate(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	obj, created, err := m.UpdateOrCreate(Lookup{"Title": "Test"}, map[string]any{"Views": 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = obj
	_ = created
}

func TestManagerCreate(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	obj := &ManagerTestModel{Title: "New Post", Published: true}
	if err := m.Create(obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagerBulkCreate(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	objs := []ManagerTestModel{
		{Title: "Post 1", Published: true},
		{Title: "Post 2", Published: false},
	}
	if err := m.BulkCreate(objs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagerBulkUpdate(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	objs := []ManagerTestModel{
		{Title: "Post 1", Published: true},
		{Title: "Post 2", Published: false},
	}
	if err := m.BulkUpdate(objs, []string{"Title"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagerUpdate(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	n, err := m.Filter(Lookup{"Published": false}).Update(map[string]any{"Published": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = n
}

func TestManagerDelete(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	n, err := m.Filter(Lookup{"Published": false}).Delete()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = n
}

func TestManagerValues(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	results, err := m.Values("title")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results == nil {
		t.Error("expected non-nil results")
	}
}

func TestManagerValuesList(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	results, err := m.ValuesList("title")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results == nil {
		t.Error("expected non-nil results")
	}
}

func TestManagerAggregate(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	result, err := m.Aggregate(NewF("Views"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestManagerDates(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	results, err := m.Dates("CreatedAt", DateKindYear)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results == nil {
		t.Error("expected non-nil results")
	}
}

func TestManagerDatetimes(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	results, err := m.Datetimes("CreatedAt", DateTimeKindDay)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results == nil {
		t.Error("expected non-nil results")
	}
}

func TestManagerUnion(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	other := NewQuerySet[ManagerTestModel]().Filter(Lookup{"Published": true})
	qs := m.Union(other)
	if qs.query.SetOp != SetOpUnion {
		t.Errorf("expected SetOpUnion, got %q", qs.query.SetOp)
	}
}

func TestManagerIntersection(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	other := NewQuerySet[ManagerTestModel]().Filter(Lookup{"Published": true})
	qs := m.Intersection(other)
	if qs.query.SetOp != SetOpIntersection {
		t.Errorf("expected SetOpIntersection, got %q", qs.query.SetOp)
	}
}

func TestManagerDifference(t *testing.T) {
	setupManagerTest(t)
	m := NewManager[ManagerTestModel]()

	other := NewQuerySet[ManagerTestModel]().Filter(Lookup{"Published": true})
	qs := m.Difference(other)
	if qs.query.SetOp != SetOpDifference {
		t.Errorf("expected SetOpDifference, got %q", qs.query.SetOp)
	}
}
