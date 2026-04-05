package http

import (
	"context"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/godjango/godjango/auth"
	"github.com/godjango/godjango/sessions"
)

// ResolverMatch contains information about the matched URL.
type ResolverMatch struct {
	Func     any
	Args     []string
	Kwargs   map[string]string
	URLName  string
	AppNames []string
}

// Request wraps an standard *http.Request and adds GoDjango specific fields.
type Request struct {
	context.Context
	*http.Request

	Method        string
	Path          string
	GET           url.Values
	POST          url.Values
	FILES         map[string][]*multipart.FileHeader
	COOKIES       map[string]string
	META          map[string]string
	User          auth.User
	Session       sessions.Session
	ResolverMatch *ResolverMatch
}

// NewRequest creates a new GoDjango Request from a standard http.Request.
func NewRequest(r *http.Request) *Request {
	req := &Request{
		Context: r.Context(),
		Request: r,
		Method:  r.Method,
		Path:    r.URL.Path,
		GET:     r.URL.Query(),
		POST:    make(url.Values),
		FILES:   make(map[string][]*multipart.FileHeader),
		COOKIES: make(map[string]string),
		META:    make(map[string]string),
	}

	// Parse form data to populate POST and FILES
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		// 32 MB default memory
		if err := r.ParseMultipartForm(32 << 20); err == nil {
			req.POST = r.MultipartForm.Value
			req.FILES = r.MultipartForm.File
		}
	} else if err := r.ParseForm(); err == nil {
		req.POST = r.PostForm
	}

	// Populate COOKIES
	for _, cookie := range r.Cookies() {
		req.COOKIES[cookie.Name] = cookie.Value
	}

	// Populate META (HTTP headers normalized, e.g., HTTP_USER_AGENT)
	for key, values := range r.Header {
		metaKey := "HTTP_" + strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
		if len(values) > 0 {
			req.META[metaKey] = values[0]
		}
	}

	// Add remote address to META
	req.META["REMOTE_ADDR"] = r.RemoteAddr

	return req
}

// WithValue creates a new Request with a context containing the specified key and value.
func WithValue(req *Request, key, val any) *Request {
	newReq := *req
	newReq.Context = context.WithValue(req.Context, key, val)
	// Important: also update the underlying *http.Request context
	newReq.Request = req.Request.WithContext(newReq.Context)
	return &newReq
}

// Value returns the value associated with the key in the request's context.
func Value(req *Request, key any) any {
	return req.Context.Value(key)
}
