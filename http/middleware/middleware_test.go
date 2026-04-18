package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	godjangohttp "github.com/pkshahid/JanGo/http"
)

func TestChain(t *testing.T) {
	order := []string{}

	mw1 := func(next Handler) Handler {
		return func(req *godjangohttp.Request) godjangohttp.Response {
			order = append(order, "mw1_req")
			resp := next(req)
			order = append(order, "mw1_res")
			return resp
		}
	}

	mw2 := func(next Handler) Handler {
		return func(req *godjangohttp.Request) godjangohttp.Response {
			order = append(order, "mw2_req")
			resp := next(req)
			order = append(order, "mw2_res")
			return resp
		}
	}

	finalHandler := func(req *godjangohttp.Request) godjangohttp.Response {
		order = append(order, "final")
		return godjangohttp.NewHttpResponse("OK", http.StatusOK)
	}

	chain := NewChain(mw1, mw2)
	handler := chain.Then(finalHandler)

	req := godjangohttp.NewRequest(httptest.NewRequest("GET", "/", nil))
	handler(req)

	expected := []string{"mw1_req", "mw2_req", "final", "mw2_res", "mw1_res"}
	if len(order) != len(expected) {
		t.Fatalf("expected order %v, got %v", expected, order)
	}

	for i, val := range expected {
		if order[i] != val {
			t.Errorf("expected %s at index %d, got %s", val, i, order[i])
		}
	}
}
