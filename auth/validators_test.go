package auth

import (
	"testing"
)

func TestMinimumLengthValidator(t *testing.T) {
	v := NewMinimumLengthValidator(8)

	// Too short
	if err := v.Validate("short", nil); err == nil {
		t.Errorf("Expected error for password shorter than min length")
	}

	// Exactly min length
	if err := v.Validate("12345678", nil); err != nil {
		t.Errorf("Expected no error for password of exactly min length, got: %v", err)
	}

	// Longer than min length
	if err := v.Validate("a-very-secure-password", nil); err != nil {
		t.Errorf("Expected no error for long password, got: %v", err)
	}

	// Help text
	if v.GetHelpText() == "" {
		t.Errorf("Expected non-empty help text")
	}
}

func TestMinimumLengthValidatorDefault(t *testing.T) {
	v := NewMinimumLengthValidator(0)
	if v.MinLength != 8 {
		t.Errorf("Expected default min length of 8, got %d", v.MinLength)
	}
}

func TestCommonPasswordValidator(t *testing.T) {
	v := NewCommonPasswordValidator()

	// Common password
	if err := v.Validate("password", nil); err == nil {
		t.Errorf("Expected error for common password 'password'")
	}

	// Common password (case-insensitive)
	if err := v.Validate("PASSWORD", nil); err == nil {
		t.Errorf("Expected error for common password 'PASSWORD' (case-insensitive)")
	}

	// Common password (numeric)
	if err := v.Validate("123456", nil); err == nil {
		t.Errorf("Expected error for common password '123456'")
	}

	// Non-common password
	if err := v.Validate("x9fK2mPq", nil); err != nil {
		t.Errorf("Expected no error for uncommon password, got: %v", err)
	}

	// Help text
	if v.GetHelpText() == "" {
		t.Errorf("Expected non-empty help text")
	}
}

func TestNumericPasswordValidator(t *testing.T) {
	v := NewNumericPasswordValidator()

	// All numeric
	if err := v.Validate("12345678", nil); err == nil {
		t.Errorf("Expected error for entirely numeric password")
	}

	// Mixed
	if err := v.Validate("abc12345", nil); err != nil {
		t.Errorf("Expected no error for mixed password, got: %v", err)
	}

	// All letters
	if err := v.Validate("abcdefgh", nil); err != nil {
		t.Errorf("Expected no error for all-letter password, got: %v", err)
	}

	// Empty string should not error (required check is on the form)
	if err := v.Validate("", nil); err != nil {
		t.Errorf("Expected no error for empty password, got: %v", err)
	}

	// Help text
	if v.GetHelpText() == "" {
		t.Errorf("Expected non-empty help text")
	}
}

func TestUserAttributeSimilarityValidator(t *testing.T) {
	v := NewUserAttributeSimilarityValidator()
	user := &AbstractUser{
		UsernameStr: "johndoe",
		EmailStr:    "john@example.com",
		FirstName:   "John",
		LastName:    "Doe",
	}

	// Password identical to username
	if err := v.Validate("johndoe", user); err == nil {
		t.Errorf("Expected error for password identical to username")
	}

	// Password identical to first name
	if err := v.Validate("john", user); err == nil {
		t.Errorf("Expected error for password identical to first name")
	}

	// Password very different from user attributes
	if err := v.Validate("x9fK2mPq", user); err != nil {
		t.Errorf("Expected no error for dissimilar password, got: %v", err)
	}

	// Nil user should not error
	if err := v.Validate("anything", nil); err != nil {
		t.Errorf("Expected no error with nil user, got: %v", err)
	}

	// Help text
	if v.GetHelpText() == "" {
		t.Errorf("Expected non-empty help text")
	}
}

func TestValidatePasswordCollectsAllErrors(t *testing.T) {
	// Save and restore original validators
	original := GetPasswordValidators()
	defer SetPasswordValidators(original)

	SetPasswordValidators([]PasswordValidator{
		NewMinimumLengthValidator(10),
		NewNumericPasswordValidator(),
	})

	// "123" is too short AND entirely numeric
	err := ValidatePassword("123", nil)
	if err == nil {
		t.Fatalf("Expected error for short numeric password")
	}
	errs := PasswordValidationErrors("123", nil)
	if len(errs) != 2 {
		t.Errorf("Expected 2 errors, got %d: %v", len(errs), errs)
	}
}

func TestValidatePasswordNoValidators(t *testing.T) {
	original := GetPasswordValidators()
	defer SetPasswordValidators(original)

	SetPasswordValidators(nil)
	if err := ValidatePassword("anything", nil); err != nil {
		t.Errorf("Expected no error with no validators, got: %v", err)
	}
}

func TestRegisterPasswordValidator(t *testing.T) {
	original := GetPasswordValidators()
	defer SetPasswordValidators(original)

	SetPasswordValidators(nil)

	// Custom validator
	customErr := "custom validation failed"
	RegisterPasswordValidator(&customValidator{errMsg: customErr})

	err := ValidatePassword("test", nil)
	if err == nil || err.Error() != customErr {
		t.Errorf("Expected custom error %q, got: %v", customErr, err)
	}
}

// customValidator is a test-only validator.
type customValidator struct {
	errMsg string
}

func (v *customValidator) Validate(password string, user User) error {
	return &testError{msg: v.errMsg}
}
func (v *customValidator) GetHelpText() string { return "custom help" }

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

func TestDefaultValidatorsRegistered(t *testing.T) {
	// The init() should have registered 3 default validators
	vs := GetPasswordValidators()
	if len(vs) < 3 {
		t.Errorf("Expected at least 3 default validators, got %d", len(vs))
	}
}
