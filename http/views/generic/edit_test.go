package generic

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	godjangohttp "github.com/pkshahid/JanGo/http"
)

func TestFormView(t *testing.T) {
	v := FormView{
		FormMixin: FormMixin{
			SuccessUrl: "/success/",
		},
	}

	// GET
	urlParsed, _ := url.Parse("http://testserver/")
	reqGet := &godjangohttp.Request{
		Request: &http.Request{URL: urlParsed, Method: "GET"},
	}
	respGet := v.Get(reqGet)
	if respGet.(*godjangohttp.HttpResponse).StatusCode != http.StatusOK {
		t.Errorf("Expected 200 GET")
	}

	// POST Valid
	reqPostValid := &godjangohttp.Request{
		Request: &http.Request{URL: urlParsed, Method: "POST"},
	}
	// Simulate FormValid
	respPost := v.Post(reqPostValid)
	if respPost.(*godjangohttp.RedirectResponse).StatusCode != http.StatusFound {
		t.Errorf("Expected 302 redirect on valid form POST")
	}

	// POST Invalid
	reqPostInvalid := &godjangohttp.Request{
		Request: &http.Request{
			URL:      urlParsed,
			Method:   "POST",
			PostForm: url.Values{"is_valid": []string{"false"}},
		},
	}
	respInvalid := v.Post(reqPostInvalid)
	if respInvalid.(*godjangohttp.HttpResponse).StatusCode != http.StatusOK {
		t.Errorf("Expected 200 on invalid form POST to render form again")
	}
}

func TestCreateView(t *testing.T) {
	saved := false
	v := CreateView[DummyModel]{
		FormView: FormView{
			FormMixin: FormMixin{SuccessUrl: "/done/"},
		},
		PerformSave: func(req *godjangohttp.Request) error {
			saved = true
			return nil
		},
	}

	urlParsed, _ := url.Parse("http://testserver/")
	reqPost := &godjangohttp.Request{
		Request: &http.Request{URL: urlParsed, Method: "POST"},
	}
	v.Post(reqPost)
	if !saved {
		t.Errorf("Expected PerformSave to be called")
	}
}

func TestUpdateView(t *testing.T) {
	updated := false
	v := UpdateView[DummyModel]{
		DetailView: DetailView[DummyModel]{
			GetObjectFunc: func(req *godjangohttp.Request, mixin *SingleObjectMixin[DummyModel]) (DummyModel, error) {
				return DummyModel{ID: 1}, nil
			},
		},
		FormMixin: FormMixin{SuccessUrl: "/done/"},
		PerformSave: func(req *godjangohttp.Request, obj DummyModel) error {
			if obj.ID == 1 {
				updated = true
			}
			return nil
		},
	}

	urlParsed, _ := url.Parse("http://testserver/")
	reqPost := &godjangohttp.Request{
		Request: &http.Request{URL: urlParsed, Method: "POST"},
	}
	v.Post(reqPost)
	if !updated {
		t.Errorf("Expected PerformSave to be called with correct object")
	}
}

func TestDeleteView(t *testing.T) {
	deleted := false
	v := DeleteView[DummyModel]{
		DetailView: DetailView[DummyModel]{
			GetObjectFunc: func(req *godjangohttp.Request, mixin *SingleObjectMixin[DummyModel]) (DummyModel, error) {
				if req.URL.Query().Get("pk") == "1" {
					return DummyModel{ID: 1}, nil
				}
				return DummyModel{}, fmt.Errorf("not found")
			},
		},
		SuccessUrl: "/deleted/",
		PerformDelete: func(req *godjangohttp.Request, obj DummyModel) error {
			if obj.ID == 1 {
				deleted = true
			}
			return nil
		},
	}

	urlParsed, _ := url.Parse("http://testserver/?pk=1")
	reqPost := &godjangohttp.Request{
		Request: &http.Request{URL: urlParsed, Method: "POST"},
	}
	resp := v.Post(reqPost)
	if !deleted {
		t.Errorf("Expected PerformDelete to be called")
	}
	if resp.(*godjangohttp.RedirectResponse).StatusCode != http.StatusFound {
		t.Errorf("Expected 302 redirect after delete")
	}
}
