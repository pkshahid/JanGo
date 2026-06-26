package orm

import (
	"strings"
	"testing"
)

// --- Test models ---

type ValidationArticle struct {
	Model
	Title    string  `gd:"CharField,max_length=200"`
	Body     string  `gd:"TextField,blank=true"`
	Author   *User   `gd:"ForeignKey,to=auth.User,on_delete=CASCADE,null=true"`
	Price    float64 `gd:"DecimalField,max_digits=10,decimal_places=2"`
	Quantity int     `gd:"IntegerField"`
}

func (a *ValidationArticle) ModelMeta() *Meta {
	return &Meta{
		DbTable:        "validation_article",
		UniqueTogether: [][]string{{"Title", "Author"}},
	}
}

// Clean implements cross-field validation: Price must be positive when
// Quantity > 0.
func (a *ValidationArticle) Clean() error {
	if a.Quantity > 0 && a.Price <= 0 {
		return &ValidationError{
			Field:   "Price",
			Message: "price must be positive when quantity is greater than zero",
			Code:    "invalid_price",
		}
	}
	return nil
}

// Model without Clean() — only field-level and unique validation applies.
type SimpleModel struct {
	Model
	Name string `gd:"CharField,max_length=50"`
}

// Model that returns multiple errors from Clean().
type MultiErrorModel struct {
	Model
	Start int `gd:"IntegerField"`
	End   int `gd:"IntegerField"`
}

func (m *MultiErrorModel) Clean() error {
	var errs ValidationErrors
	if m.Start < 0 {
		errs = append(errs, &ValidationError{
			Field:   "Start",
			Message: "start must be non-negative",
			Code:    "negative_start",
		})
	}
	if m.End < m.Start {
		errs = append(errs, &ValidationError{
			Field:   "End",
			Message: "end must be greater than or equal to start",
			Code:    "end_before_start",
		})
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Model that returns a plain error from Clean().
type PlainErrorModel struct {
	Model
	Value int `gd:"IntegerField"`
}

func (m *PlainErrorModel) Clean() error {
	if m.Value == 42 {
		return &ValidationError{
			Field:   "__all__",
			Message: "42 is not allowed",
			Code:    "forbidden_value",
		}
	}
	return nil
}

// --- Tests ---

func TestFullClean_ValidModel(t *testing.T) {
	ClearRegistry()
	Register(&ValidationArticle{}, &User{})

	author := &User{Username: "alice"}
	article := &ValidationArticle{
		Title:    "Hello World",
		Body:     "Some content",
		Author:   author,
		Price:    9.99,
		Quantity: 5,
	}

	if err := FullClean(article); err != nil {
		t.Errorf("expected no validation error, got: %v", err)
	}
}

func TestFullClean_MaxLengthExceeded(t *testing.T) {
	ClearRegistry()
	Register(&ValidationArticle{}, &User{})

	article := &ValidationArticle{
		Title:    strings.Repeat("x", 201),
		Author:   &User{Username: "alice"},
		Price:    1.0,
		Quantity: 1,
	}

	err := FullClean(article)
	if err == nil {
		t.Fatal("expected validation error for max_length, got nil")
	}

	verrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	found := false
	for _, e := range verrs {
		if e.Field == "Title" && e.Code == "max_length" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected max_length error on Title, got: %v", verrs)
	}
}

func TestFullClean_BlankNotAllowed(t *testing.T) {
	ClearRegistry()
	Register(&ValidationArticle{}, &User{})

	article := &ValidationArticle{
		Title:    "",
		Author:   &User{Username: "alice"},
		Price:    1.0,
		Quantity: 1,
	}

	err := FullClean(article)
	if err == nil {
		t.Fatal("expected validation error for blank Title, got nil")
	}

	verrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	found := false
	for _, e := range verrs {
		if e.Field == "Title" && e.Code == "blank" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected blank error on Title, got: %v", verrs)
	}
}

func TestFullClean_BlankAllowed(t *testing.T) {
	ClearRegistry()
	Register(&ValidationArticle{}, &User{})

	article := &ValidationArticle{
		Title:    "Hello",
		Body:     "", // blank=true in tag
		Author:   &User{Username: "alice"},
		Price:    1.0,
		Quantity: 1,
	}

	if err := FullClean(article); err != nil {
		t.Errorf("expected no error for blank Body (blank=true), got: %v", err)
	}
}

func TestFullClean_CleanHook(t *testing.T) {
	ClearRegistry()
	Register(&ValidationArticle{}, &User{})

	article := &ValidationArticle{
		Title:    "Hello",
		Author:   &User{Username: "alice"},
		Price:    0, // invalid: Quantity > 0 but Price <= 0
		Quantity: 5,
	}

	err := FullClean(article)
	if err == nil {
		t.Fatal("expected validation error from Clean(), got nil")
	}

	verrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	found := false
	for _, e := range verrs {
		if e.Field == "Price" && e.Code == "invalid_price" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected invalid_price error from Clean(), got: %v", verrs)
	}
}

func TestFullClean_CleanHookMultipleErrors(t *testing.T) {
	ClearRegistry()
	Register(&MultiErrorModel{})

	m := &MultiErrorModel{
		Start: -1,
		End:   -5,
	}

	err := FullClean(m)
	if err == nil {
		t.Fatal("expected validation errors from Clean(), got nil")
	}

	verrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	if len(verrs) < 2 {
		t.Errorf("expected at least 2 errors, got %d: %v", len(verrs), verrs)
	}

	hasStart := false
	hasEnd := false
	for _, e := range verrs {
		if e.Field == "Start" && e.Code == "negative_start" {
			hasStart = true
		}
		if e.Field == "End" && e.Code == "end_before_start" {
			hasEnd = true
		}
	}
	if !hasStart || !hasEnd {
		t.Errorf("expected both Start and End errors, got: %v", verrs)
	}
}

func TestFullClean_CleanHookPlainValidationError(t *testing.T) {
	ClearRegistry()
	Register(&PlainErrorModel{})

	m := &PlainErrorModel{Value: 42}

	err := FullClean(m)
	if err == nil {
		t.Fatal("expected validation error from Clean(), got nil")
	}

	verrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	found := false
	for _, e := range verrs {
		if e.Field == "__all__" && e.Code == "forbidden_value" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected forbidden_value error, got: %v", verrs)
	}
}

func TestFullClean_SkipClean(t *testing.T) {
	ClearRegistry()
	Register(&PlainErrorModel{})

	m := &PlainErrorModel{Value: 42}

	// With SkipClean, the Clean() hook is not called, so no error.
	if err := FullClean(m, SkipClean()); err != nil {
		t.Errorf("expected no error with SkipClean(), got: %v", err)
	}
}

func TestFullClean_SkipFieldValidation(t *testing.T) {
	ClearRegistry()
	Register(&ValidationArticle{}, &User{})

	article := &ValidationArticle{
		Title:    strings.Repeat("x", 201), // would fail max_length
		Author:   &User{Username: "alice"},
		Price:    1.0,
		Quantity: 1,
	}

	// With SkipFieldValidation, field checks are skipped.
	if err := FullClean(article, SkipFieldValidation()); err != nil {
		t.Errorf("expected no error with SkipFieldValidation(), got: %v", err)
	}
}

func TestFullClean_ExcludeFields(t *testing.T) {
	ClearRegistry()
	Register(&ValidationArticle{}, &User{})

	article := &ValidationArticle{
		Title:    strings.Repeat("x", 201), // would fail max_length
		Author:   &User{Username: "alice"},
		Price:    1.0,
		Quantity: 1,
	}

	// Excluding Title should skip its max_length check.
	if err := FullClean(article, ExcludeFields("Title")); err != nil {
		t.Errorf("expected no error with Title excluded, got: %v", err)
	}
}

func TestFullClean_UniqueTogetherMissingFields(t *testing.T) {
	ClearRegistry()
	Register(&ValidationArticle{}, &User{})

	// Author is nil — unique_together {Title, Author} should fail.
	article := &ValidationArticle{
		Title:    "Hello",
		Author:   nil,
		Price:    1.0,
		Quantity: 1,
	}

	err := FullClean(article)
	if err == nil {
		t.Fatal("expected unique_together error, got nil")
	}

	verrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	found := false
	for _, e := range verrs {
		if e.Code == "unique_together" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected unique_together error, got: %v", verrs)
	}
}

func TestFullClean_UniqueTogetherValid(t *testing.T) {
	ClearRegistry()
	Register(&ValidationArticle{}, &User{})

	article := &ValidationArticle{
		Title:    "Hello",
		Author:   &User{Username: "alice"},
		Price:    1.0,
		Quantity: 1,
	}

	// Both Title and Author are populated — unique_together should pass.
	if err := FullClean(article); err != nil {
		t.Errorf("expected no error for valid unique_together, got: %v", err)
	}
}

func TestFullClean_SkipUniqueValidation(t *testing.T) {
	ClearRegistry()
	Register(&ValidationArticle{}, &User{})

	// Author is nil — would fail unique_together, but we skip it.
	article := &ValidationArticle{
		Title:    "Hello",
		Author:   nil,
		Price:    1.0,
		Quantity: 1,
	}

	if err := FullClean(article, SkipUniqueValidation()); err != nil {
		t.Errorf("expected no error with SkipUniqueValidation(), got: %v", err)
	}
}

func TestFullClean_AggregatesAllErrors(t *testing.T) {
	ClearRegistry()
	Register(&ValidationArticle{}, &User{})

	// Multiple problems: Title too long, Clean fails, unique_together fails.
	article := &ValidationArticle{
		Title:    strings.Repeat("x", 300),
		Author:   nil,
		Price:    0,
		Quantity: 10,
	}

	err := FullClean(article)
	if err == nil {
		t.Fatal("expected multiple validation errors, got nil")
	}

	verrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	// Should have: max_length on Title, invalid_price from Clean, unique_together
	if len(verrs) < 3 {
		t.Errorf("expected at least 3 errors, got %d: %v", len(verrs), verrs)
	}
}

func TestFullClean_NoCleanerInterface(t *testing.T) {
	ClearRegistry()
	Register(&SimpleModel{})

	m := &SimpleModel{Name: "test"}

	// SimpleModel does not implement Cleaner — should still pass field validation.
	if err := FullClean(m); err != nil {
		t.Errorf("expected no error for valid SimpleModel, got: %v", err)
	}
}

func TestFullClean_UnregisteredModel(t *testing.T) {
	ClearRegistry()

	type Unregistered struct {
		Model
		Name string `gd:"CharField,max_length=50"`
	}

	err := FullClean(&Unregistered{Name: "test"})
	if err == nil {
		t.Fatal("expected error for unregistered model, got nil")
	}
}

func TestValidationErrors_HasField(t *testing.T) {
	errs := ValidationErrors{
		{Field: "Title", Message: "too long", Code: "max_length"},
		{Field: "Price", Message: "invalid", Code: "invalid_price"},
	}

	if !errs.HasField("Title") {
		t.Error("expected HasField(Title) to be true")
	}
	if errs.HasField("Body") {
		t.Error("expected HasField(Body) to be false")
	}
}

func TestValidationErrors_AsMap(t *testing.T) {
	errs := ValidationErrors{
		{Field: "Title", Message: "too long", Code: "max_length"},
		{Field: "Title", Message: "also bad", Code: "other"},
		{Field: "Price", Message: "invalid", Code: "invalid_price"},
	}

	m := errs.AsMap()
	if len(m["Title"]) != 2 {
		t.Errorf("expected 2 messages for Title, got %d", len(m["Title"]))
	}
	if len(m["Price"]) != 1 {
		t.Errorf("expected 1 message for Price, got %d", len(m["Price"]))
	}
}

func TestValidationError_Error(t *testing.T) {
	e := &ValidationError{Field: "Title", Message: "too long"}
	if e.Error() != "Title: too long" {
		t.Errorf("expected 'Title: too long', got '%s'", e.Error())
	}

	e2 := &ValidationError{Field: "", Message: "generic error"}
	if e2.Error() != "generic error" {
		t.Errorf("expected 'generic error', got '%s'", e2.Error())
	}
}
