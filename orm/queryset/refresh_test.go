package queryset

import (
	"testing"

	"github.com/pkshahid/JanGo/orm"
)

type refreshTestModel struct {
	orm.Model
	Username string `gd:"CharField,max_length=150"`
	IsActive bool   `gd:"BooleanField,default=true"`
}

func TestRefreshFromDB_NilPointer(t *testing.T) {
	err := RefreshFromDB(nil)
	if err == nil {
		t.Fatal("Expected error for nil pointer")
	}
}

func TestRefreshFromDB_NonPointer(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&refreshTestModel{})

	m := refreshTestModel{Model: orm.Model{ID: 1}}
	err := RefreshFromDB(m)
	if err == nil {
		t.Fatal("Expected error for non-pointer")
	}
}

func TestRefreshFromDB_UnregisteredModel(t *testing.T) {
	orm.ClearRegistry()

	type Unregistered struct {
		orm.Model
		Name string `gd:"CharField"`
	}
	m := &Unregistered{Model: orm.Model{ID: 1}}
	err := RefreshFromDB(m)
	if err == nil {
		t.Fatal("Expected error for unregistered model")
	}
}

func TestRefreshFromDB_ZeroPK(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&refreshTestModel{})

	m := &refreshTestModel{}
	err := RefreshFromDB(m)
	if err == nil {
		t.Fatal("Expected error for zero primary key")
	}
}

func TestRefreshFromDB_UnknownField(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&refreshTestModel{})

	m := &refreshTestModel{Model: orm.Model{ID: 1}}
	err := RefreshFromDB(m, RefreshOptions{Fields: []string{"NonExistent"}})
	if err == nil {
		t.Fatal("Expected error for unknown field name")
	}
}

func TestRefreshFromDB_NoBackend(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&refreshTestModel{})

	m := &refreshTestModel{Model: orm.Model{ID: 1}}
	err := RefreshFromDB(m)
	if err == nil {
		t.Fatal("Expected error when no backend is configured")
	}
}

func TestRefreshFromDB_ValidFieldOption(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&refreshTestModel{})

	m := &refreshTestModel{Model: orm.Model{ID: 5}}

	// Valid field names should pass field validation and fail at DB lookup.
	err := RefreshFromDB(m, RefreshOptions{Fields: []string{"Username"}})
	if err == nil {
		t.Fatal("Expected DB error (no backend), but field validation should have passed")
	}
}
