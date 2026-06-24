package contenttypes

import (
	"reflect"
	"testing"
)

type Article struct {
	ID    int
	Title string
}

type Comment struct {
	ID            int
	Text          string
	ContentTypeID int
	ObjectID      int
}

func TestRegisterAndGet(t *testing.T) {
	Clear()

	ct := Register("blog", "article", reflect.TypeOf(Article{}))
	if ct.ID != 1 {
		t.Errorf("expected ID 1, got %d", ct.ID)
	}
	if ct.AppLabel != "blog" {
		t.Errorf("expected app_label 'blog', got %q", ct.AppLabel)
	}
	if ct.Model != "article" {
		t.Errorf("expected model 'article', got %q", ct.Model)
	}

	// Test Get by ID
	found, err := Get(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != ct {
		t.Error("Get returned different content type")
	}

	// Test Get non-existent
	_, err = Get(999)
	if err == nil {
		t.Error("expected error for non-existent ID")
	}
}

func TestGetForModel(t *testing.T) {
	Clear()

	Register("blog", "article", reflect.TypeOf(Article{}))

	ct, err := GetForModel("blog", "article")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ct.Model != "article" {
		t.Errorf("expected model 'article', got %q", ct.Model)
	}

	// Test case-insensitive lookup
	ct2, err := GetForModel("Blog", "Article")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ct2 != ct {
		t.Error("case-insensitive lookup failed")
	}
}

func TestGetForGoType(t *testing.T) {
	Clear()

	Register("blog", "article", reflect.TypeOf(Article{}))

	ct, err := GetForGoType(reflect.TypeOf(Article{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ct.Model != "article" {
		t.Errorf("expected model 'article', got %q", ct.Model)
	}
}

func TestGenericForeignKey(t *testing.T) {
	Clear()

	Register("blog", "article", reflect.TypeOf(Article{}))

	gfk := NewGenericForeignKey(1, 42)
	ct, err := gfk.Resolve()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ct.Model != "article" {
		t.Errorf("expected 'article', got %q", ct.Model)
	}
}

func TestGenericRelation(t *testing.T) {
	Clear()

	gr := NewGenericRelation(reflect.TypeOf(Comment{}), "", "")
	if gr.ContentTypeField != "ContentTypeID" {
		t.Errorf("expected default ContentTypeField, got %q", gr.ContentTypeField)
	}
	if gr.ObjectIDField != "ObjectID" {
		t.Errorf("expected default ObjectIDField, got %q", gr.ObjectIDField)
	}
}

func TestPolymorphicQuery(t *testing.T) {
	Clear()

	ct1 := Register("blog", "article", reflect.TypeOf(Article{}))
	ct2 := Register("blog", "comment", reflect.TypeOf(Comment{}))

	pq := NewPolymorphicQuery(ct1, ct2)

	byType := pq.FilterByType("article")
	if len(byType) != 1 {
		t.Errorf("expected 1 result, got %d", len(byType))
	}

	byApp := pq.FilterByApp("blog")
	if len(byApp) != 2 {
		t.Errorf("expected 2 results, got %d", len(byApp))
	}
}

func TestAll(t *testing.T) {
	Clear()

	Register("blog", "article", reflect.TypeOf(Article{}))
	Register("blog", "comment", reflect.TypeOf(Comment{}))

	all := All()
	if len(all) != 2 {
		t.Errorf("expected 2 content types, got %d", len(all))
	}
}

func TestDuplicateRegister(t *testing.T) {
	Clear()

	ct1 := Register("blog", "article", reflect.TypeOf(Article{}))
	ct2 := Register("blog", "article", reflect.TypeOf(Article{}))

	if ct1 != ct2 {
		t.Error("duplicate registration should return same content type")
	}
}

func TestGetObjectForType(t *testing.T) {
	Clear()

	ct := Register("blog", "article", reflect.TypeOf(Article{}))

	obj, err := GetObjectForType(ct)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	article, ok := obj.(*Article)
	if !ok {
		t.Fatalf("expected *Article, got %T", obj)
	}
	if article.ID != 0 {
		t.Error("expected zero value")
	}
}

func TestContentTypeString(t *testing.T) {
	Clear()

	ct := Register("blog", "article", reflect.TypeOf(Article{}))
	expected := "blog | article"
	if ct.String() != expected {
		t.Errorf("expected %q, got %q", expected, ct.String())
	}
}
