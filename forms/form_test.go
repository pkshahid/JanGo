package forms

import (
	"fmt"
	"strings"
	"testing"
)

type ContactForm struct {
	Form
}

func NewContactForm() *ContactForm {
	fields := map[string]Field{
		"subject": &CharField{BaseField: BaseField{IsRequired: true}, MaxLength: 100},
		"message": &CharField{BaseField: BaseField{IsRequired: true}},
		"sender":  &EmailField{CharField: CharField{BaseField: BaseField{IsRequired: true}}},
		"cc":      &BooleanField{BaseField: BaseField{IsRequired: false}},
	}
	order := []string{"subject", "message", "sender", "cc"}

	f := &ContactForm{
		Form: *NewForm(fields, order),
	}
	// Cross-field validation via CleanFunc
	f.CleanFunc = func() error {
		if sender, ok := f.CleanedData["sender"].(string); ok {
			if subject, ok := f.CleanedData["subject"].(string); ok {
				if strings.Contains(subject, "spam") && !strings.Contains(sender, "admin") {
					return fmt.Errorf("Spam subject requires admin sender")
				}
			}
		}
		return nil
	}
	return f
}

func TestFormValidation(t *testing.T) {
	form := NewContactForm()

	// Unbound
	if form.IsValid() {
		t.Errorf("Unbound form shouldn't be valid")
	}

	// Invalid empty
	form.Bind(map[string]any{}, nil)
	if form.IsValid() {
		t.Errorf("Empty required form shouldn't be valid")
	}
	if len(form.Errors["subject"]) == 0 {
		t.Errorf("Expected subject required error")
	}

	// Valid
	validData := map[string]any{
		"subject": "Hello",
		"message": "This is a test message",
		"sender":  "test@example.com",
		"cc":      "on",
	}
	form.Bind(validData, nil)
	if !form.IsValid() {
		t.Errorf("Expected valid form, got errors: %v", form.Errors)
	}
	if form.CleanedData["cc"] != true {
		t.Errorf("Expected cc to be true, got %v", form.CleanedData["cc"])
	}

	// Cross-field validation fail
	spamData := map[string]any{
		"subject": "This is spam",
		"message": "Buy now",
		"sender":  "spammer@example.com",
	}
	form.Bind(spamData, nil)
	if form.IsValid() {
		t.Errorf("Expected invalid form due to cross-field validation")
	}
	if len(form.NonFieldErrors()) == 0 {
		t.Errorf("Expected non-field error")
	}
}

func TestFormRendering(t *testing.T) {
	form := NewContactForm()

	// Add error to test error rendering
	form.Bind(map[string]any{"subject": "hi"}, nil)
	form.IsValid()

	html := string(form.Render())

	// Verify error output
	if !strings.Contains(html, `<ul class="errorlist">`) {
		t.Errorf("Expected error list rendering")
	}

	// Verify input output
	if !strings.Contains(html, `<input type="text" name="subject" value="hi"`) {
		t.Errorf("Expected subject input with bound value, got: %s", html)
	}

	// Verify required *
	if !strings.Contains(html, `Subject *:`) {
		t.Errorf("Expected required indicator")
	}
}
