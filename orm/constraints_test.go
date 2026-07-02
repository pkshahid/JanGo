package orm

import (
	"strings"
	"testing"
)

// --- Test models for constraints ---

type ConstraintProduct struct {
	Model
	Name     string  `gd:"CharField,max_length=200"`
	Price    float64 `gd:"DecimalField,max_digits=10,decimal_places=2"`
	Quantity int     `gd:"IntegerField"`
	Active   bool    `gd:"BooleanField,default=true"`
	Category string  `gd:"CharField,max_length=100,blank=true"`
}

func (p *ConstraintProduct) ModelMeta() *Meta {
	return &Meta{
		DbTable: "constraint_product",
		Constraints: []Constraint{
			CheckConstraint{
				Name:  "price_non_negative",
				Check: "price >= 0",
				Validator: func(obj any) bool {
					p, ok := obj.(*ConstraintProduct)
					if !ok {
						return false
					}
					return p.Price >= 0
				},
			},
			CheckConstraint{
				Name:  "quantity_non_negative",
				Check: "quantity >= 0",
				Validator: func(obj any) bool {
					p, ok := obj.(*ConstraintProduct)
					if !ok {
						return false
					}
					return p.Quantity >= 0
				},
			},
			UniqueConstraint{
				Name:      "unique_active_name",
				Fields:    []string{"Name", "Category"},
				Condition: "active = 1",
				ConditionValidator: func(obj any) bool {
					p, ok := obj.(*ConstraintProduct)
					if !ok {
						return false
					}
					return p.Active
				},
			},
		},
	}
}

type SimpleCheckModel struct {
	Model
	Value int `gd:"IntegerField"`
}

func (m *SimpleCheckModel) ModelMeta() *Meta {
	return &Meta{
		DbTable: "simple_check_model",
		Constraints: []Constraint{
			CheckConstraint{
				Name:  "value_positive",
				Check: "value > 0",
				Validator: func(obj any) bool {
					m, ok := obj.(*SimpleCheckModel)
					if !ok {
						return false
					}
					return m.Value > 0
				},
			},
		},
	}
}

type UniqueConstraintModel struct {
	Model
	Name   string `gd:"CharField,max_length=200"`
	Active bool   `gd:"BooleanField,default=true"`
}

func (m *UniqueConstraintModel) ModelMeta() *Meta {
	return &Meta{
		DbTable: "unique_constraint_model",
		Constraints: []Constraint{
			UniqueConstraint{
				Name:      "unique_name_when_active",
				Fields:    []string{"Name"},
				Condition: "active = 1",
				ConditionValidator: func(obj any) bool {
					m, ok := obj.(*UniqueConstraintModel)
					if !ok {
						return false
					}
					return m.Active
				},
			},
		},
	}
}

// --- CheckConstraint tests ---

func TestCheckConstraint_Valid(t *testing.T) {
	ClearRegistry()
	Register(&ConstraintProduct{})

	product := &ConstraintProduct{
		Name:     "Widget",
		Price:    9.99,
		Quantity: 10,
		Active:   true,
		Category: "Electronics",
	}

	if err := FullClean(product); err != nil {
		t.Errorf("expected no error for valid product, got: %v", err)
	}
}

func TestCheckConstraint_Violated(t *testing.T) {
	ClearRegistry()
	Register(&ConstraintProduct{})

	product := &ConstraintProduct{
		Name:     "Widget",
		Price:    -1.0, // violates price_non_negative
		Quantity: 10,
		Active:   true,
	}

	err := FullClean(product)
	if err == nil {
		t.Fatal("expected check constraint error, got nil")
	}

	verrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	found := false
	for _, e := range verrs {
		if e.Code == "check_constraint" && strings.Contains(e.Message, "price_non_negative") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected price_non_negative check constraint error, got: %v", verrs)
	}
}

func TestCheckConstraint_MultipleViolations(t *testing.T) {
	ClearRegistry()
	Register(&ConstraintProduct{})

	product := &ConstraintProduct{
		Name:     "Widget",
		Price:    -1.0, // violates price_non_negative
		Quantity: -5,   // violates quantity_non_negative
		Active:   true,
	}

	err := FullClean(product)
	if err == nil {
		t.Fatal("expected multiple check constraint errors, got nil")
	}

	verrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	checkCount := 0
	for _, e := range verrs {
		if e.Code == "check_constraint" {
			checkCount++
		}
	}
	if checkCount != 2 {
		t.Errorf("expected 2 check_constraint errors, got %d: %v", checkCount, verrs)
	}
}

func TestCheckConstraint_SkipConstraintValidation(t *testing.T) {
	ClearRegistry()
	Register(&SimpleCheckModel{})

	m := &SimpleCheckModel{Value: -1} // violates value_positive

	if err := FullClean(m, SkipConstraintValidation()); err != nil {
		t.Errorf("expected no error with SkipConstraintValidation(), got: %v", err)
	}
}

func TestCheckConstraint_NilValidator(t *testing.T) {
	ClearRegistry()
	Register(&SimpleCheckModel{})

	// Override with a constraint that has no Validator
	c := CheckConstraint{
		Name:      "no_validator",
		Check:     "1 = 1",
		Validator: nil,
	}

	errs := c.Validate(&SimpleCheckModel{Value: -1}, nil, nil)
	if len(errs) != 0 {
		t.Errorf("expected no errors when Validator is nil, got: %v", errs)
	}
}

// --- UniqueConstraint tests ---

func TestUniqueConstraint_ValidWhenActive(t *testing.T) {
	ClearRegistry()
	Register(&UniqueConstraintModel{})

	m := &UniqueConstraintModel{
		Name:   "TestItem",
		Active: true,
	}

	if err := FullClean(m); err != nil {
		t.Errorf("expected no error for valid model with unique constraint, got: %v", err)
	}
}

func TestUniqueConstraint_MissingFieldWhenActive(t *testing.T) {
	ClearRegistry()
	Register(&UniqueConstraintModel{})

	m := &UniqueConstraintModel{
		Name:   "", // empty name should trigger unique constraint error
		Active: true,
	}

	err := FullClean(m)
	if err == nil {
		t.Fatal("expected unique constraint error for missing field, got nil")
	}

	verrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	found := false
	for _, e := range verrs {
		if e.Code == "unique_constraint" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected unique_constraint error, got: %v", verrs)
	}
}

func TestUniqueConstraint_SkippedWhenInactive(t *testing.T) {
	ClearRegistry()
	Register(&UniqueConstraintModel{})

	m := &UniqueConstraintModel{
		Name:   "TestItem", // non-empty to avoid field blank validation
		Active: false,      // condition validator returns false, so constraint doesn't apply
	}

	if err := FullClean(m); err != nil {
		t.Errorf("expected no error when condition validator returns false, got: %v", err)
	}
}

func TestUniqueConstraint_ExcludedField(t *testing.T) {
	ClearRegistry()
	Register(&UniqueConstraintModel{})

	m := &UniqueConstraintModel{
		Name:   "",
		Active: true,
	}

	// Excluding Name should skip the unique constraint
	if err := FullClean(m, ExcludeFields("Name")); err != nil {
		t.Errorf("expected no error with Name excluded, got: %v", err)
	}
}

func TestUniqueConstraint_NoConditionValidator(t *testing.T) {
	ClearRegistry()

	type NoConditionModel struct {
		Model
		Name string `gd:"CharField,max_length=200"`
	}

	// Register with a custom meta that has a UniqueConstraint without ConditionValidator
	Register(&NoConditionModel{})

	// Get the model info and override its Meta
	info, err := GetModelInfo(&NoConditionModel{})
	if err != nil {
		t.Fatal(err)
	}
	info.Meta = &Meta{
		DbTable: "no_condition_model",
		Constraints: []Constraint{
			UniqueConstraint{
				Name:   "unique_name",
				Fields: []string{"Name"},
			},
		},
	}

	m := &NoConditionModel{Name: ""} // empty name should trigger error

	err = FullClean(m)
	if err == nil {
		t.Fatal("expected unique constraint error for missing field, got nil")
	}

	verrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	found := false
	for _, e := range verrs {
		if e.Code == "unique_constraint" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected unique_constraint error, got: %v", verrs)
	}
}

// --- Combined constraint + unique_together tests ---

type CombinedConstraintModel struct {
	Model
	Title    string  `gd:"CharField,max_length=200"`
	Price    float64 `gd:"DecimalField,max_digits=10,decimal_places=2"`
	Category string  `gd:"CharField,max_length=100"`
}

func (m *CombinedConstraintModel) ModelMeta() *Meta {
	return &Meta{
		DbTable:        "combined_constraint_model",
		UniqueTogether: [][]string{{"Title", "Category"}},
		Constraints: []Constraint{
			CheckConstraint{
				Name:  "price_positive",
				Check: "price > 0",
				Validator: func(obj any) bool {
					m, ok := obj.(*CombinedConstraintModel)
					if !ok {
						return false
					}
					return m.Price > 0
				},
			},
		},
	}
}

func TestFullClean_ConstraintsAndUniqueTogether(t *testing.T) {
	ClearRegistry()
	Register(&CombinedConstraintModel{})

	// Both check constraint and unique_together should fail
	m := &CombinedConstraintModel{
		Title:    "",
		Price:    -1.0,
		Category: "",
	}

	err := FullClean(m)
	if err == nil {
		t.Fatal("expected errors, got nil")
	}

	verrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	hasCheck := false
	hasUnique := false
	for _, e := range verrs {
		if e.Code == "check_constraint" {
			hasCheck = true
		}
		if e.Code == "unique_together" {
			hasUnique = true
		}
	}
	if !hasCheck {
		t.Errorf("expected check_constraint error, got: %v", verrs)
	}
	if !hasUnique {
		t.Errorf("expected unique_together error, got: %v", verrs)
	}
}

// --- Constraint interface tests ---

func TestCheckConstraint_ConstraintName(t *testing.T) {
	c := CheckConstraint{Name: "my_check"}
	if c.ConstraintName() != "my_check" {
		t.Errorf("expected 'my_check', got '%s'", c.ConstraintName())
	}
}

func TestUniqueConstraint_ConstraintName(t *testing.T) {
	c := UniqueConstraint{Name: "my_unique"}
	if c.ConstraintName() != "my_unique" {
		t.Errorf("expected 'my_unique', got '%s'", c.ConstraintName())
	}
}

// --- Aggregation with other validation errors ---

func TestFullClean_AggregatesCheckAndFieldErrors(t *testing.T) {
	ClearRegistry()
	Register(&ConstraintProduct{})

	product := &ConstraintProduct{
		Name:     strings.Repeat("x", 201), // max_length violation
		Price:    -1.0,                     // check constraint violation
		Quantity: 10,
		Active:   true,
	}

	err := FullClean(product)
	if err == nil {
		t.Fatal("expected multiple errors, got nil")
	}

	verrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	hasMaxLength := false
	hasCheck := false
	for _, e := range verrs {
		if e.Code == "max_length" {
			hasMaxLength = true
		}
		if e.Code == "check_constraint" {
			hasCheck = true
		}
	}
	if !hasMaxLength {
		t.Errorf("expected max_length error, got: %v", verrs)
	}
	if !hasCheck {
		t.Errorf("expected check_constraint error, got: %v", verrs)
	}
}
