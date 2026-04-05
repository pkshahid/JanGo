package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHttpResponse(t *testing.T) {
	resp := NewHttpResponse("hello world", http.StatusOK)
	resp.Headers.Set("X-Custom-Header", "test")

	recorder := httptest.NewRecorder()
	resp.Write(recorder)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, recorder.Code)
	}
	if recorder.Body.String() != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", recorder.Body.String())
	}
	if recorder.Header().Get("X-Custom-Header") != "test" {
		t.Errorf("expected 'test', got '%s'", recorder.Header().Get("X-Custom-Header"))
	}
}

func TestJsonResponse(t *testing.T) {
	data := map[string]string{"message": "success"}
	resp := NewJsonResponse(data)

	recorder := httptest.NewRecorder()
	resp.Write(recorder)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, recorder.Code)
	}
	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json, got %s", recorder.Header().Get("Content-Type"))
	}

	var parsed map[string]string
	json.Unmarshal(recorder.Body.Bytes(), &parsed)
	if parsed["message"] != "success" {
		t.Errorf("expected success, got %s", parsed["message"])
	}
}

func TestRedirectResponse(t *testing.T) {
	resp := NewRedirectResponse("/login", false)

	recorder := httptest.NewRecorder()
	resp.Write(recorder)

	if recorder.Code != http.StatusFound {
		t.Errorf("expected %d, got %d", http.StatusFound, recorder.Code)
	}
	if recorder.Header().Get("Location") != "/login" {
		t.Errorf("expected /login, got %s", recorder.Header().Get("Location"))
	}
}

func TestStreamingHttpResponse(t *testing.T) {
	ch := make(chan []byte)
	resp := NewStreamingHttpResponse(ch)

	go func() {
		ch <- []byte("part1 ")
		ch <- []byte("part2")
		close(ch)
	}()

	recorder := httptest.NewRecorder()
	resp.Write(recorder)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, recorder.Code)
	}
	if recorder.Body.String() != "part1 part2" {
		t.Errorf("expected 'part1 part2', got '%s'", recorder.Body.String())
	}
}

func TestFileResponse(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := []byte("file content")
	tmpFile.Write(content)
	tmpFile.Close()

	resp := NewFileResponse(tmpFile.Name())
	recorder := httptest.NewRecorder()
	resp.Write(recorder)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, recorder.Code)
	}
	if recorder.Body.String() != "file content" {
		t.Errorf("expected 'file content', got '%s'", recorder.Body.String())
	}
}

func TestErrorResponses(t *testing.T) {
	resp404 := HttpResponseNotFound("")
	if resp404.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp404.StatusCode)
	}

	resp403 := HttpResponseForbidden("")
	if resp403.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp403.StatusCode)
	}

	resp400 := HttpResponseBadRequest("")
	if resp400.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp400.StatusCode)
	}
}
