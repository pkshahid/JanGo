package auth

import (
	"testing"
)

// saveValidators saves and restores the global validator chain around a test.
func saveValidators(t *testing.T) {
	t.Helper()
	original := GetPasswordValidators()
	t.Cleanup(func() { SetPasswordValidators(original) })
}

func TestSetPasswordFormValidation(t *testing.T) {
	saveValidators(t)

	// Use only minimum length for predictable form tests
	SetPasswordValidators([]PasswordValidator{
		NewMinimumLengthValidator(8),
	})

	// --- Valid ---
	form := NewSetPasswordForm(nil)
	form.Bind(map[string]any{
		"new_password1": "secure-pass-123",
		"new_password2": "secure-pass-123",
	}, nil)
	if !form.IsValid() {
		t.Errorf("Expected valid form, got errors: %v", form.Errors)
	}

	// --- Mismatched passwords ---
	form2 := NewSetPasswordForm(nil)
	form2.Bind(map[string]any{
		"new_password1": "secure-pass-123",
		"new_password2": "different-pass",
	}, nil)
	if form2.IsValid() {
		t.Errorf("Expected invalid form for mismatched passwords")
	}
	if len(form2.NonFieldErrors()) == 0 {
		t.Errorf("Expected non-field error for mismatched passwords")
	}

	// --- Too short (validator failure) ---
	form3 := NewSetPasswordForm(nil)
	form3.Bind(map[string]any{
		"new_password1": "short",
		"new_password2": "short",
	}, nil)
	if form3.IsValid() {
		t.Errorf("Expected invalid form for short password")
	}

	// --- Missing fields ---
	form4 := NewSetPasswordForm(nil)
	form4.Bind(map[string]any{}, nil)
	if form4.IsValid() {
		t.Errorf("Expected invalid form for missing fields")
	}
}

func TestPasswordChangeFormValidation(t *testing.T) {
	saveValidators(t)
	SetPasswordValidators([]PasswordValidator{
		NewMinimumLengthValidator(8),
	})

	// PasswordChangeForm verifies old password via Authenticate,
	// which requires a DB. We test only the mismatch and validator
	// paths here (with nil user, old-password check is skipped).

	// --- Mismatched new passwords ---
	form := NewPasswordChangeForm(nil)
	form.Bind(map[string]any{
		"old_password":  "old-pass-123",
		"new_password1": "new-secure-pass",
		"new_password2": "different-pass",
	}, nil)
	if form.IsValid() {
		t.Errorf("Expected invalid form for mismatched new passwords")
	}

	// --- New password too short ---
	form2 := NewPasswordChangeForm(nil)
	form2.Bind(map[string]any{
		"old_password":  "old-pass-123",
		"new_password1": "short",
		"new_password2": "short",
	}, nil)
	if form2.IsValid() {
		t.Errorf("Expected invalid form for short new password")
	}

	// --- Missing old password ---
	form3 := NewPasswordChangeForm(nil)
	form3.Bind(map[string]any{
		"new_password1": "new-secure-pass",
		"new_password2": "new-secure-pass",
	}, nil)
	if form3.IsValid() {
		t.Errorf("Expected invalid form for missing old password")
	}
}

func TestPasswordResetFormValidation(t *testing.T) {
	// --- Valid email ---
	form := NewPasswordResetForm()
	form.Bind(map[string]any{
		"email": "user@example.com",
	}, nil)
	if !form.IsValid() {
		t.Errorf("Expected valid form, got errors: %v", form.Errors)
	}
	if form.GetEmail() != "user@example.com" {
		t.Errorf("Expected email 'user@example.com', got '%s'", form.GetEmail())
	}

	// --- Invalid email ---
	form2 := NewPasswordResetForm()
	form2.Bind(map[string]any{
		"email": "not-an-email",
	}, nil)
	if form2.IsValid() {
		t.Errorf("Expected invalid form for bad email")
	}

	// --- Missing email ---
	form3 := NewPasswordResetForm()
	form3.Bind(map[string]any{}, nil)
	if form3.IsValid() {
		t.Errorf("Expected invalid form for missing email")
	}
}

func TestUserCreationFormValidation(t *testing.T) {
	saveValidators(t)
	SetPasswordValidators([]PasswordValidator{
		NewMinimumLengthValidator(8),
	})

	existing := map[string]bool{"admin": true}

	// --- Valid ---
	form := NewUserCreationForm(existing)
	form.Bind(map[string]any{
		"username":  "newuser",
		"password1": "secure-pass-123",
		"password2": "secure-pass-123",
	}, nil)
	if !form.IsValid() {
		t.Errorf("Expected valid form, got errors: %v", form.Errors)
	}

	// --- Duplicate username ---
	form2 := NewUserCreationForm(existing)
	form2.Bind(map[string]any{
		"username":  "admin",
		"password1": "secure-pass-123",
		"password2": "secure-pass-123",
	}, nil)
	if form2.IsValid() {
		t.Errorf("Expected invalid form for duplicate username")
	}

	// --- Mismatched passwords ---
	form3 := NewUserCreationForm(existing)
	form3.Bind(map[string]any{
		"username":  "newuser2",
		"password1": "secure-pass-123",
		"password2": "different-pass",
	}, nil)
	if form3.IsValid() {
		t.Errorf("Expected invalid form for mismatched passwords")
	}

	// --- Password too short ---
	form4 := NewUserCreationForm(existing)
	form4.Bind(map[string]any{
		"username":  "newuser3",
		"password1": "short",
		"password2": "short",
	}, nil)
	if form4.IsValid() {
		t.Errorf("Expected invalid form for short password")
	}

	// --- Missing username ---
	form5 := NewUserCreationForm(existing)
	form5.Bind(map[string]any{
		"password1": "secure-pass-123",
		"password2": "secure-pass-123",
	}, nil)
	if form5.IsValid() {
		t.Errorf("Expected invalid form for missing username")
	}
}

func TestFormErrorsToString(t *testing.T) {
	saveValidators(t)
	SetPasswordValidators([]PasswordValidator{
		NewMinimumLengthValidator(8),
	})

	// Valid form → empty string
	form := NewSetPasswordForm(nil)
	form.Bind(map[string]any{
		"new_password1": "secure-pass-123",
		"new_password2": "secure-pass-123",
	}, nil)
	form.IsValid()
	if s := FormErrorsToString(&form.Form); s != "" {
		t.Errorf("Expected empty error string for valid form, got '%s'", s)
	}

	// Invalid form → non-empty string
	form2 := NewSetPasswordForm(nil)
	form2.Bind(map[string]any{
		"new_password1": "short",
		"new_password2": "short",
	}, nil)
	form2.IsValid()
	if s := FormErrorsToString(&form2.Form); s == "" {
		t.Errorf("Expected non-empty error string for invalid form")
	}
}
