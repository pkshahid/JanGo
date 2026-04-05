package http

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRequest(t *testing.T) {
	// Create a multipart body
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("field1", "value1")
	writer.Close()

	r := httptest.NewRequest(http.MethodPost, "/test/path?foo=bar", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	r.Header.Set("User-Agent", "TestAgent")
	r.AddCookie(&http.Cookie{Name: "sessionid", Value: "12345"})

	req := NewRequest(r)

	if req.Method != http.MethodPost {
		t.Errorf("expected POST, got %s", req.Method)
	}
	if req.Path != "/test/path" {
		t.Errorf("expected /test/path, got %s", req.Path)
	}
	if req.GET.Get("foo") != "bar" {
		t.Errorf("expected bar, got %s", req.GET.Get("foo"))
	}
	if req.POST.Get("field1") != "value1" {
		t.Errorf("expected value1, got %s", req.POST.Get("field1"))
	}
	if req.COOKIES["sessionid"] != "12345" {
		t.Errorf("expected 12345, got %s", req.COOKIES["sessionid"])
	}
	if req.META["HTTP_USER_AGENT"] != "TestAgent" {
		t.Errorf("expected TestAgent, got %s", req.META["HTTP_USER_AGENT"])
	}
}

func TestContextValues(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	req := NewRequest(r)

	type contextKey string
	key := contextKey("testKey")
	val := "testValue"

	newReq := WithValue(req, key, val)

	if Value(newReq, key) != val {
		t.Errorf("expected %s, got %v", val, Value(newReq, key))
	}

	// Verify the original request remains unchanged
	if Value(req, key) != nil {
		t.Errorf("original request context should be unmodified")
	}

	// Verify that the underlying *http.Request was updated
	if newReq.Request.Context().Value(key) != val {
		t.Errorf("underlying http.Request context should have the value")
	}
}
