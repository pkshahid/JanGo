package auth

import (
	"fmt"
	"strings"
	"unicode"
)

// PasswordValidator is the interface every password validator must implement.
type PasswordValidator interface {
	// Validate returns an error if the password fails validation.
	Validate(password string, user User) error
	// GetHelpText returns help text shown to users describing the requirement.
	GetHelpText() string
}

// ----------------------------------------------------------------------------
// Global registry
// ----------------------------------------------------------------------------

var passwordValidators []PasswordValidator

// RegisterPasswordValidator adds a validator to the global chain.
func RegisterPasswordValidator(v PasswordValidator) {
	passwordValidators = append(passwordValidators, v)
}

// SetPasswordValidators replaces the entire validator chain.
// Pass nil to clear all validators.
func SetPasswordValidators(validators []PasswordValidator) {
	passwordValidators = validators
}

// GetPasswordValidators returns the currently registered validators.
func GetPasswordValidators() []PasswordValidator {
	return passwordValidators
}

// ValidatePassword runs every registered validator against the password.
// It collects all errors and returns them joined with "; ".
// Returns nil if there are no errors (or no validators registered).
func ValidatePassword(password string, user User) error {
	var errs []string
	for _, v := range passwordValidators {
		if err := v.Validate(password, user); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("%s", strings.Join(errs, "; "))
}

// PasswordValidationErrors runs every registered validator and returns
// the individual error messages as a slice (useful for form field errors).
func PasswordValidationErrors(password string, user User) []string {
	var errs []string
	for _, v := range passwordValidators {
		if err := v.Validate(password, user); err != nil {
			errs = append(errs, err.Error())
		}
	}
	return errs
}

// ----------------------------------------------------------------------------
// MinimumLengthValidator
// ----------------------------------------------------------------------------

// MinimumLengthValidator rejects passwords shorter than MinLength.
type MinimumLengthValidator struct {
	MinLength int
}

func NewMinimumLengthValidator(minLength int) *MinimumLengthValidator {
	if minLength <= 0 {
		minLength = 8
	}
	return &MinimumLengthValidator{MinLength: minLength}
}

func (v *MinimumLengthValidator) Validate(password string, user User) error {
	if len(password) < v.MinLength {
		return fmt.Errorf("this password is too short. it must contain at least %d characters", v.MinLength)
	}
	return nil
}

func (v *MinimumLengthValidator) GetHelpText() string {
	return fmt.Sprintf("your password must contain at least %d characters", v.MinLength)
}

// ----------------------------------------------------------------------------
// CommonPasswordValidator
// ----------------------------------------------------------------------------

// commonPasswords is a small built-in list of frequently used passwords.
// In production this would be loaded from a file (e.g. Django's
// common-passwords.txt.gz with ~20k entries).
var commonPasswords = map[string]bool{
	"password":   true,
	"123456":     true,
	"12345678":   true,
	"123456789":  true,
	"1234567890": true,
	"qwerty":     true,
	"abc123":     true,
	"111111":     true,
	"12345":      true,
	"admin":      true,
	"letmein":    true,
	"welcome":    true,
	"monkey":     true,
	"dragon":     true,
	"master":     true,
	"sunshine":   true,
	"princess":   true,
	"football":   true,
	"shadow":     true,
	"superman":   true,
	"michael":    true,
	"ninja":      true,
	"mustang":    true,
	"password1":  true,
	"iloveyou":   true,
	"trustno1":   true,
	"000000":     true,
	"qazwsx":     true,
	"123qwe":     true,
	"changeme":   true,
}

// CommonPasswordValidator rejects passwords found in a common-password list.
type CommonPasswordValidator struct {
	// Passwords can be overridden for testing or custom lists.
	Passwords map[string]bool
}

func NewCommonPasswordValidator() *CommonPasswordValidator {
	return &CommonPasswordValidator{Passwords: commonPasswords}
}

func (v *CommonPasswordValidator) Validate(password string, user User) error {
	list := v.Passwords
	if list == nil {
		list = commonPasswords
	}
	if list[strings.ToLower(password)] {
		return fmt.Errorf("this password is too common")
	}
	return nil
}

func (v *CommonPasswordValidator) GetHelpText() string {
	return "your password can't be a commonly used password"
}

// ----------------------------------------------------------------------------
// NumericPasswordValidator
// ----------------------------------------------------------------------------

// NumericPasswordValidator rejects passwords that are entirely numeric.
type NumericPasswordValidator struct{}

func NewNumericPasswordValidator() *NumericPasswordValidator {
	return &NumericPasswordValidator{}
}

func (v *NumericPasswordValidator) Validate(password string, user User) error {
	if password == "" {
		return nil
	}
	for _, r := range password {
		if !unicode.IsDigit(r) {
			return nil
		}
	}
	return fmt.Errorf("this password is entirely numeric")
}

func (v *NumericPasswordValidator) GetHelpText() string {
	return "your password can't be entirely numeric"
}

// ----------------------------------------------------------------------------
// UserAttributeSimilarityValidator
// ----------------------------------------------------------------------------

// UserAttributeSimilarityValidator rejects passwords that are too similar
// to user attributes (username, email, first name, last name).
// Similarity is measured by the ratio of the longest common substring
// to the password length.
type UserAttributeSimilarityValidator struct {
	// MaxSimilarity is the threshold above which the password is rejected.
	// A value of 0.7 means if 70% or more of the password matches a user
	// attribute, it is rejected. Default is 0.7.
	MaxSimilarity float64
	// UserAttributes defines which attributes to check.
	// Defaults to ["username", "email", "first_name", "last_name"].
	UserAttributes []string
}

func NewUserAttributeSimilarityValidator() *UserAttributeSimilarityValidator {
	return &UserAttributeSimilarityValidator{
		MaxSimilarity:  0.7,
		UserAttributes: []string{"username", "email", "first_name", "last_name"},
	}
}

func (v *UserAttributeSimilarityValidator) Validate(password string, user User) error {
	if user == nil || password == "" {
		return nil
	}

	threshold := v.MaxSimilarity
	if threshold <= 0 {
		threshold = 0.7
	}

	attrs := v.UserAttributes
	if attrs == nil {
		attrs = []string{"username", "email", "first_name", "last_name"}
	}

	passwordLower := strings.ToLower(password)

	for _, attrName := range attrs {
		attrVal := getUserAttribute(user, attrName)
		if attrVal == "" {
			continue
		}
		attrLower := strings.ToLower(attrVal)
		similarity := longestCommonSubsequenceRatio(passwordLower, attrLower)
		if similarity >= threshold {
			return fmt.Errorf("the password is too similar to the %s", attrName)
		}
	}

	return nil
}

func (v *UserAttributeSimilarityValidator) GetHelpText() string {
	return "your password can't be too similar to your personal information"
}

// getUserAttribute retrieves a string attribute from the User interface
// by checking known methods. This avoids reflection for the common case.
func getUserAttribute(user User, attrName string) string {
	switch attrName {
	case "username":
		return user.Username()
	case "email":
		return user.Email()
	case "first_name":
		if u, ok := user.(*AbstractUser); ok {
			return u.FirstName
		}
	case "last_name":
		if u, ok := user.(*AbstractUser); ok {
			return u.LastName
		}
	}
	return ""
}

// longestCommonSubsequenceRatio computes the ratio of the length of the
// longest common subsequence to the shorter string's length.
// This is a simplified version of Django's SequenceMatcher.ratio().
func longestCommonSubsequenceRatio(a, b string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	shorter, longer := a, b
	if len(a) > len(b) {
		shorter, longer = b, a
	}

	// DP table for LCS
	prev := make([]int, len(shorter)+1)
	curr := make([]int, len(shorter)+1)

	for i := 1; i <= len(longer); i++ {
		for j := 1; j <= len(shorter); j++ {
			if longer[i-1] == shorter[j-1] {
				curr[j] = prev[j-1] + 1
			} else {
				if prev[j] > curr[j-1] {
					curr[j] = prev[j]
				} else {
					curr[j] = curr[j-1]
				}
			}
		}
		prev, curr = curr, prev
		for k := range curr {
			curr[k] = 0
		}
	}

	lcsLen := prev[len(shorter)]
	return float64(2*lcsLen) / float64(len(a)+len(b))
}

// ----------------------------------------------------------------------------
// Default validators
// ----------------------------------------------------------------------------

// RegisterDefaultPasswordValidators registers the standard set of validators
// (minimum length 8, common password check, numeric check).
// This is called automatically on package init but can be overridden
// via SetPasswordValidators.
func RegisterDefaultPasswordValidators() {
	SetPasswordValidators([]PasswordValidator{
		NewMinimumLengthValidator(8),
		NewCommonPasswordValidator(),
		NewNumericPasswordValidator(),
	})
}

func init() {
	RegisterDefaultPasswordValidators()
}
