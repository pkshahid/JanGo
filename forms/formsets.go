package forms

import (
	"fmt"
	"mime/multipart"
)

// BaseFormSet manages a collection of instances of a specific form.
type BaseFormSet struct {
	FormFactory func() *Form // Factory to create the forms
	Forms       []*Form      // The instantiated forms
	Data        map[string]any
	Files       map[string][]*multipart.FileHeader
	Prefix      string

	TotalFormCount   int
	InitialFormCount int
	MaxNum           int

	Errors  []FormErrors
	IsBound bool
}

// NewBaseFormSet initializes a BaseFormSet.
func NewBaseFormSet(factory func() *Form, prefix string, total, initial, maxNum int) *BaseFormSet {
	if prefix == "" {
		prefix = "form"
	}

	fs := &BaseFormSet{
		FormFactory:      factory,
		Prefix:           prefix,
		TotalFormCount:   total,
		InitialFormCount: initial,
		MaxNum:           maxNum,
	}

	fs.Forms = make([]*Form, total)
	for i := 0; i < total; i++ {
		form := factory()
		// Rewrite form fields to use prefix
		fs.applyPrefix(form, i)
		fs.Forms[i] = form
	}

	return fs
}

func (fs *BaseFormSet) applyPrefix(form *Form, index int) {
	prefixedFields := make(map[string]Field)
	var newOrder []string

	for _, name := range form.Order {
		field := form.Fields[name]
		prefixedName := fmt.Sprintf("%s-%d-%s", fs.Prefix, index, name)
		prefixedFields[prefixedName] = field
		newOrder = append(newOrder, prefixedName)
	}

	form.Fields = prefixedFields
	form.Order = newOrder
}

// Bind assigns data to all forms in the set.
func (fs *BaseFormSet) Bind(data map[string]any, files map[string][]*multipart.FileHeader) {
	fs.Data = data
	fs.Files = files
	fs.IsBound = true

	// A real implementation would parse the management form (e.g. `form-TOTAL_FORMS`)
	// to know how many forms are actually submitted.
	// For proto, we assume the submitted forms match `TotalFormCount`.
	for _, form := range fs.Forms {
		form.Bind(data, files)
	}
}

// IsValid runs validation on all forms.
func (fs *BaseFormSet) IsValid() bool {
	if !fs.IsBound {
		return false
	}

	valid := true
	fs.Errors = make([]FormErrors, len(fs.Forms))

	for i, form := range fs.Forms {
		if !form.IsValid() {
			valid = false
			fs.Errors[i] = form.Errors
		}
	}
	return valid
}

// ModelFormSetFactory generates a factory for ModelFormSets.
// In Django this returns a class. Here it returns a configured BaseFormSet initialization function.
func ModelFormSetFactory(modelInstance any, includes, excludes []string, extra int) func() *BaseFormSet {
	return func() *BaseFormSet {
		factory := func() *Form {
			mf, _ := NewModelForm(modelInstance, includes, excludes)
			return &mf.Form
		}
		// Typically, a ModelFormSet queries the DB for initial items, sets InitialFormCount to len(results),
		// and TotalFormCount to len(results) + extra.
		// For this prototype, we just mock the empty 'extra' forms.
		return NewBaseFormSet(factory, "form", extra, 0, 1000)
	}
}
