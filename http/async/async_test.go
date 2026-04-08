package async

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	godjangohttp "github.com/godjango/godjango/http"
)

func TestSSEResponse(t *testing.T) {
	sse := NewSSEResponse()

	go func() {
		sse.Send("message", "hello world")
		time.Sleep(10 * time.Millisecond)
		sse.Close()
	}()

	rec := httptest.NewRecorder()
	sse.Write(rec)

	if rec.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected event-stream content type")
	}

	body := rec.Body.String()
	if !strings.Contains(body, "event: message") {
		t.Errorf("Missing event name in body")
	}
	if !strings.Contains(body, "data: hello world") {
		t.Errorf("Missing data payload in body")
	}
}

func TestAsyncHandler(t *testing.T) {
	asyncView := func(req *godjangohttp.Request) <-chan godjangohttp.Response {
		ch := make(chan godjangohttp.Response, 1)
		go func() {
			time.Sleep(50 * time.Millisecond)
			ch <- godjangohttp.NewHttpResponse("Async OK", 200)
		}()
		return ch
	}

	handler := AsyncHandler(asyncView)

	r, _ := http.NewRequest("GET", "/", nil)
	req := godjangohttp.NewRequest(r)

	// Test successful resolution
	resp := handler(req)
	httpResp := resp.(*godjangohttp.HttpResponse)
	if httpResp.StatusCode != 200 {
		t.Errorf("Expected 200 OK")
	}

	// Read body
	buf := new(bytes.Buffer)
	buf.ReadFrom(httpResp.Body)
	if buf.String() != "Async OK" {
		t.Errorf("Expected 'Async OK', got %s", buf.String())
	}

	// Test cancellation
	ctx, cancel := context.WithCancel(context.Background())
	r2, _ := http.NewRequestWithContext(ctx, "GET", "/", nil)
	req2 := godjangohttp.NewRequest(r2)

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	resp2 := handler(req2)
	httpResp2 := resp2.(*godjangohttp.HttpResponse)
	if httpResp2.StatusCode != 400 {
		t.Errorf("Expected 400 on cancel, got %d", httpResp2.StatusCode)
	}
}
