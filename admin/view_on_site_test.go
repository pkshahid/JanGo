package admin

import (
	"testing"

	"github.com/pkshahid/JanGo/orm"
)

type modelWithAbsoluteURL struct {
	orm.Model
	Slug string `gd:"CharField,max_length=100"`
}

func (m *modelWithAbsoluteURL) GetAbsoluteURL() string {
	return "/items/" + m.Slug
}

type modelWithoutAbsoluteURL struct {
	orm.Model
	Name string `gd:"CharField,max_length=100"`
}

func TestViewOnSiteURL_AutoDetect(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&modelWithAbsoluteURL{})

	ma, err := NewModelAdmin(&modelWithAbsoluteURL{})
	if err != nil {
		t.Fatalf("Failed to create ModelAdmin: %v", err)
	}

	obj := &modelWithAbsoluteURL{Slug: "hello-world"}
	url := ma.ViewOnSiteURL(obj)
	if url != "/items/hello-world" {
		t.Errorf("Expected '/items/hello-world', got '%s'", url)
	}
}

func TestViewOnSiteURL_NotImplemented(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&modelWithoutAbsoluteURL{})

	ma, err := NewModelAdmin(&modelWithoutAbsoluteURL{})
	if err != nil {
		t.Fatalf("Failed to create ModelAdmin: %v", err)
	}

	obj := &modelWithoutAbsoluteURL{Name: "test"}
	url := ma.ViewOnSiteURL(obj)
	if url != "" {
		t.Errorf("Expected empty URL for model without GetAbsoluteURL, got '%s'", url)
	}
}

func TestViewOnSiteURL_Disabled(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&modelWithAbsoluteURL{})

	ma, err := NewModelAdmin(&modelWithAbsoluteURL{})
	if err != nil {
		t.Fatalf("Failed to create ModelAdmin: %v", err)
	}

	ma.ViewOnSite = false
	obj := &modelWithAbsoluteURL{Slug: "hello-world"}
	url := ma.ViewOnSiteURL(obj)
	if url != "" {
		t.Errorf("Expected empty URL when ViewOnSite=false, got '%s'", url)
	}
}

func TestViewOnSiteURL_CustomFunc(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&modelWithAbsoluteURL{})

	ma, err := NewModelAdmin(&modelWithAbsoluteURL{})
	if err != nil {
		t.Fatalf("Failed to create ModelAdmin: %v", err)
	}

	ma.ViewOnSite = func(obj any) string {
		m := obj.(*modelWithAbsoluteURL)
		return "/custom/" + m.Slug
	}

	obj := &modelWithAbsoluteURL{Slug: "test"}
	url := ma.ViewOnSiteURL(obj)
	if url != "/custom/test" {
		t.Errorf("Expected '/custom/test', got '%s'", url)
	}
}

func TestViewOnSiteURL_EnabledBool(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&modelWithAbsoluteURL{})

	ma, err := NewModelAdmin(&modelWithAbsoluteURL{})
	if err != nil {
		t.Fatalf("Failed to create ModelAdmin: %v", err)
	}

	ma.ViewOnSite = true
	obj := &modelWithAbsoluteURL{Slug: "foo"}
	url := ma.ViewOnSiteURL(obj)
	if url != "/items/foo" {
		t.Errorf("Expected '/items/foo' when ViewOnSite=true, got '%s'", url)
	}
}
