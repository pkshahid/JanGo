package forms

import (
	"strings"
	"testing"
	"time"
)

func TestFields(t *testing.T) {
	// CharField
	cf := &CharField{BaseField: BaseField{IsRequired: true}, MaxLength: 10, MinLength: 3, Strip: true}

	_, err := cf.Clean("  ab  ")
	if err == nil { t.Errorf("Expected MinLength error") }

	_, err = cf.Clean("this is too long")
	if err == nil { t.Errorf("Expected MaxLength error") }

	cleanCf, _ := cf.Clean("  test  ")
	if cleanCf != "test" { t.Errorf("Expected striped 'test', got '%s'", cleanCf) }

	_, err = cf.Clean("")
	if err == nil { t.Errorf("Expected required error") }

	// IntegerField
	intf := &IntegerField{BaseField: BaseField{IsRequired: true}}
	_, err = intf.Clean("abc")
	if err == nil { t.Errorf("Expected valid integer error") }

	cleanInt, _ := intf.Clean(" 42 ")
	if cleanInt != 42 { t.Errorf("Expected 42, got %v", cleanInt) }

	// FloatField
	ff := &FloatField{BaseField: BaseField{IsRequired: true}}
	cleanFf, _ := ff.Clean(" 3.14 ")
	if cleanFf != 3.14 { t.Errorf("Expected 3.14, got %v", cleanFf) }

	// BooleanField
	bf := &BooleanField{BaseField: BaseField{IsRequired: true}}
	cleanBf, _ := bf.Clean("on")
	if cleanBf != true { t.Errorf("Expected true, got %v", cleanBf) }

	_, err = bf.Clean("off")
	if err == nil { t.Errorf("Expected required error on false/off") } // Required boolean must be true

	// EmailField
	ef := &EmailField{CharField{BaseField: BaseField{IsRequired: true}}}
	_, err = ef.Clean("not-an-email")
	if err == nil { t.Errorf("Expected email validation error") }

	// SlugField
	sf := &SlugField{CharField{BaseField: BaseField{IsRequired: true}}}
	_, err = sf.Clean("invalid slug!")
	if err == nil { t.Errorf("Expected slug validation error") }

	cleanSf, _ := sf.Clean("valid-slug_123")
	if cleanSf != "valid-slug_123" { t.Errorf("Expected valid-slug_123, got %s", cleanSf) }

	// URLField
	uf := &URLField{CharField{BaseField: BaseField{IsRequired: true}}}
	cleanUf, _ := uf.Clean("example.com")
	if cleanUf != "http://example.com" { t.Errorf("Expected http://example.com, got %s", cleanUf) }

	// DateField
	df := &DateField{BaseField: BaseField{IsRequired: true}}
	cleanDf, _ := df.Clean("2023-10-15")
	tDate, _ := cleanDf.(time.Time)
	if tDate.Year() != 2023 || tDate.Month() != 10 || tDate.Day() != 15 {
		t.Errorf("Expected 2023-10-15, got %v", tDate)
	}

	// ChoiceField
	choices := []Choice{{Value: "A", Label: "A"}, {Value: "B", Label: "B"}}
	chf := &ChoiceField{BaseField: BaseField{IsRequired: true}, Choices: choices}
	cleanChf, _ := chf.Clean("A")
	if cleanChf != "A" { t.Errorf("Expected A, got %s", cleanChf) }

	_, err = chf.Clean("C")
	if err == nil { t.Errorf("Expected invalid choice error") }

	// MultipleChoiceField
	mcf := &MultipleChoiceField{BaseField: BaseField{IsRequired: true}, Choices: choices}
	cleanMcf, _ := mcf.Clean([]string{"A", "B"})
	if strings.Join(cleanMcf, ",") != "A,B" { t.Errorf("Expected A,B") }

	_, err = mcf.Clean([]string{"A", "C"})
	if err == nil { t.Errorf("Expected invalid choice error for C") }
}
