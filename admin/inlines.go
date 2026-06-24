package admin

import (
	"fmt"
)

// InlineType defines the display style for inline models.
type InlineType string

const (
	TabularInline InlineType = "tabular"
	StackedInline InlineType = "stacked"
)

// InlineFormset represents a set of forms for inline editing.
type InlineFormset struct {
	Inline     *InlineModelAdmin
	Forms      []*InlineForm
	CanDelete  bool
	CanOrder   bool
	Prefix     string
}

// InlineForm represents a single form within an inline formset.
type InlineForm struct {
	Instance    interface{}
	Fields      map[string]interface{}
	IsNew       bool
	OrderIndex  int
	DeleteFlag  bool
}

// NewInlineFormset creates a new inline formset for the given inline configuration.
func NewInlineFormset(inline *InlineModelAdmin, parentID interface{}) *InlineFormset {
	fs := &InlineFormset{
		Inline:    inline,
		CanDelete: true,
		CanOrder:  false,
		Prefix:    fmt.Sprintf("inline_%s", inline.FkName),
	}

	// Create extra empty forms
	extra := inline.Extra
	if extra == 0 {
		extra = 3 // Django default
	}
	for i := 0; i < extra; i++ {
		fs.Forms = append(fs.Forms, &InlineForm{
			IsNew:      true,
			Fields:     make(map[string]interface{}),
			OrderIndex: i,
		})
	}

	return fs
}

// AddExistingForm adds an existing object as a form in the formset.
func (fs *InlineFormset) AddExistingForm(instance interface{}, fields map[string]interface{}) {
	form := &InlineForm{
		Instance:   instance,
		Fields:     fields,
		IsNew:      false,
		OrderIndex: len(fs.Forms),
	}
	// Insert existing forms before new ones
	existing := make([]*InlineForm, 0)
	newForms := make([]*InlineForm, 0)
	for _, f := range fs.Forms {
		if f.IsNew {
			newForms = append(newForms, f)
		} else {
			existing = append(existing, f)
		}
	}
	existing = append(existing, form)
	fs.Forms = append(existing, newForms...)
}

// TotalFormCount returns the total number of forms.
func (fs *InlineFormset) TotalFormCount() int {
	return len(fs.Forms)
}

// InitialFormCount returns the number of pre-filled forms (existing objects).
func (fs *InlineFormset) InitialFormCount() int {
	count := 0
	for _, f := range fs.Forms {
		if !f.IsNew {
			count++
		}
	}
	return count
}

// ExtraFormCount returns the number of empty extra forms.
func (fs *InlineFormset) ExtraFormCount() int {
	count := 0
	for _, f := range fs.Forms {
		if f.IsNew {
			count++
		}
	}
	return count
}

// Validate checks if all forms in the formset are valid.
func (fs *InlineFormset) Validate() []error {
	var errors []error

	minNum := fs.Inline.MinNum
	maxNum := fs.Inline.MaxNum

	filledCount := 0
	for _, f := range fs.Forms {
		if !f.IsNew || len(f.Fields) > 0 {
			filledCount++
		}
	}

	if minNum > 0 && filledCount < minNum {
		errors = append(errors, fmt.Errorf("minimum %d %s required", minNum, fs.Inline.FkName))
	}
	if maxNum > 0 && filledCount > maxNum {
		errors = append(errors, fmt.Errorf("maximum %d %s allowed", maxNum, fs.Inline.FkName))
	}

	return errors
}

// GetInlineFormsets returns all inline formsets for a ModelAdmin.
func GetInlineFormsets(ma *ModelAdmin, parentID interface{}) []*InlineFormset {
	var formsets []*InlineFormset
	for i := range ma.InlineModels {
		fs := NewInlineFormset(&ma.InlineModels[i], parentID)
		formsets = append(formsets, fs)
	}
	return formsets
}

// TabularInlineAdmin is a convenience constructor for tabular inlines.
func TabularInlineAdmin(model interface{}, fkName string) InlineModelAdmin {
	return InlineModelAdmin{
		Model:    model,
		FkName:   fkName,
		Template: string(TabularInline),
		Extra:    3,
	}
}

// StackedInlineAdmin is a convenience constructor for stacked inlines.
func StackedInlineAdmin(model interface{}, fkName string) InlineModelAdmin {
	return InlineModelAdmin{
		Model:    model,
		FkName:   fkName,
		Template: string(StackedInline),
		Extra:    3,
	}
}
