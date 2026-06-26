package orm

import (
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"
)

// ValidationError represents a single validation failure associated with a field.
type ValidationError struct {
	Field   string
	Message string
	Code    string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// ValidationErrors is a collection of ValidationErrors, similar to Django's
// ValidationErrorList. It allows FullClean to aggregate all errors before
// returning.
type ValidationErrors []*ValidationError

func (e ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// HasField returns true if any error in the collection targets the given field.
func (e ValidationErrors) HasField(field string) bool {
	for _, err := range e {
		if err.Field == field {
			return true
		}
	}
	return false
}

// AsMap converts the error collection to a map of field -> messages.
func (e ValidationErrors) AsMap() map[string][]string {
	m := make(map[string][]string)
	for _, err := range e {
		m[err.Field] = append(m[err.Field], err.Message)
	}
	return m
}

// Cleaner is an interface that models can implement to provide cross-field
// validation, similar to Django's Model.clean(). The method should return
// an error (typically *ValidationError or ValidationErrors) when the model
// instance is in an invalid state.
type Cleaner interface {
	Clean() error
}

// cleanConfig holds options for FullClean.
type cleanConfig struct {
	exclude            map[string]bool
	validateUnique     bool
	validateFields     bool
	validateClean      bool
	validateConstraints bool
}

// CleanOption configures the behavior of FullClean.
type CleanOption func(*cleanConfig)

// ExcludeFields returns a CleanOption that excludes the given fields from
// all validation steps (field validation, unique checks, and the Clean hook
// still runs but the model can inspect the exclude set if needed).
func ExcludeFields(fields ...string) CleanOption {
	return func(c *cleanConfig) {
		for _, f := range fields {
			c.exclude[f] = true
		}
	}
}

// SkipUniqueValidation disables unique / unique_together constraint validation.
func SkipUniqueValidation() CleanOption {
	return func(c *cleanConfig) {
		c.validateUnique = false
	}
}

// SkipFieldValidation disables individual field-level validation.
func SkipFieldValidation() CleanOption {
	return func(c *cleanConfig) {
		c.validateFields = false
	}
}

// SkipClean disables calling the model's Clean() method.
func SkipClean() CleanOption {
	return func(c *cleanConfig) {
		c.validateClean = false
	}
}

// SkipConstraintValidation disables CheckConstraint and UniqueConstraint
// validation declared in Meta.Constraints.
func SkipConstraintValidation() CleanOption {
	return func(c *cleanConfig) {
		c.validateConstraints = false
	}
}

// FullClean validates a model instance by running four phases, mirroring
// Django's Model.full_clean():
//
//  1. Field-level validation (clean_fields) — checks each non-excluded field
//     against its field-type constraints (max_length, null, blank, etc.).
//  2. Model-level cross-field validation (clean) — calls the model's Clean()
//     method if it implements the Cleaner interface.
//  3. Unique constraint validation (validate_unique) — checks unique_together
//     constraints declared in the model's Meta.
//  4. Constraint validation (validate_constraints) — checks CheckConstraint
//     and UniqueConstraint declared in Meta.Constraints.
//
// All errors are collected and returned together as ValidationErrors.
// Returns nil if the model passes all validation.
func FullClean(obj any, opts ...CleanOption) error {
	cfg := &cleanConfig{
		exclude:             make(map[string]bool),
		validateUnique:      true,
		validateFields:      true,
		validateClean:       true,
		validateConstraints: true,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	info, err := GetModelInfo(obj)
	if err != nil {
		return err
	}

	var errs ValidationErrors

	// Phase 1: Field-level validation
	if cfg.validateFields {
		if fieldErrs := cleanFields(obj, info, cfg.exclude); len(fieldErrs) > 0 {
			errs = append(errs, fieldErrs...)
		}
	}

	// Phase 2: Model-level clean() hook
	if cfg.validateClean {
		if cleaner, ok := obj.(Cleaner); ok {
			if cleanErr := cleaner.Clean(); cleanErr != nil {
				errs = append(errs, normalizeCleanErr(cleanErr)...)
			}
		}
	}

	// Phase 3: Unique constraint validation
	if cfg.validateUnique {
		if uniqueErrs := validateUniqueTogether(obj, info, cfg.exclude); len(uniqueErrs) > 0 {
			errs = append(errs, uniqueErrs...)
		}
	}

	// Phase 4: Constraint validation (CheckConstraint, UniqueConstraint)
	if cfg.validateConstraints {
		if constraintErrs := validateConstraints(obj, info, cfg.exclude); len(constraintErrs) > 0 {
			errs = append(errs, constraintErrs...)
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// cleanFields validates each non-excluded field on the model against its
// field-type constraints. This mirrors Django's Model.clean_fields().
func cleanFields(obj any, info *ModelInfo, exclude map[string]bool) ValidationErrors {
	var errs ValidationErrors

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ValidationErrors{
			{Field: "__all__", Message: "model must be a struct", Code: "invalid"},
		}
	}

	for _, field := range info.Fields {
		if exclude[field.Name] {
			continue
		}

		// Skip auto-created fields (ID, CreatedAt, UpdatedAt, etc.)
		if field.Options.AutoCreated {
			continue
		}

		fv := v.FieldByName(field.Name)
		if !fv.IsValid() {
			continue
		}

		if fieldErr := validateField(field, fv); fieldErr != nil {
			errs = append(errs, fieldErr)
		}
	}

	return errs
}

// validateField checks a single field value against its field-type constraints.
func validateField(field *Field, fv reflect.Value) *ValidationError {
	// Null check: if the field is not nullable and the value is a nil pointer,
	// that's an error.
	if !field.Options.Null && fv.Kind() == reflect.Ptr && fv.IsNil() {
		return &ValidationError{
			Field:   field.Name,
			Message: "this field cannot be null",
			Code:    "null",
		}
	}

	// Dereference pointer for further checks
	val := fv
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil // nullable and nil — OK
		}
		val = val.Elem()
	}

	switch field.Type {
	case CharField, TextField, EmailField, URLField, SlugField, IPAddressField, UUIDField, FileField, ImageField:
		if val.Kind() != reflect.String {
			return &ValidationError{
				Field:   field.Name,
				Message: fmt.Sprintf("expected string value, got %s", val.Kind()),
				Code:    "invalid_type",
			}
		}
		s := val.String()
		// Blank check
		if s == "" && !field.Options.Blank {
			return &ValidationError{
				Field:   field.Name,
				Message: "this field cannot be blank",
				Code:    "blank",
			}
		}
		// MaxLength check
		if field.Options.MaxLength > 0 && utf8.RuneCountInString(s) > field.Options.MaxLength {
			return &ValidationError{
				Field:   field.Name,
				Message: fmt.Sprintf("ensure this value has at most %d characters (it has %d)", field.Options.MaxLength, utf8.RuneCountInString(s)),
				Code:    "max_length",
			}
		}

	case IntegerField, SmallIntegerField, BigIntegerField, BigAutoField:
		if val.Kind() != reflect.Int && val.Kind() != reflect.Int8 &&
			val.Kind() != reflect.Int16 && val.Kind() != reflect.Int32 &&
			val.Kind() != reflect.Int64 && val.Kind() != reflect.Uint &&
			val.Kind() != reflect.Uint8 && val.Kind() != reflect.Uint16 &&
			val.Kind() != reflect.Uint32 && val.Kind() != reflect.Uint64 {
			return &ValidationError{
				Field:   field.Name,
				Message: fmt.Sprintf("expected integer value, got %s", val.Kind()),
				Code:    "invalid_type",
			}
		}

	case FloatField:
		if val.Kind() != reflect.Float32 && val.Kind() != reflect.Float64 {
			return &ValidationError{
				Field:   field.Name,
				Message: fmt.Sprintf("expected float value, got %s", val.Kind()),
				Code:    "invalid_type",
			}
		}

	case DecimalField:
		if val.Kind() != reflect.Float32 && val.Kind() != reflect.Float64 {
			return &ValidationError{
				Field:   field.Name,
				Message: fmt.Sprintf("expected decimal value, got %s", val.Kind()),
				Code:    "invalid_type",
			}
		}

	case BooleanField:
		if val.Kind() != reflect.Bool {
			return &ValidationError{
				Field:   field.Name,
				Message: fmt.Sprintf("expected boolean value, got %s", val.Kind()),
				Code:    "invalid_type",
			}
		}

	case NullBooleanField:
		if val.Kind() != reflect.Bool && val.Kind() != reflect.Ptr {
			return &ValidationError{
				Field:   field.Name,
				Message: fmt.Sprintf("expected boolean or *bool value, got %s", val.Kind()),
				Code:    "invalid_type",
			}
		}
	}

	return nil
}

// validateUniqueTogether checks that all fields listed in each UniqueTogether
// group are non-zero. A zero value in a unique_together group would make the
// constraint partially satisfied, which Django flags as a validation error.
//
// Note: actual database-level uniqueness checking requires a DB connection and
// is performed at the queryset layer. This method validates that the
// unique_together fields are populated and structurally valid.
func validateUniqueTogether(obj any, info *ModelInfo, exclude map[string]bool) ValidationErrors {
	var errs ValidationErrors

	if len(info.Meta.UniqueTogether) == 0 {
		return nil
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for _, group := range info.Meta.UniqueTogether {
		allPresent := true
		for _, fieldName := range group {
			if exclude[fieldName] {
				allPresent = false
				break
			}
		}
		if !allPresent {
			continue
		}

		var zeroFields []string
		for _, fieldName := range group {
			fv := v.FieldByName(fieldName)
			if !fv.IsValid() {
				zeroFields = append(zeroFields, fieldName)
				continue
			}
			if isZeroValue(fv) {
				zeroFields = append(zeroFields, fieldName)
			}
		}

		if len(zeroFields) > 0 {
			errs = append(errs, &ValidationError{
				Field:   "__all__",
				Message: fmt.Sprintf("fields %s must be populated for unique_together constraint", strings.Join(group, ", ")),
				Code:    "unique_together",
			})
		}
	}

	return errs
}

// isZeroValue returns true if the reflect.Value is the zero value for its type.
func isZeroValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	if v.Kind() == reflect.Ptr {
		return v.IsNil()
	}
	return v.IsZero()
}

// normalizeCleanErr converts the error returned by a model's Clean() method
// into a ValidationErrors slice.
func normalizeCleanErr(err error) ValidationErrors {
	switch e := err.(type) {
	case ValidationErrors:
		return e
	case *ValidationError:
		return ValidationErrors{e}
	default:
		return ValidationErrors{
			{Field: "__all__", Message: err.Error(), Code: "model_clean"},
		}
	}
}
