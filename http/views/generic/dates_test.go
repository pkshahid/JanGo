package generic

import (
	"net/http"
	"net/url"
	"testing"

	godjangohttp "github.com/godjango/godjango/http"
)

func TestArchiveIndexView(t *testing.T) {
	v := ArchiveIndexView[DummyModel]{
		ListView: ListView[DummyModel]{
			MultipleObjectMixin: MultipleObjectMixin[DummyModel]{
				QuerySet: []DummyModel{{ID: 1}},
			},
		},
		DateMixin: DateMixin{DateField: "created_at"},
	}
	urlParsed, _ := url.Parse("http://testserver/")
	req := &godjangohttp.Request{
		Request: &http.Request{URL: urlParsed, Method: "GET"},
	}
	resp := v.Get(req)
	if resp.(*godjangohttp.HttpResponse).StatusCode != http.StatusOK {
		t.Errorf("Expected 200")
	}
}

func TestYearArchiveView(t *testing.T) {
	v := YearArchiveView[DummyModel]{
		ListView: ListView[DummyModel]{
			MultipleObjectMixin: MultipleObjectMixin[DummyModel]{
				QuerySet: []DummyModel{{ID: 1}},
			},
		},
	}
	// Missing year
	urlParsed, _ := url.Parse("http://testserver/")
	req := &godjangohttp.Request{
		Request: &http.Request{URL: urlParsed, Method: "GET"},
	}
	resp := v.Get(req)
	if resp.(*godjangohttp.HttpResponse).StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 without year")
	}

	// With year
	urlParsed2, _ := url.Parse("http://testserver/?year=2023")
	req2 := &godjangohttp.Request{
		Request: &http.Request{URL: urlParsed2, Method: "GET"},
	}
	resp2 := v.Get(req2)
	if resp2.(*godjangohttp.HttpResponse).StatusCode != http.StatusOK {
		t.Errorf("Expected 200 with year")
	}
}

func TestDateDetailView(t *testing.T) {
	v := DateDetailView[DummyModel]{
		DetailView: DetailView[DummyModel]{
			GetObjectFunc: func(req *godjangohttp.Request, mixin *SingleObjectMixin[DummyModel]) (DummyModel, error) {
				return DummyModel{ID: 1}, nil
			},
		},
	}

	// Missing params
	urlParsed, _ := url.Parse("http://testserver/?year=2023&month=12")
	req := &godjangohttp.Request{
		Request: &http.Request{URL: urlParsed, Method: "GET"},
	}
	resp := v.Get(req)
	if resp.(*godjangohttp.HttpResponse).StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for missing day")
	}

	// All params
	urlParsed2, _ := url.Parse("http://testserver/?year=2023&month=12&day=01")
	req2 := &godjangohttp.Request{
		Request: &http.Request{URL: urlParsed2, Method: "GET"},
	}
	resp2 := v.Get(req2)
	if resp2.(*godjangohttp.HttpResponse).StatusCode != http.StatusOK {
		t.Errorf("Expected 200")
	}
}
