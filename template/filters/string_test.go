package filters

import (
	"testing"
	godjango "github.com/godjango/godjango/template"
)

func TestStringHtmlFilters(t *testing.T) {
	lib := godjango.NewLibrary()
	RegisterStringHtmlFilters(lib)

	f := lib.Filters["addslashes"]
	if res, _ := f("I'm testing \"slashes\"", ""); res != "I\\'m testing \\\"slashes\\\"" {
		t.Errorf("Expected addslashes, got %v", res)
	}

	f = lib.Filters["capfirst"]
	if res, _ := f("hello world", ""); res != "Hello world" {
		t.Errorf("Expected capfirst, got %v", res)
	}

	f = lib.Filters["center"]
	if res, _ := f("test", "10"); res != "   test   " {
		t.Errorf("Expected center, got %v", res)
	}

	f = lib.Filters["cut"]
	if res, _ := f("String with spaces", " "); res != "Stringwithspaces" {
		t.Errorf("Expected cut, got %v", res)
	}

	f = lib.Filters["ljust"]
	if res, _ := f("test", "10"); res != "test      " {
		t.Errorf("Expected ljust, got '%v'", res)
	}

	f = lib.Filters["lower"]
	if res, _ := f("TEST", ""); res != "test" {
		t.Errorf("Expected lower, got %v", res)
	}

	f = lib.Filters["striptags"]
	if res, _ := f("<b>test</b>", ""); res != "test" {
		t.Errorf("Expected striptags, got %v", res)
	}

	f = lib.Filters["linebreaks"]
	if res, _ := f("line1\nline2\n\nline3", ""); string(res.(godjango.SafeString)) != "<p>line1<br>line2</p>\n\n<p>line3</p>" {
		t.Errorf("Expected linebreaks, got %v", res)
	}

	f = lib.Filters["slugify"]
	if res, _ := f("This Is A Test!", ""); res != "this-is-a-test" {
		t.Errorf("Expected slugify, got %v", res)
	}

	f = lib.Filters["truncatechars"]
	if res, _ := f("Long string here", "8"); res != "Long ..." {
		t.Errorf("Expected truncatechars, got %v", res)
	}

	f = lib.Filters["wordcount"]
	if res, _ := f("one two three", ""); res != 3 {
		t.Errorf("Expected wordcount 3, got %v", res)
	}

	f = lib.Filters["wordwrap"]
	if res, _ := f("this is a long sentence", "8"); res != "this is\na long\nsentence" {
		t.Errorf("Expected wordwrap, got %v", res)
	}

	f = lib.Filters["phone2numeric"]
	if res, _ := f("1-800-COLLECT", ""); res != "1-800-2655328" {
		t.Errorf("Expected phone2numeric, got %v", res)
	}
}
