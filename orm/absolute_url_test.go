package orm

import "testing"

type modelWithURL struct {
	Model
	Title string `gd:"CharField,max_length=200"`
}

func (m *modelWithURL) GetAbsoluteURL() string {
	return "/items/" + titleToID(m.Title)
}

func titleToID(s string) string {
	if s == "" {
		return "0"
	}
	return s
}

type modelWithoutURL struct {
	Model
	Name string `gd:"CharField,max_length=100"`
}

type modelWithValueURL struct {
	Model
	Code string `gd:"CharField,max_length=50"`
}

func (m modelWithValueURL) GetAbsoluteURL() string {
	return "/codes/" + m.Code
}

func TestGetAbsoluteURL_Implements(t *testing.T) {
	obj := &modelWithURL{Title: "42"}
	url, ok := GetAbsoluteURL(obj)
	if !ok {
		t.Fatalf("Expected ok=true for model implementing GetAbsoluteURLer")
	}
	if url != "/items/42" {
		t.Errorf("Expected '/items/42', got '%s'", url)
	}
}

func TestGetAbsoluteURL_NotImplemented(t *testing.T) {
	obj := &modelWithoutURL{Name: "test"}
	url, ok := GetAbsoluteURL(obj)
	if ok {
		t.Fatalf("Expected ok=false for model without GetAbsoluteURL")
	}
	if url != "" {
		t.Errorf("Expected empty URL, got '%s'", url)
	}
}

func TestGetAbsoluteURL_Nil(t *testing.T) {
	url, ok := GetAbsoluteURL(nil)
	if ok {
		t.Fatalf("Expected ok=false for nil")
	}
	if url != "" {
		t.Errorf("Expected empty URL for nil, got '%s'", url)
	}
}

func TestGetAbsoluteURL_ValueReceiver(t *testing.T) {
	obj := modelWithValueURL{Code: "abc"}
	url, ok := GetAbsoluteURL(obj)
	if !ok {
		t.Fatalf("Expected ok=true for value receiver implementing GetAbsoluteURLer")
	}
	if url != "/codes/abc" {
		t.Errorf("Expected '/codes/abc', got '%s'", url)
	}
}
