package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	godjangohttp "github.com/pkshahid/JanGo/http"
)

func TestRequestLoggingMiddleware(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	mw := RequestLoggingMiddleware(func(req *godjangohttp.Request) godjangohttp.Response {
		return godjangohttp.NewHttpResponse("logging test", http.StatusCreated)
	})

	req := godjangohttp.NewRequest(httptest.NewRequest("GET", "/test-log", nil))
	mw(req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "GET /test-log") {
		t.Errorf("Expected log to contain GET /test-log, got %s", logOutput)
	}
	if !strings.Contains(logOutput, "201") {
		t.Errorf("Expected log to contain status 201, got %s", logOutput)
	}
	// "logging test" is 12 bytes
	if !strings.Contains(logOutput, "12 bytes") {
		t.Errorf("Expected log to contain 12 bytes, got %s", logOutput)
	}
}
