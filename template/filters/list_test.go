package filters

import (
	"reflect"
	"testing"
	godjango "github.com/godjango/godjango/template"
)

func TestListSeqFilters(t *testing.T) {
	lib := godjango.NewLibrary()
	RegisterListSeqFilters(lib)

	f := lib.Filters["first"]
	if res, _ := f([]string{"a", "b", "c"}, ""); res != "a" {
		t.Errorf("Expected first, got %v", res)
	}

	f = lib.Filters["last"]
	if res, _ := f([]string{"a", "b", "c"}, ""); res != "c" {
		t.Errorf("Expected last, got %v", res)
	}

	f = lib.Filters["join"]
	if res, _ := f([]string{"a", "b"}, ", "); res != "a, b" {
		t.Errorf("Expected join, got %v", res)
	}

	f = lib.Filters["length"]
	if res, _ := f([]string{"a", "b"}, ""); res != 2 {
		t.Errorf("Expected length 2, got %v", res)
	}
	if res, _ := f("test", ""); res != 4 {
		t.Errorf("Expected string length 4, got %v", res)
	}

	f = lib.Filters["slice"]
	resSlice, _ := f([]string{"a", "b", "c", "d"}, "1:3")
	v := reflect.ValueOf(resSlice)
	if v.Len() != 2 || v.Index(0).Interface() != "b" {
		t.Errorf("Expected slice b, c got %v", resSlice)
	}

	resStr, _ := f("hello", "1:4")
	if resStr != "ell" {
		t.Errorf("Expected slice ell, got %v", resStr)
	}

	f = lib.Filters["dictsort"]
	dicts := []map[string]any{
		{"name": "Zebra", "age": 20},
		{"name": "Alpha", "age": 10},
	}
	resDict, _ := f(dicts, "name")
	rList := resDict.([]any)
	if getMapKeyValue(rList[0], "name") != "Alpha" {
		t.Errorf("Expected Alpha first, got %v", rList[0])
	}
}
