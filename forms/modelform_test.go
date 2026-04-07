package forms

import (
	"github.com/godjango/godjango/orm"
	"testing"
)

type Profile struct {
	orm.Model
	Username string `gd:"CharField,max_length=50"`
	Age      int    `gd:"IntegerField"`
	IsActive bool   `gd:"BooleanField,default=true"`
	Bio      string `gd:"TextField,blank=true"`
}

func TestModelForm(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&Profile{})

	p := &Profile{Username: "john_doe", Age: 25}

	// Create ModelForm
	mf, err := NewModelForm(p, nil, nil)
	if err != nil {
		t.Fatalf("ModelForm init failed: %v", err)
	}

	// Verify fields generated
	if len(mf.Fields) != 4 {
		t.Errorf("Expected 4 fields generated, got %d", len(mf.Fields))
	}

	if _, ok := mf.Fields["Username"].(*CharField); !ok {
		t.Errorf("Expected CharField for Username")
	}
	if _, ok := mf.Fields["Age"].(*IntegerField); !ok {
		t.Errorf("Expected IntegerField for Age")
	}
	if _, ok := mf.Fields["IsActive"].(*BooleanField); !ok {
		t.Errorf("Expected BooleanField for IsActive")
	}

	// Verify Initial data loaded
	if mf.Data["Username"] != "john_doe" {
		t.Errorf("Initial data missing")
	}

	// Bind data to simulate POST
	postData := map[string]any{
		"Username": "jane_doe",
		"Age":      "30",
		"IsActive": "true", // bound boolean
		"Bio":      "Developer",
	}
	mf.Bind(postData, nil)

	// Validate
	if !mf.IsValid() {
		t.Errorf("ModelForm should be valid, got errors: %v", mf.Errors)
	}

	// Save
	savedInstance, err := mf.Save(false)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	savedProfile := savedInstance.(*Profile)
	if savedProfile.Username != "jane_doe" {
		t.Errorf("Model missing updated Username")
	}
	if savedProfile.Age != 30 {
		t.Errorf("Model missing updated Age")
	}
	if savedProfile.Bio != "Developer" {
		t.Errorf("Model missing updated Bio")
	}
	if savedProfile.IsActive != true {
		t.Errorf("Model missing updated IsActive")
	}
}

func TestModelFormExcludes(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&Profile{})

	p := &Profile{}
	mf, _ := NewModelForm(p, nil, []string{"Bio", "IsActive"})

	if len(mf.Fields) != 2 {
		t.Errorf("Expected 2 fields after excludes, got %d", len(mf.Fields))
	}
	if _, ok := mf.Fields["Username"]; !ok {
		t.Errorf("Missing Username")
	}
	if _, ok := mf.Fields["Age"]; !ok {
		t.Errorf("Missing Age")
	}
	if _, ok := mf.Fields["Bio"]; ok {
		t.Errorf("Bio should be excluded")
	}
}
