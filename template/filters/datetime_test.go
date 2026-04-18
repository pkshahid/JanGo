package filters

import (
	"testing"
	"time"
	godjango "github.com/pkshahid/JanGo/template"
)

func TestDateTimeFilters(t *testing.T) {
	lib := godjango.NewLibrary()
	RegisterDateTimeFilters(lib)

	date := time.Date(2023, 10, 15, 14, 30, 0, 0, time.UTC)

	f := lib.Filters["date"]
	if res, _ := f(date, "Y-m-d"); res != "2023-10-15" {
		t.Errorf("Expected 2023-10-15, got %v", res)
	}

	f = lib.Filters["time"]
	if res, _ := f(date, "H:i"); res != "14:30" {
		t.Errorf("Expected 14:30, got %v", res)
	}

	f = lib.Filters["filesizeformat"]
	if res, _ := f(1500, ""); res != "1.5 KB" {
		t.Errorf("Expected 1.5 KB, got %v", res)
	}

	f = lib.Filters["pluralize"]
	if res, _ := f(1, ""); res != "" {
		t.Errorf("Expected empty string, got '%v'", res)
	}
	if res, _ := f(2, ""); res != "s" {
		t.Errorf("Expected 's', got '%v'", res)
	}
	if res, _ := f(2, "y,ies"); res != "ies" {
		t.Errorf("Expected 'ies', got '%v'", res)
	}
	if res, _ := f(1, "y,ies"); res != "y" {
		t.Errorf("Expected 'y', got '%v'", res)
	}
}
