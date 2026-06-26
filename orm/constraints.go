package orm

import (
	"fmt"
	"reflect"
	"strings"
)

// Constraint is the interface for model-level constraints, similar to Django's
// models.BaseConstraint. Implementations include CheckConstraint and
// UniqueConstraint.
type Constraint interface {
	ConstraintName() string
	Validate(obj any, info *ModelInfo, exclude map[string]bool) ValidationErrors
}

// CheckConstraint enforces a database-level CHECK constraint and validates the
// condition in Go via a Validator function. It mirrors Django's
// models.CheckConstraint(check=Q(...), name=...).
//
// At the database level, the Check string is emitted as
// `CONSTRAINT name CHECK (check)`.
//
// At the Go validation level, the Validator function is called with the model
// instance. If it returns false, a ValidationError is raised.
type CheckConstraint struct {
	Name      string
	Check     string         // SQL expression for DB-level CHECK, e.g. "price > 0"
	Validator func(obj any) bool // Returns true if valid; false triggers a ValidationError
}

func (c CheckConstraint) ConstraintName() string { return c.Name }

func (c CheckConstraint) Validate(obj any, info *ModelInfo, exclude map[string]bool) ValidationErrors {
	if c.Validator == nil {
		return nil
	}
	if !c.Validator(obj) {
		return ValidationErrors{
			&ValidationError{
				Field:   "__all__",
				Message: fmt.Sprintf("check constraint %q failed", c.Name),
				Code:    "check_constraint",
			},
		}
	}
	return nil
}

// UniqueConstraint enforces uniqueness across one or more fields, optionally
// with a condition (partial unique constraint). It mirrors Django's
// models.UniqueConstraint(fields=[...], condition=Q(...), name=...).
//
// When Condition is empty, the constraint is emitted inline in CREATE TABLE as
// `CONSTRAINT name UNIQUE (col1, col2)`.
//
// When Condition is set, a partial unique index is created after the table:
// `CREATE UNIQUE INDEX name ON table (col1, col2) WHERE condition`.
//
// At the Go validation level, the ConditionValidator function (if non-nil)
// determines whether the unique constraint should be enforced for this
// instance. If nil, the constraint is always enforced. When enforced, all
// fields must be non-zero (same semantics as UniqueTogether).
type UniqueConstraint struct {
	Name               string
	Fields             []string
	Condition          string         // SQL WHERE clause for partial unique index, e.g. "active = 1"
	ConditionValidator func(obj any) bool // If set, constraint is only enforced when this returns true
}

func (c UniqueConstraint) ConstraintName() string { return c.Name }

func (c UniqueConstraint) Validate(obj any, info *ModelInfo, exclude map[string]bool) ValidationErrors {
	// Skip if any field is excluded
	for _, f := range c.Fields {
		if exclude[f] {
			return nil
		}
	}

	// Check condition validator — if it returns false, constraint doesn't apply
	if c.ConditionValidator != nil && !c.ConditionValidator(obj) {
		return nil
	}

	// Check that all fields are non-zero (same logic as validateUniqueTogether)
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var zeroFields []string
	for _, fieldName := range c.Fields {
		fv := v.FieldByName(fieldName)
		if !fv.IsValid() || isZeroValue(fv) {
			zeroFields = append(zeroFields, fieldName)
		}
	}

	if len(zeroFields) > 0 {
		return ValidationErrors{
			&ValidationError{
				Field:   "__all__",
				Message: fmt.Sprintf("fields %s must be populated for unique constraint %q", strings.Join(c.Fields, ", "), c.Name),
				Code:    "unique_constraint",
			},
		}
	}

	return nil
}

// validateConstraints runs validation for all constraints declared in the
// model's Meta.Constraints. This is called as part of FullClean.
func validateConstraints(obj any, info *ModelInfo, exclude map[string]bool) ValidationErrors {
	var errs ValidationErrors

	if len(info.Meta.Constraints) == 0 {
		return nil
	}

	for _, c := range info.Meta.Constraints {
		if c == nil {
			continue
		}
		errs = append(errs, c.Validate(obj, info, exclude)...)
	}

	return errs
}
