package queryset

import (
	"reflect"
	"testing"

	"github.com/pkshahid/JanGo/orm"
)

// --- Test models ---

type JoinUser struct {
	orm.Model
	Username string `gd:"CharField,max_length=150"`
	Email    string `gd:"CharField,max_length=200"`
}

type JoinProfile struct {
	orm.Model
	User   *JoinUser   `gd:"ForeignKey,to=auth.JoinUser,on_delete=CASCADE,db_column=user_id"`
	Bio    string      `gd:"TextField"`
	Avatar string      `gd:"CharField,max_length=200"`
}

type JoinArticle struct {
	orm.Model
	Title   string       `gd:"CharField,max_length=200"`
	Author  *JoinUser    `gd:"ForeignKey,to=auth.JoinUser,on_delete=CASCADE,db_column=author_id"`
	Profile *JoinProfile `gd:"ForeignKey,to=auth.JoinProfile,on_delete=SET_NULL,db_column=profile_id"`
}

func setupJoinModels(t *testing.T) {
	t.Helper()
	orm.ClearRegistry()
	if err := orm.Register(&JoinUser{}); err != nil {
		t.Fatalf("failed to register JoinUser: %v", err)
	}
	if err := orm.Register(&JoinProfile{}); err != nil {
		t.Fatalf("failed to register JoinProfile: %v", err)
	}
	if err := orm.Register(&JoinArticle{}); err != nil {
		t.Fatalf("failed to register JoinArticle: %v", err)
	}
}

// --- SelectRelated tests ---

func TestSelectRelatedSingleJoin(t *testing.T) {
	setupJoinModels(t)

	qs := NewQuerySet[JoinArticle]().SelectRelated(1)
	sql, _ := qs.query.ToSQL()

	// Should contain LEFT JOIN for Author and Profile (both are FK fields at depth 1).
	if !contains(sql, "LEFT JOIN joinuser AS T1 ON T0.author_id = T1.id") {
		t.Errorf("Expected LEFT JOIN for Author, got: %s", sql)
	}
	if !contains(sql, "LEFT JOIN joinprofile AS T2 ON T0.profile_id = T2.id") {
		t.Errorf("Expected LEFT JOIN for Profile, got: %s", sql)
	}

	// Should select base table columns and related columns.
	if !contains(sql, "T0.*") {
		t.Errorf("Expected T0.* in SELECT, got: %s", sql)
	}
	// Related columns should be aliased with path__column format.
	if !contains(sql, "T1.id AS Author__id") {
		t.Errorf("Expected T1.id AS Author__id, got: %s", sql)
	}
	if !contains(sql, "T1.username AS Author__username") {
		t.Errorf("Expected T1.username AS Author__username, got: %s", sql)
	}
}

func TestSelectRelatedNestedJoin(t *testing.T) {
	setupJoinModels(t)

	qs := NewQuerySet[JoinArticle]().SelectRelated(2)
	sql, _ := qs.query.ToSQL()

	// Depth 2 should also join Profile → User.
	if !contains(sql, "LEFT JOIN joinuser AS T3 ON T2.user_id = T3.id") {
		t.Errorf("Expected nested JOIN for Profile.User at depth 2, got: %s", sql)
	}
}

func TestSelectRelatedZeroNoJoins(t *testing.T) {
	setupJoinModels(t)

	qs := NewQuerySet[JoinArticle]().SelectRelated(0)
	sql, _ := qs.query.ToSQL()

	// No joins should be generated.
	if contains(sql, "LEFT JOIN") {
		t.Errorf("Expected no JOINs with SelectRelated(0), got: %s", sql)
	}
}

// --- IN lookup expansion tests ---

func TestInLookupWithSlice(t *testing.T) {
	setupJoinModels(t)

	qs := NewQuerySet[JoinArticle]().Filter(Lookup{"Title__in": []string{"Go", "Rust", "Python"}})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM joinarticle WHERE (title IN (?, ?, ?))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if !reflect.DeepEqual(params, []any{"Go", "Rust", "Python"}) {
		t.Errorf("Params mismatch: got %v", params)
	}
}

func TestInLookupWithIntSlice(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&TestArticle{})

	qs := NewQuerySet[TestArticle]().Filter(Lookup{"Views__in": []int{1, 2, 3}})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM test_article WHERE (views IN (?, ?, ?))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if !reflect.DeepEqual(params, []any{1, 2, 3}) {
		t.Errorf("Params mismatch: got %v", params)
	}
}

func TestInLookupWithEmptySlice(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&TestArticle{})

	qs := NewQuerySet[TestArticle]().Filter(Lookup{"Views__in": []int{}})
	sql, params := qs.query.ToSQL()

	// Empty IN should produce a always-false condition.
	if !contains(sql, "1=0") {
		t.Errorf("Expected 1=0 for empty IN, got: %s", sql)
	}
	if len(params) != 0 {
		t.Errorf("Expected 0 params for empty IN, got %d", len(params))
	}
}

func TestInLookupWithSingleValue(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&TestArticle{})

	qs := NewQuerySet[TestArticle]().Filter(Lookup{"Views__in": 42})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM test_article WHERE (views IN (?))"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if !reflect.DeepEqual(params, []any{42}) {
		t.Errorf("Params mismatch: got %v", params)
	}
}

// --- Range lookup tests ---

func TestRangeLookup(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&TestArticle{})

	qs := NewQuerySet[TestArticle]().Filter(Lookup{"Views__range": []int{10, 100}})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM test_article WHERE (views BETWEEN ? AND ?)"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if !reflect.DeepEqual(params, []any{10, 100}) {
		t.Errorf("Params mismatch: got %v", params)
	}
}

// --- FK path lookup tests (JOINs in WHERE) ---

func TestFKPathLookupInWhere(t *testing.T) {
	setupJoinModels(t)

	qs := NewQuerySet[JoinArticle]().Filter(Lookup{"Author__Username__exact": "alice"})
	sql, params := qs.query.ToSQL()

	// Should create a LEFT JOIN for Author and filter on T1.username.
	if !contains(sql, "LEFT JOIN joinuser AS T1 ON T0.author_id = T1.id") {
		t.Errorf("Expected LEFT JOIN for Author path, got: %s", sql)
	}
	if !contains(sql, "T1.username = ?") {
		t.Errorf("Expected T1.username = ? in WHERE, got: %s", sql)
	}
	if !reflect.DeepEqual(params, []any{"alice"}) {
		t.Errorf("Params mismatch: got %v", params)
	}
}

func TestFKPathLookupIContains(t *testing.T) {
	setupJoinModels(t)

	qs := NewQuerySet[JoinArticle]().Filter(Lookup{"Author__Username__icontains": "alic"})
	sql, params := qs.query.ToSQL()

	if !contains(sql, "LOWER(T1.username) LIKE LOWER(?)") {
		t.Errorf("Expected LOWER(T1.username) LIKE LOWER(?), got: %s", sql)
	}
	if len(params) != 1 {
		t.Errorf("Expected 1 param, got %d", len(params))
	}
}

func TestNestedFKPathLookup(t *testing.T) {
	setupJoinModels(t)

	// Article → Profile → User → Username
	qs := NewQuerySet[JoinArticle]().Filter(Lookup{"Profile__User__Username__exact": "bob"})
	sql, params := qs.query.ToSQL()

	// Should create two JOINs: Profile (T1) and then Profile.User (T2).
	if !contains(sql, "LEFT JOIN joinprofile AS T1 ON T0.profile_id = T1.id") {
		t.Errorf("Expected LEFT JOIN for Profile, got: %s", sql)
	}
	if !contains(sql, "LEFT JOIN joinuser AS T2 ON T1.user_id = T2.id") {
		t.Errorf("Expected LEFT JOIN for Profile.User, got: %s", sql)
	}
	if !contains(sql, "T2.username = ?") {
		t.Errorf("Expected T2.username = ? in WHERE, got: %s", sql)
	}
	if !reflect.DeepEqual(params, []any{"bob"}) {
		t.Errorf("Params mismatch: got %v", params)
	}
}

func TestFKPathLookupWithIn(t *testing.T) {
	setupJoinModels(t)

	qs := NewQuerySet[JoinArticle]().Filter(Lookup{"Author__Username__in": []string{"alice", "bob"}})
	sql, params := qs.query.ToSQL()

	if !contains(sql, "T1.username IN (?, ?)") {
		t.Errorf("Expected T1.username IN (?, ?), got: %s", sql)
	}
	if !reflect.DeepEqual(params, []any{"alice", "bob"}) {
		t.Errorf("Params mismatch: got %v", params)
	}
}

func TestFKDirectLookupWithModelInstance(t *testing.T) {
	setupJoinModels(t)

	user := &JoinUser{}
	user.ID = 42

	qs := NewQuerySet[JoinArticle]().Filter(Lookup{"Author": user})
	sql, params := qs.query.ToSQL()

	// Should extract PK from the model instance and filter on the FK column.
	if !contains(sql, "author_id = ?") {
		t.Errorf("Expected author_id = ?, got: %s", sql)
	}
	if !reflect.DeepEqual(params, []any{uint64(42)}) {
		t.Errorf("Params mismatch: got %v", params)
	}
}

func TestFKDirectLookupWithIntValue(t *testing.T) {
	setupJoinModels(t)

	qs := NewQuerySet[JoinArticle]().Filter(Lookup{"Author": 42})
	sql, params := qs.query.ToSQL()

	if !contains(sql, "author_id = ?") {
		t.Errorf("Expected author_id = ?, got: %s", sql)
	}
	if !reflect.DeepEqual(params, []any{42}) {
		t.Errorf("Params mismatch: got %v", params)
	}
}

// --- ORDER BY with FK path ---

func TestOrderByWithFKPath(t *testing.T) {
	setupJoinModels(t)

	qs := NewQuerySet[JoinArticle]().OrderBy("Author__Username")
	sql, _ := qs.query.ToSQL()

	// Should create a JOIN and use T1.username in ORDER BY.
	if !contains(sql, "LEFT JOIN joinuser AS T1 ON T0.author_id = T1.id") {
		t.Errorf("Expected LEFT JOIN for Author in ORDER BY, got: %s", sql)
	}
	if !contains(sql, "ORDER BY T1.username ASC") {
		t.Errorf("Expected ORDER BY T1.username ASC, got: %s", sql)
	}
}

func TestOrderByWithFKPathDesc(t *testing.T) {
	setupJoinModels(t)

	qs := NewQuerySet[JoinArticle]().OrderBy("-Author__Username")
	sql, _ := qs.query.ToSQL()

	if !contains(sql, "ORDER BY T1.username DESC") {
		t.Errorf("Expected ORDER BY T1.username DESC, got: %s", sql)
	}
}

// --- pk lookup ---

func TestPKLookup(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&TestArticle{})

	qs := NewQuerySet[TestArticle]().Filter(Lookup{"pk": 1})
	sql, params := qs.query.ToSQL()

	expectedSQL := "SELECT * FROM test_article WHERE (id = ?)"
	if sql != expectedSQL {
		t.Errorf("Expected %q, got %q", expectedSQL, sql)
	}
	if !reflect.DeepEqual(params, []any{1}) {
		t.Errorf("Params mismatch: got %v", params)
	}
}

// --- Combined SelectRelated + Filter ---

func TestSelectRelatedAndFilterCombined(t *testing.T) {
	setupJoinModels(t)

	qs := NewQuerySet[JoinArticle]().
		SelectRelated(1).
		Filter(Lookup{"Author__Username__exact": "alice"}).
		OrderBy("-Author__Username")

	sql, params := qs.query.ToSQL()

	// The Author join should only appear once (shared between SelectRelated and filter).
	joinCount := countOccurrences(sql, "LEFT JOIN joinuser AS T1")
	if joinCount != 1 {
		t.Errorf("Expected exactly 1 LEFT JOIN for joinuser T1, got %d in: %s", joinCount, sql)
	}

	if !contains(sql, "T1.username = ?") {
		t.Errorf("Expected T1.username = ? in WHERE, got: %s", sql)
	}
	if !contains(sql, "ORDER BY T1.username DESC") {
		t.Errorf("Expected ORDER BY T1.username DESC, got: %s", sql)
	}
	if !reflect.DeepEqual(params, []any{"alice"}) {
		t.Errorf("Params mismatch: got %v", params)
	}
}

// --- Only with SelectRelated ---

func TestOnlyWithSelectRelated(t *testing.T) {
	setupJoinModels(t)

	qs := NewQuerySet[JoinArticle]().Only("Title", "Author").SelectRelated(1)
	sql, _ := qs.query.ToSQL()

	// Should qualify OnlyFields with T0 alias when joins exist.
	if !contains(sql, "T0.title") {
		t.Errorf("Expected T0.title in SELECT, got: %s", sql)
	}
	if !contains(sql, "T0.author_id") {
		t.Errorf("Expected T0.author_id in SELECT, got: %s", sql)
	}
}

// --- helpers ---
// (contains, containsStr are defined in case_test.go)

func countOccurrences(s, substr string) int {
	count := 0
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			count++
			i += len(substr) - 1
		}
	}
	return count
}
