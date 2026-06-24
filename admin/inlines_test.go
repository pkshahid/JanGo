package admin

import (
	"testing"
)

func TestNewInlineFormset(t *testing.T) {
	inline := &InlineModelAdmin{
		FkName: "article_id",
		Extra:  2,
	}

	fs := NewInlineFormset(inline, 1)
	if fs.TotalFormCount() != 2 {
		t.Errorf("expected 2 forms (extra), got %d", fs.TotalFormCount())
	}
	if fs.InitialFormCount() != 0 {
		t.Errorf("expected 0 initial forms, got %d", fs.InitialFormCount())
	}
	if fs.ExtraFormCount() != 2 {
		t.Errorf("expected 2 extra forms, got %d", fs.ExtraFormCount())
	}
}

func TestInlineFormsetAddExisting(t *testing.T) {
	inline := &InlineModelAdmin{
		FkName: "article_id",
		Extra:  1,
	}

	fs := NewInlineFormset(inline, 1)
	fs.AddExistingForm("obj1", map[string]interface{}{"text": "hello"})
	fs.AddExistingForm("obj2", map[string]interface{}{"text": "world"})

	if fs.TotalFormCount() != 3 {
		t.Errorf("expected 3 forms, got %d", fs.TotalFormCount())
	}
	if fs.InitialFormCount() != 2 {
		t.Errorf("expected 2 initial, got %d", fs.InitialFormCount())
	}
	if fs.ExtraFormCount() != 1 {
		t.Errorf("expected 1 extra, got %d", fs.ExtraFormCount())
	}
}

func TestInlineFormsetValidate(t *testing.T) {
	inline := &InlineModelAdmin{
		FkName: "article_id",
		MinNum: 2,
		MaxNum: 5,
		Extra:  1,
	}

	fs := NewInlineFormset(inline, 1)
	errors := fs.Validate()
	if len(errors) == 0 {
		t.Error("expected validation error for minimum forms")
	}

	// Add enough forms
	fs.AddExistingForm("obj1", map[string]interface{}{"x": 1})
	fs.AddExistingForm("obj2", map[string]interface{}{"x": 2})
	errors = fs.Validate()
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

func TestInlineFormsetMaxValidation(t *testing.T) {
	inline := &InlineModelAdmin{
		FkName: "article_id",
		MaxNum: 2,
		Extra:  0,
	}

	fs := NewInlineFormset(inline, 1)
	fs.AddExistingForm("obj1", map[string]interface{}{"x": 1})
	fs.AddExistingForm("obj2", map[string]interface{}{"x": 2})
	fs.AddExistingForm("obj3", map[string]interface{}{"x": 3})

	errors := fs.Validate()
	if len(errors) == 0 {
		t.Error("expected validation error for maximum forms")
	}
}

func TestTabularInlineAdmin(t *testing.T) {
	inline := TabularInlineAdmin(struct{}{}, "parent_id")
	if inline.Template != string(TabularInline) {
		t.Errorf("expected tabular template, got %q", inline.Template)
	}
	if inline.Extra != 3 {
		t.Errorf("expected extra=3, got %d", inline.Extra)
	}
}

func TestStackedInlineAdmin(t *testing.T) {
	inline := StackedInlineAdmin(struct{}{}, "parent_id")
	if inline.Template != string(StackedInline) {
		t.Errorf("expected stacked template, got %q", inline.Template)
	}
}

func TestGetInlineFormsets(t *testing.T) {
	ma := &ModelAdmin{
		InlineModels: []InlineModelAdmin{
			TabularInlineAdmin(struct{}{}, "fk1"),
			StackedInlineAdmin(struct{}{}, "fk2"),
		},
	}

	formsets := GetInlineFormsets(ma, 42)
	if len(formsets) != 2 {
		t.Errorf("expected 2 formsets, got %d", len(formsets))
	}
}

func TestDefaultExtra(t *testing.T) {
	inline := &InlineModelAdmin{
		FkName: "test_id",
		Extra:  0, // 0 means use default (3)
	}

	fs := NewInlineFormset(inline, 1)
	if fs.TotalFormCount() != 3 {
		t.Errorf("expected 3 forms (default extra), got %d", fs.TotalFormCount())
	}
}
