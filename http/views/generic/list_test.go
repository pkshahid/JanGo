package generic

import (
	"net/http"
	"net/url"
	"testing"

	godjangohttp "github.com/pkshahid/JanGo/http"
)

func TestListView(t *testing.T) {
	objects := []DummyModel{
		{ID: 1, Name: "A"},
		{ID: 2, Name: "B"},
		{ID: 3, Name: "C"},
	}

	v := ListView[DummyModel]{
		MultipleObjectMixin: MultipleObjectMixin[DummyModel]{
			QuerySet:   objects,
			PaginateBy: 2,
		},
	}
	v.TemplateName = "test.html"

	// Page 1
	urlParsed, _ := url.Parse("http://testserver/?page=1")
	req := &godjangohttp.Request{
		Request: &http.Request{URL: urlParsed, Method: "GET"},
	}

	resp := v.Get(req)
	httpResp := resp.(*godjangohttp.HttpResponse)
	if httpResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", httpResp.StatusCode)
	}

	isPaginated := v.ContextData["is_paginated"].(bool)
	if !isPaginated {
		t.Errorf("Expected is_paginated to be true")
	}

	pageObj := v.ContextData["page_obj"].(*Page[DummyModel])
	if pageObj.Number != 1 {
		t.Errorf("Expected page 1, got %d", pageObj.Number)
	}
	if len(pageObj.ObjectList) != 2 {
		t.Errorf("Expected 2 objects, got %d", len(pageObj.ObjectList))
	}

	// Page 2
	urlParsed2, _ := url.Parse("http://testserver/?page=2")
	req2 := &godjangohttp.Request{
		Request: &http.Request{URL: urlParsed2, Method: "GET"},
	}

	v.Get(req2)
	pageObj2 := v.ContextData["page_obj"].(*Page[DummyModel])
	if pageObj2.Number != 2 {
		t.Errorf("Expected page 2, got %d", pageObj2.Number)
	}
	if len(pageObj2.ObjectList) != 1 {
		t.Errorf("Expected 1 object, got %d", len(pageObj2.ObjectList))
	}
}

func TestListView_AllowEmpty(t *testing.T) {
	v := ListView[DummyModel]{
		MultipleObjectMixin: MultipleObjectMixin[DummyModel]{
			QuerySet:   []DummyModel{},
			AllowEmpty: false,
		},
	}
	urlParsed, _ := url.Parse("http://testserver/")
	req := &godjangohttp.Request{
		Request: &http.Request{URL: urlParsed, Method: "GET"},
	}

	resp := v.Get(req)
	httpResp := resp.(*godjangohttp.HttpResponse)
	if httpResp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", httpResp.StatusCode)
	}
}
