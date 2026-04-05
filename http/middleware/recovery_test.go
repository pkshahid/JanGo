package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"
	"io"

	godjangohttp "github.com/godjango/godjango/http"
)

func TestPanicRecoveryMiddleware(t *testing.T) {
	setupTestSettings()

	handler := PanicRecoveryMiddleware(func(req *godjangohttp.Request) godjangohttp.Response {
		panic("test panic")
	})

	req := godjangohttp.NewRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req).(*godjangohttp.HttpResponse)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected 500 Internal Server Error, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Internal Server Error") && !strings.Contains(string(body), "Server Error") {
		t.Errorf("Expected Server Error message in body, got: %s", string(body))
	}
}
