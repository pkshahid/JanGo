// Package validators provides Django-style reusable validation functions.
// Validators can be used with forms, models, and serializers.
package validators

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ValidationError represents a validation failure.
type ValidationError struct {
	Message string
	Code    string
	Params  map[string]any
}

func (e *ValidationError) Error() string {
	return e.Message
}

// Validator is a function that validates a value and returns an error if invalid.
type Validator func(value any) *ValidationError

// Compose combines multiple validators into one.
func Compose(validators ...Validator) Validator {
	return func(value any) *ValidationError {
		for _, v := range validators {
			if err := v(value); err != nil {
				return err
			}
		}
		return nil
	}
}

// Required validates that a value is not empty.
func Required() Validator {
	return func(value any) *ValidationError {
		if value == nil {
			return &ValidationError{Message: "This field is required.", Code: "required"}
		}
		if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
			return &ValidationError{Message: "This field is required.", Code: "required"}
		}
		return nil
	}
}

// MinLength validates minimum string length.
func MinLength(min int) Validator {
	return func(value any) *ValidationError {
		s := fmt.Sprint(value)
		if utf8.RuneCountInString(s) < min {
			return &ValidationError{
				Message: fmt.Sprintf("Ensure this value has at least %d characters (it has %d).", min, utf8.RuneCountInString(s)),
				Code:    "min_length",
				Params:  map[string]any{"min_length": min, "show_value": utf8.RuneCountInString(s)},
			}
		}
		return nil
	}
}

// MaxLength validates maximum string length.
func MaxLength(max int) Validator {
	return func(value any) *ValidationError {
		s := fmt.Sprint(value)
		if utf8.RuneCountInString(s) > max {
			return &ValidationError{
				Message: fmt.Sprintf("Ensure this value has at most %d characters (it has %d).", max, utf8.RuneCountInString(s)),
				Code:    "max_length",
				Params:  map[string]any{"max_length": max, "show_value": utf8.RuneCountInString(s)},
			}
		}
		return nil
	}
}

// MinValue validates minimum numeric value.
func MinValue(min float64) Validator {
	return func(value any) *ValidationError {
		var v float64
		switch n := value.(type) {
		case int:
			v = float64(n)
		case int64:
			v = float64(n)
		case float64:
			v = n
		case float32:
			v = float64(n)
		default:
			return &ValidationError{Message: "A valid number is required.", Code: "invalid"}
		}
		if v < min {
			return &ValidationError{
				Message: fmt.Sprintf("Ensure this value is greater than or equal to %g.", min),
				Code:    "min_value",
				Params:  map[string]any{"min_value": min, "show_value": v},
			}
		}
		return nil
	}
}

// MaxValue validates maximum numeric value.
func MaxValue(max float64) Validator {
	return func(value any) *ValidationError {
		var v float64
		switch n := value.(type) {
		case int:
			v = float64(n)
		case int64:
			v = float64(n)
		case float64:
			v = n
		case float32:
			v = float64(n)
		default:
			return &ValidationError{Message: "A valid number is required.", Code: "invalid"}
		}
		if v > max {
			return &ValidationError{
				Message: fmt.Sprintf("Ensure this value is less than or equal to %g.", max),
				Code:    "max_value",
				Params:  map[string]any{"max_value": max, "show_value": v},
			}
		}
		return nil
	}
}

// EmailValidator validates email addresses.
func EmailValidator() Validator {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return func(value any) *ValidationError {
		s := fmt.Sprint(value)
		if !re.MatchString(s) {
			return &ValidationError{Message: "Enter a valid email address.", Code: "invalid"}
		}
		return nil
	}
}

// URLValidator validates URLs.
func URLValidator(schemes ...string) Validator {
	if len(schemes) == 0 {
		schemes = []string{"http", "https"}
	}
	return func(value any) *ValidationError {
		s := fmt.Sprint(value)
		u, err := url.Parse(s)
		if err != nil || u.Host == "" {
			return &ValidationError{Message: "Enter a valid URL.", Code: "invalid"}
		}
		schemeValid := false
		for _, scheme := range schemes {
			if u.Scheme == scheme {
				schemeValid = true
				break
			}
		}
		if !schemeValid {
			return &ValidationError{
				Message: fmt.Sprintf("Enter a valid URL with scheme: %s.", strings.Join(schemes, ", ")),
				Code:    "invalid",
			}
		}
		return nil
	}
}

// RegexValidator validates against a regular expression.
func RegexValidator(pattern string, message string) Validator {
	re := regexp.MustCompile(pattern)
	if message == "" {
		message = "Enter a valid value."
	}
	return func(value any) *ValidationError {
		s := fmt.Sprint(value)
		if !re.MatchString(s) {
			return &ValidationError{Message: message, Code: "invalid"}
		}
		return nil
	}
}

// SlugValidator validates that a string contains only letters, numbers, hyphens, and underscores.
func SlugValidator() Validator {
	return RegexValidator(`^[-a-zA-Z0-9_]+$`, "Enter a valid 'slug' consisting of letters, numbers, underscores or hyphens.")
}

// IPAddressValidator validates IPv4 and IPv6 addresses.
func IPAddressValidator() Validator {
	return func(value any) *ValidationError {
		s := fmt.Sprint(value)
		if net.ParseIP(s) == nil {
			return &ValidationError{Message: "Enter a valid IP address.", Code: "invalid"}
		}
		return nil
	}
}

// IntegerValidator validates that a value is a valid integer string.
func IntegerValidator() Validator {
	re := regexp.MustCompile(`^-?\d+$`)
	return func(value any) *ValidationError {
		s := fmt.Sprint(value)
		if !re.MatchString(s) {
			return &ValidationError{Message: "Enter a valid integer.", Code: "invalid"}
		}
		return nil
	}
}

// DecimalValidator validates decimal numbers with max digits and decimal places.
func DecimalValidator(maxDigits, decimalPlaces int) Validator {
	return func(value any) *ValidationError {
		s := fmt.Sprint(value)
		s = strings.TrimLeft(s, "-")
		parts := strings.SplitN(s, ".", 2)

		intPart := parts[0]
		decPart := ""
		if len(parts) == 2 {
			decPart = parts[1]
		}

		totalDigits := len(intPart) + len(decPart)
		if totalDigits > maxDigits {
			return &ValidationError{
				Message: fmt.Sprintf("Ensure that there are no more than %d digits in total.", maxDigits),
				Code:    "max_digits",
			}
		}
		if len(decPart) > decimalPlaces {
			return &ValidationError{
				Message: fmt.Sprintf("Ensure that there are no more than %d decimal places.", decimalPlaces),
				Code:    "max_decimal_places",
			}
		}
		return nil
	}
}

// FileExtensionValidator validates file extensions.
func FileExtensionValidator(allowedExtensions []string) Validator {
	return func(value any) *ValidationError {
		s := fmt.Sprint(value)
		ext := strings.ToLower(strings.TrimPrefix(getFileExt(s), "."))
		for _, allowed := range allowedExtensions {
			if ext == strings.ToLower(allowed) {
				return nil
			}
		}
		return &ValidationError{
			Message: fmt.Sprintf("File extension %q is not allowed. Allowed extensions: %s.", ext, strings.Join(allowedExtensions, ", ")),
			Code:    "invalid_extension",
		}
	}
}

func getFileExt(name string) string {
	idx := strings.LastIndex(name, ".")
	if idx == -1 {
		return ""
	}
	return name[idx:]
}

// ProhibitNullCharacters validates that a string doesn't contain null bytes.
func ProhibitNullCharacters() Validator {
	return func(value any) *ValidationError {
		s := fmt.Sprint(value)
		if strings.ContainsRune(s, '\x00') {
			return &ValidationError{Message: "Null characters are not allowed.", Code: "null_characters_not_allowed"}
		}
		return nil
	}
}
