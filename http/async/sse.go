package async

import (
	"fmt"
	"net/http"
	"time"

	godjangohttp "github.com/godjango/godjango/http"
)

import (
	"sync"
)

// SSEResponse wraps StreamingHttpResponse to provide Server-Sent Events behavior.
type SSEResponse struct {
	stream chan []byte
	resp   *godjangohttp.StreamingHttpResponse
	done   chan struct{}
	mu     sync.Mutex
	closed bool
}

// NewSSEResponse creates a new SSE response.
func NewSSEResponse() *SSEResponse {
	stream := make(chan []byte, 100)
	resp := godjangohttp.NewStreamingHttpResponse(stream)

	resp.Headers.Set("Content-Type", "text/event-stream")
	resp.Headers.Set("Cache-Control", "no-cache")
	resp.Headers.Set("Connection", "keep-alive")

	sse := &SSEResponse{
		stream: stream,
		resp:   resp,
		done:   make(chan struct{}),
	}

	// Start heartbeat to keep connection alive
	go sse.heartbeat()

	return sse
}

// Send sends an event and data payload.
func (s *SSEResponse) Send(event, data string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return fmt.Errorf("sse connection closed")
	}

	var payload string
	if event != "" {
		payload += fmt.Sprintf("event: %s\n", event)
	}
	payload += fmt.Sprintf("data: %s\n\n", data)

	s.stream <- []byte(payload)
	return nil
}

// Close signals that the stream is finished.
func (s *SSEResponse) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	s.mu.Unlock()

	close(s.done)
	close(s.stream)
}

// Write writes the response using the underlying StreamingHttpResponse.
func (s *SSEResponse) Write(w http.ResponseWriter) {
	s.resp.Write(w)
}

func (s *SSEResponse) heartbeat() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			// Send a comment as a heartbeat
			s.mu.Lock()
			if !s.closed {
				s.stream <- []byte(":\n\n")
			}
			s.mu.Unlock()
		}
	}
}
