package generic

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	godjangohttp "github.com/godjango/godjango/http"
)

type DummyModel struct {
	ID   int
	Name string
}

func TestDetailView_GetObjectFunc(t *testing.T) {
	v := DetailView[DummyModel]{
		GetObjectFunc: func(req *godjangohttp.Request, mixin *SingleObjectMixin[DummyModel]) (DummyModel, error) {
			if req.URL.Query().Get("pk") == "1" {
				return DummyModel{ID: 1, Name: "Test Object"}, nil
			}
			return DummyModel{}, fmt.Errorf("not found")
		},
	}
	v.TemplateName = "test.html"

	urlParsed, _ := url.Parse("http://testserver/?pk=1")
	req := &godjangohttp.Request{
		Request: &http.Request{URL: urlParsed, Method: "GET"},
	}

	resp := v.Get(req)
	httpResp := resp.(*godjangohttp.HttpResponse)
	if httpResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", httpResp.StatusCode)
	}

	// Test 404
	urlParsed2, _ := url.Parse("http://testserver/?pk=2")
	req2 := &godjangohttp.Request{
		Request: &http.Request{URL: urlParsed2, Method: "GET"},
	}
	resp2 := v.Get(req2)
	httpResp2 := resp2.(*godjangohttp.HttpResponse)
	if httpResp2.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", httpResp2.StatusCode)
	}
}
