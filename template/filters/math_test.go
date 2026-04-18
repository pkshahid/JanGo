package filters

import (
	"testing"
	godjango "github.com/pkshahid/JanGo/template"
)

func TestMathLogicFilters(t *testing.T) {
	lib := godjango.NewLibrary()
	RegisterMathLogicFilters(lib)

	// add
	f := lib.Filters["add"]
	if res, _ := f(2, "3"); res != "5" { t.Errorf("Expected 5, got %v", res) }
	if res, _ := f("foo", "bar"); res != "foobar" { t.Errorf("Expected foobar, got %v", res) }

	// divisibleby
	f = lib.Filters["divisibleby"]
	if res, _ := f(10, "5"); res != true { t.Errorf("Expected true") }
	if res, _ := f(10, "3"); res != false { t.Errorf("Expected false") }

	// floatformat
	f = lib.Filters["floatformat"]
	if res, _ := f(34.23234, "2"); res != "34.23" { t.Errorf("Expected 34.23, got %v", res) }
	if res, _ := f(34.0, "-2"); res != "34" { t.Errorf("Expected 34, got %v", res) }
	if res, _ := f(34.23234, ""); res != "34.2" { t.Errorf("Expected 34.2, got %v", res) }
	if res, _ := f(34.0, ""); res != "34" { t.Errorf("Expected 34, got %v", res) }

	// get_digit
	f = lib.Filters["get_digit"]
	if res, _ := f(12345, "2"); res != "4" { t.Errorf("Expected 4, got %v", res) }
	if res, _ := f(12345, "1"); res != "5" { t.Errorf("Expected 5, got %v", res) }

	// default
	f = lib.Filters["default"]
	if res, _ := f("", "def"); res != "def" { t.Errorf("Expected def, got %v", res) }
	if res, _ := f(nil, "def"); res != "def" { t.Errorf("Expected def, got %v", res) }
	if res, _ := f("val", "def"); res != "val" { t.Errorf("Expected val, got %v", res) }

	// default_if_none
	f = lib.Filters["default_if_none"]
	if res, _ := f(nil, "def"); res != "def" { t.Errorf("Expected def, got %v", res) }
	if res, _ := f("", "def"); res != "" { t.Errorf("Expected '', got %v", res) }

	// yesno
	f = lib.Filters["yesno"]
	if res, _ := f(true, "y,n,m"); res != "y" { t.Errorf("Expected y, got %v", res) }
	if res, _ := f(false, "y,n,m"); res != "n" { t.Errorf("Expected n, got %v", res) }
	if res, _ := f(nil, "y,n,m"); res != "m" { t.Errorf("Expected m, got %v", res) }
}
