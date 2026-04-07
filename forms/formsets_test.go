package forms

import (
	"github.com/godjango/godjango/orm"
	"strings"
	"testing"
)

func TestFormSets(t *testing.T) {
	// Basic FormFactory
	factory := func() *Form {
		f := NewContactForm()
	return &f.Form
	}

	formset := NewBaseFormSet(factory, "testprefix", 2, 0, 2)

	if len(formset.Forms) != 2 {
		t.Fatalf("Expected 2 forms in formset")
	}

	// Verify prefixes
	f1 := formset.Forms[0]
	if _, ok := f1.Fields["testprefix-0-subject"]; !ok {
		t.Errorf("Prefix application failed on first form")
	}

	html := string(f1.Render())
	if !strings.Contains(html, `name="testprefix-0-subject"`) {
		t.Errorf("Form rendering failed to use prefixed names: %s", html)
	}

	// Binding
	data := map[string]any{
		"testprefix-0-subject": "A",
		"testprefix-0-message": "B",
		"testprefix-0-sender":  "test@example.com",
		"testprefix-1-subject": "C", // missing other required fields
	}

	formset.Bind(data, nil)
	isValid := formset.IsValid()

	if isValid {
		t.Errorf("Formset should be invalid")
	}

	if len(formset.Errors) != 2 {
		t.Errorf("Expected errors slice of len 2")
	}

	if len(formset.Errors[0]) > 0 {
		t.Errorf("First form should be valid, got errors: %v", formset.Errors[0])
	}
	if len(formset.Errors[1]) == 0 {
		t.Errorf("Second form should have errors")
	}
}

func TestModelFormSetFactory(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&Profile{})

	p := &Profile{}
	factory := ModelFormSetFactory(p, nil, nil, 3)
	formset := factory()

	if len(formset.Forms) != 3 {
		t.Errorf("Expected 3 forms in model formset")
	}

	// Each should have Username, Age, IsActive, Bio, appropriately prefixed
	f := formset.Forms[0]
	if _, ok := f.Fields["form-0-Username"]; !ok {
		t.Errorf("Model formset prefix generation failed")
	}
}
