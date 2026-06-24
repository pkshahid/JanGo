package validators

import (
	"testing"
)

func TestRequired(t *testing.T) {
	v := Required()

	if err := v("hello"); err != nil {
		t.Errorf("Expected no error for non-empty string, got: %v", err)
	}
	if err := v(""); err == nil {
		t.Error("Expected error for empty string")
	}
	if err := v("   "); err == nil {
		t.Error("Expected error for whitespace-only string")
	}
	if err := v(nil); err == nil {
		t.Error("Expected error for nil")
	}
	if err := v(42); err != nil {
		t.Errorf("Expected no error for non-string value, got: %v", err)
	}
}

func TestMinLength(t *testing.T) {
	v := MinLength(3)

	if err := v("abc"); err != nil {
		t.Errorf("Expected no error for 'abc', got: %v", err)
	}
	if err := v("ab"); err == nil {
		t.Error("Expected error for 'ab' (< 3 chars)")
	}
	if err := v("abcdef"); err != nil {
		t.Errorf("Expected no error for 'abcdef', got: %v", err)
	}
}

func TestMaxLength(t *testing.T) {
	v := MaxLength(5)

	if err := v("hello"); err != nil {
		t.Errorf("Expected no error for 'hello', got: %v", err)
	}
	if err := v("toolong"); err == nil {
		t.Error("Expected error for 'toolong' (> 5 chars)")
	}
	if err := v("hi"); err != nil {
		t.Errorf("Expected no error for 'hi', got: %v", err)
	}
}

func TestMinValue(t *testing.T) {
	v := MinValue(10)

	if err := v(15); err != nil {
		t.Errorf("Expected no error for 15, got: %v", err)
	}
	if err := v(10); err != nil {
		t.Errorf("Expected no error for 10 (equal), got: %v", err)
	}
	if err := v(5); err == nil {
		t.Error("Expected error for 5 (< 10)")
	}
	if err := v(15.5); err != nil {
		t.Errorf("Expected no error for 15.5, got: %v", err)
	}
	if err := v("not a number"); err == nil {
		t.Error("Expected error for non-numeric value")
	}
}

func TestMaxValue(t *testing.T) {
	v := MaxValue(100)

	if err := v(50); err != nil {
		t.Errorf("Expected no error for 50, got: %v", err)
	}
	if err := v(100); err != nil {
		t.Errorf("Expected no error for 100 (equal), got: %v", err)
	}
	if err := v(150); err == nil {
		t.Error("Expected error for 150 (> 100)")
	}
}

func TestEmailValidator(t *testing.T) {
	v := EmailValidator()

	valid := []string{"user@example.com", "test.name@domain.org", "a+b@c.co"}
	for _, email := range valid {
		if err := v(email); err != nil {
			t.Errorf("Expected %q to be valid, got: %v", email, err)
		}
	}

	invalid := []string{"notanemail", "@missing.user", "user@", "user@.com", ""}
	for _, email := range invalid {
		if err := v(email); err == nil {
			t.Errorf("Expected %q to be invalid", email)
		}
	}
}

func TestURLValidator(t *testing.T) {
	v := URLValidator()

	valid := []string{"http://example.com", "https://example.com/path", "https://sub.domain.com:8080/path?q=1"}
	for _, u := range valid {
		if err := v(u); err != nil {
			t.Errorf("Expected %q to be valid, got: %v", u, err)
		}
	}

	invalid := []string{"not-a-url", "ftp://example.com", "example.com"}
	for _, u := range invalid {
		if err := v(u); err == nil {
			t.Errorf("Expected %q to be invalid", u)
		}
	}

	// Test custom schemes
	v2 := URLValidator("ftp", "https")
	if err := v2("ftp://files.example.com"); err != nil {
		t.Errorf("Expected ftp to be valid with custom schemes, got: %v", err)
	}
}

func TestRegexValidator(t *testing.T) {
	v := RegexValidator(`^\d{3}-\d{4}$`, "Enter a valid phone number (XXX-XXXX).")

	if err := v("123-4567"); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if err := v("12-4567"); err == nil {
		t.Error("Expected error for invalid format")
	}
}

func TestSlugValidator(t *testing.T) {
	v := SlugValidator()

	valid := []string{"hello-world", "test_slug", "CamelCase", "123"}
	for _, s := range valid {
		if err := v(s); err != nil {
			t.Errorf("Expected %q to be a valid slug, got: %v", s, err)
		}
	}

	invalid := []string{"hello world", "test/slug", "special@char"}
	for _, s := range invalid {
		if err := v(s); err == nil {
			t.Errorf("Expected %q to be an invalid slug", s)
		}
	}
}

func TestIPAddressValidator(t *testing.T) {
	v := IPAddressValidator()

	valid := []string{"192.168.1.1", "10.0.0.1", "::1", "2001:db8::1"}
	for _, ip := range valid {
		if err := v(ip); err != nil {
			t.Errorf("Expected %q to be a valid IP, got: %v", ip, err)
		}
	}

	invalid := []string{"not.an.ip", "256.256.256.256", "abc"}
	for _, ip := range invalid {
		if err := v(ip); err == nil {
			t.Errorf("Expected %q to be an invalid IP", ip)
		}
	}
}

func TestIntegerValidator(t *testing.T) {
	v := IntegerValidator()

	valid := []string{"123", "-456", "0"}
	for _, s := range valid {
		if err := v(s); err != nil {
			t.Errorf("Expected %q to be a valid integer, got: %v", s, err)
		}
	}

	invalid := []string{"12.5", "abc", "12a"}
	for _, s := range invalid {
		if err := v(s); err == nil {
			t.Errorf("Expected %q to be an invalid integer", s)
		}
	}
}

func TestDecimalValidator(t *testing.T) {
	v := DecimalValidator(5, 2)

	if err := v("123.45"); err != nil {
		t.Errorf("Expected no error for '123.45', got: %v", err)
	}
	if err := v("1.1"); err != nil {
		t.Errorf("Expected no error for '1.1', got: %v", err)
	}
	if err := v("123.456"); err == nil {
		t.Error("Expected error for too many decimal places")
	}
	if err := v("123456"); err == nil {
		t.Error("Expected error for too many digits")
	}
}

func TestFileExtensionValidator(t *testing.T) {
	v := FileExtensionValidator([]string{"jpg", "png", "gif"})

	if err := v("photo.jpg"); err != nil {
		t.Errorf("Expected no error for jpg, got: %v", err)
	}
	if err := v("image.PNG"); err != nil {
		t.Errorf("Expected no error for PNG (case insensitive), got: %v", err)
	}
	if err := v("doc.pdf"); err == nil {
		t.Error("Expected error for pdf extension")
	}
}

func TestProhibitNullCharacters(t *testing.T) {
	v := ProhibitNullCharacters()

	if err := v("normal text"); err != nil {
		t.Errorf("Expected no error for normal text, got: %v", err)
	}
	if err := v("has\x00null"); err == nil {
		t.Error("Expected error for string with null character")
	}
}

func TestCompose(t *testing.T) {
	v := Compose(
		Required(),
		MinLength(3),
		MaxLength(10),
	)

	if err := v("hello"); err != nil {
		t.Errorf("Expected no error for 'hello', got: %v", err)
	}
	if err := v(""); err == nil {
		t.Error("Expected error for empty (required)")
	}
	if err := v("ab"); err == nil {
		t.Error("Expected error for 'ab' (min_length)")
	}
	if err := v("this is too long string"); err == nil {
		t.Error("Expected error for too-long string (max_length)")
	}
}

func TestValidationErrorInterface(t *testing.T) {
	err := &ValidationError{
		Message: "test error",
		Code:    "test",
		Params:  map[string]any{"key": "value"},
	}

	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %q", err.Error())
	}
}
