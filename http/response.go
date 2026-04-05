package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
)

// Response is the interface that all GoDjango responses must implement.
type Response interface {
	Write(w http.ResponseWriter)
}

// HttpResponse is the base HTTP response type.
type HttpResponse struct {
	StatusCode int
	Headers    http.Header
	Body       io.ReadCloser
}

func (r *HttpResponse) Write(w http.ResponseWriter) {
	for k, v := range r.Headers {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}
	w.WriteHeader(r.StatusCode)
	if r.Body != nil {
		defer r.Body.Close()
		io.Copy(w, r.Body)
	}
}

// NewHttpResponse creates a new base HttpResponse.
func NewHttpResponse(content string, status int) *HttpResponse {
	return &HttpResponse{
		StatusCode: status,
		Headers:    make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(content)),
	}
}

// JsonResponse marshals the given data to JSON and sets Content-Type.
type JsonResponse struct {
	HttpResponse
}

// NewJsonResponse creates a new JsonResponse.
func NewJsonResponse(data any) *JsonResponse {
	b, err := json.Marshal(data)
	if err != nil {
		return &JsonResponse{
			HttpResponse: HttpResponse{
				StatusCode: http.StatusInternalServerError,
				Headers:    make(http.Header),
				Body:       io.NopCloser(bytes.NewBufferString("Internal Server Error")),
			},
		}
	}
	resp := &JsonResponse{
		HttpResponse: HttpResponse{
			StatusCode: http.StatusOK,
			Headers:    make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(b)),
		},
	}
	resp.Headers.Set("Content-Type", "application/json")
	return resp
}

// RedirectResponse handles 301 and 302 redirects.
type RedirectResponse struct {
	HttpResponse
	URL       string
	Permanent bool
}

// NewRedirectResponse creates a new RedirectResponse.
func NewRedirectResponse(url string, permanent bool) *RedirectResponse {
	status := http.StatusFound
	if permanent {
		status = http.StatusMovedPermanently
	}
	resp := &RedirectResponse{
		HttpResponse: HttpResponse{
			StatusCode: status,
			Headers:    make(http.Header),
		},
		URL:       url,
		Permanent: permanent,
	}
	resp.Headers.Set("Location", url)
	return resp
}

// StreamingHttpResponse streams data using a goroutine and channel.
type StreamingHttpResponse struct {
	HttpResponse
	Stream <-chan []byte
}

func (r *StreamingHttpResponse) Write(w http.ResponseWriter) {
	for k, v := range r.Headers {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}
	w.WriteHeader(r.StatusCode)

	flusher, canFlush := w.(http.Flusher)
	for chunk := range r.Stream {
		w.Write(chunk)
		if canFlush {
			flusher.Flush()
		}
	}
}

// NewStreamingHttpResponse creates a new StreamingHttpResponse.
func NewStreamingHttpResponse(stream <-chan []byte) *StreamingHttpResponse {
	return &StreamingHttpResponse{
		HttpResponse: HttpResponse{
			StatusCode: http.StatusOK,
			Headers:    make(http.Header),
		},
		Stream: stream,
	}
}

// FileResponse serves a file.
type FileResponse struct {
	HttpResponse
	FilePath string
}

func (r *FileResponse) Write(w http.ResponseWriter) {
	// Let http.ServeFile handle writing headers, status, and body
	// We construct a fake request here since ServeFile requires it to handle Range requests, etc.
	// But since this is a low-level framework implementation, we often serve via w and req.
	// Here we need to implement Response interface, which takes only w.
	// We'll open the file and copy it, setting basic headers if not handled elsewhere.

	f, err := os.Open(r.FilePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err == nil {
		w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	}
	for k, v := range r.Headers {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}
	w.WriteHeader(r.StatusCode)
	io.Copy(w, f)
}

// NewFileResponse creates a new FileResponse.
func NewFileResponse(filePath string) *FileResponse {
	return &FileResponse{
		HttpResponse: HttpResponse{
			StatusCode: http.StatusOK,
			Headers:    make(http.Header),
		},
		FilePath: filePath,
	}
}

// HttpResponseNotFound represents a 404 response.
func HttpResponseNotFound(content string) *HttpResponse {
	if content == "" {
		content = "Not Found"
	}
	return NewHttpResponse(content, http.StatusNotFound)
}

// HttpResponseForbidden represents a 403 response.
func HttpResponseForbidden(content string) *HttpResponse {
	if content == "" {
		content = "Forbidden"
	}
	return NewHttpResponse(content, http.StatusForbidden)
}

// HttpResponseBadRequest represents a 400 response.
func HttpResponseBadRequest(content string) *HttpResponse {
	if content == "" {
		content = "Bad Request"
	}
	return NewHttpResponse(content, http.StatusBadRequest)
}
