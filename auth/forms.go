package auth

import (
	"fmt"

	"github.com/pkshahid/JanGo/forms"
)

// ----------------------------------------------------------------------------
// AuthenticationForm
// ----------------------------------------------------------------------------

// AuthenticationForm is the login form. It validates that the username
// and password fields are present, then authenticates the user.
type AuthenticationForm struct {
	forms.Form
}

// NewAuthenticationForm returns an unbound AuthenticationForm.
func NewAuthenticationForm() *AuthenticationForm {
	fields := map[string]forms.Field{
		"username": &forms.CharField{
			BaseField: forms.BaseField{IsRequired: true, LabelStr: "Username", WidgetField: forms.NewTextInput(nil)},
			MaxLength: 150,
			Strip:     false,
		},
		"password": &forms.CharField{
			BaseField: forms.BaseField{IsRequired: true, LabelStr: "Password", WidgetField: forms.NewPasswordInput(nil)},
			MaxLength: 128,
			Strip:     false,
		},
	}
	order := []string{"username", "password"}

	f := &AuthenticationForm{
		Form: *forms.NewForm(fields, order),
	}
	f.CleanFunc = func() error {
		username, _ := f.CleanedData["username"].(string)
		password, _ := f.CleanedData["password"].(string)

		user, err := Authenticate(username, password)
		if err != nil || user == nil {
			return fmt.Errorf("please enter a correct username and password. note that both fields may be case-sensitive")
		}

		f.CleanedData["user"] = user
		return nil
	}
	return f
}

// GetUser returns the authenticated user after Clean() succeeds.
func (f *AuthenticationForm) GetUser() User {
	if user, ok := f.CleanedData["user"].(User); ok {
		return user
	}
	return nil
}

// ----------------------------------------------------------------------------
// SetPasswordForm
// ----------------------------------------------------------------------------

// SetPasswordForm is used when setting a new password without verifying
// the old one (e.g. password reset confirmation, admin user creation).
// It runs all registered password validators against the new password.
type SetPasswordForm struct {
	forms.Form
	// User is the user whose password is being set.
	// May be nil for password-reset flows where the user is resolved from a token.
	User User
}

// NewSetPasswordForm returns an unbound SetPasswordForm for the given user.
func NewSetPasswordForm(user User) *SetPasswordForm {
	fields := map[string]forms.Field{
		"new_password1": &forms.CharField{
			BaseField: forms.BaseField{IsRequired: true, LabelStr: "New password", WidgetField: forms.NewPasswordInput(nil)},
			MaxLength: 128,
			Strip:     false,
		},
		"new_password2": &forms.CharField{
			BaseField: forms.BaseField{IsRequired: true, LabelStr: "New password confirmation", WidgetField: forms.NewPasswordInput(nil)},
			MaxLength: 128,
			Strip:     false,
		},
	}
	order := []string{"new_password1", "new_password2"}

	f := &SetPasswordForm{
		Form: *forms.NewForm(fields, order),
		User: user,
	}
	f.CleanFunc = func() error {
		pw1, _ := f.CleanedData["new_password1"].(string)
		pw2, _ := f.CleanedData["new_password2"].(string)

		if pw1 != pw2 {
			return fmt.Errorf("the two password fields didn't match")
		}

		if err := ValidatePassword(pw1, f.User); err != nil {
			return err
		}

		return nil
	}
	return f
}

// ----------------------------------------------------------------------------
// PasswordChangeForm
// ----------------------------------------------------------------------------

// PasswordChangeForm requires the user's old password in addition to the
// new password (with confirmation). It verifies the old password via
// Authenticate, then runs validators on the new password.
type PasswordChangeForm struct {
	forms.Form
	// User is the user changing their password.
	User User
}

// NewPasswordChangeForm returns an unbound PasswordChangeForm for the given user.
func NewPasswordChangeForm(user User) *PasswordChangeForm {
	fields := map[string]forms.Field{
		"old_password": &forms.CharField{
			BaseField: forms.BaseField{IsRequired: true, LabelStr: "Old password", WidgetField: forms.NewPasswordInput(nil)},
			MaxLength: 128,
			Strip:     false,
		},
		"new_password1": &forms.CharField{
			BaseField: forms.BaseField{IsRequired: true, LabelStr: "New password", WidgetField: forms.NewPasswordInput(nil)},
			MaxLength: 128,
			Strip:     false,
		},
		"new_password2": &forms.CharField{
			BaseField: forms.BaseField{IsRequired: true, LabelStr: "New password confirmation", WidgetField: forms.NewPasswordInput(nil)},
			MaxLength: 128,
			Strip:     false,
		},
	}
	order := []string{"old_password", "new_password1", "new_password2"}

	f := &PasswordChangeForm{
		Form: *forms.NewForm(fields, order),
		User: user,
	}
	f.CleanFunc = func() error {
		oldPw, _ := f.CleanedData["old_password"].(string)
		newPw1, _ := f.CleanedData["new_password1"].(string)
		newPw2, _ := f.CleanedData["new_password2"].(string)

		// 1. Verify old password
		if f.User != nil {
			_, err := Authenticate(f.User.Username(), oldPw)
			if err != nil {
				return fmt.Errorf("your old password was entered incorrectly. please enter it again")
			}
		}

		// 2. Verify new passwords match
		if newPw1 != newPw2 {
			return fmt.Errorf("the two password fields didn't match")
		}

		// 3. Run password validators
		if err := ValidatePassword(newPw1, f.User); err != nil {
			return err
		}

		return nil
	}
	return f
}

// ----------------------------------------------------------------------------
// PasswordResetForm
// ----------------------------------------------------------------------------

// PasswordResetForm accepts an email address and validates that a user
// with that email exists (when an email backend is available).
type PasswordResetForm struct {
	forms.Form
}

// NewPasswordResetForm returns an unbound PasswordResetForm.
func NewPasswordResetForm() *PasswordResetForm {
	fields := map[string]forms.Field{
		"email": &forms.EmailField{
			CharField: forms.CharField{
				BaseField: forms.BaseField{IsRequired: true, LabelStr: "Email", WidgetField: forms.NewEmailInput(nil)},
				MaxLength: 254,
			},
		},
	}
	order := []string{"email"}

	f := &PasswordResetForm{
		Form: *forms.NewForm(fields, order),
	}
	return f
}

// GetEmail returns the cleaned email after validation.
func (f *PasswordResetForm) GetEmail() string {
	if email, ok := f.CleanedData["email"].(string); ok {
		return email
	}
	return ""
}

// ----------------------------------------------------------------------------
// UserCreationForm
// ----------------------------------------------------------------------------

// UserCreationForm is used to create a new user with a username and
// password (with confirmation). It runs password validators and checks
// that the username is not already taken.
type UserCreationForm struct {
	forms.Form
	// ExistingUsernames is a set of usernames already in use.
	// In a full implementation this would be checked via the ORM.
	ExistingUsernames map[string]bool
}

// NewUserCreationForm returns an unbound UserCreationForm.
// Pass a map of existing usernames for uniqueness checking, or nil
// to skip the uniqueness check (e.g. in tests).
func NewUserCreationForm(existing map[string]bool) *UserCreationForm {
	fields := map[string]forms.Field{
		"username": &forms.CharField{
			BaseField: forms.BaseField{IsRequired: true, LabelStr: "Username", WidgetField: forms.NewTextInput(nil)},
			MaxLength: 150,
			Strip:     false,
		},
		"password1": &forms.CharField{
			BaseField: forms.BaseField{IsRequired: true, LabelStr: "Password", WidgetField: forms.NewPasswordInput(nil)},
			MaxLength: 128,
			Strip:     false,
		},
		"password2": &forms.CharField{
			BaseField: forms.BaseField{IsRequired: true, LabelStr: "Password confirmation", WidgetField: forms.NewPasswordInput(nil)},
			MaxLength: 128,
			Strip:     false,
		},
	}
	order := []string{"username", "password1", "password2"}

	f := &UserCreationForm{
		Form:              *forms.NewForm(fields, order),
		ExistingUsernames: existing,
	}
	f.CleanFunc = func() error {
		username, _ := f.CleanedData["username"].(string)
		pw1, _ := f.CleanedData["password1"].(string)
		pw2, _ := f.CleanedData["password2"].(string)

		// 1. Check username uniqueness
		if f.ExistingUsernames != nil && f.ExistingUsernames[username] {
			return fmt.Errorf("a user with that username already exists")
		}

		// 2. Verify passwords match
		if pw1 != pw2 {
			return fmt.Errorf("the two password fields didn't match")
		}

		// 3. Run password validators (user is nil during creation)
		if err := ValidatePassword(pw1, nil); err != nil {
			return err
		}

		return nil
	}
	return f
}

// ----------------------------------------------------------------------------
// Helper: convert form errors to a simple string for template rendering
// ----------------------------------------------------------------------------

// FormErrorsToString joins all form errors into a single string suitable
// for display in template contexts. This bridges the gap between the
// forms.Form error map and the simple string error format used by the
// existing auth views.
func FormErrorsToString(form *forms.Form) string {
	if form.Errors == nil || len(form.Errors) == 0 {
		return ""
	}

	var parts []string
	// Non-field errors first
	for _, e := range form.NonFieldErrors() {
		parts = append(parts, e)
	}
	// Then field-specific errors
	for _, fieldName := range form.Order {
		if errs, ok := form.Errors[fieldName]; ok {
			for _, e := range errs {
				parts = append(parts, e)
			}
		}
	}

	if len(parts) == 0 {
		return ""
	}

	result := parts[0]
	for _, p := range parts[1:] {
		result += " " + p
	}
	return result
}
