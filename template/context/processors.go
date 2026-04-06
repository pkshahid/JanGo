package context

import (
	"sync"

	"github.com/godjango/godjango/core/settings"
	godjangohttp "github.com/godjango/godjango/http"
	godjango "github.com/godjango/godjango/template"
)

// ProcessorFunc is a function that returns a map of context variables to add to the template.
type ProcessorFunc func(req *godjangohttp.Request) map[string]any

// BuildRequestContext runs all provided context processors concurrently and merges the results.
func BuildRequestContext(req *godjangohttp.Request, processors []ProcessorFunc, base map[string]any) *godjango.Context {
	ctxData := make(map[string]any)
	if base != nil {
		for k, v := range base {
			ctxData[k] = v
		}
	}

	if len(processors) == 0 {
		return godjango.NewContext(ctxData)
	}

	results := make(chan map[string]any, len(processors))
	var wg sync.WaitGroup

	for _, proc := range processors {
		wg.Add(1)
		go func(p ProcessorFunc) {
			defer wg.Done()
			res := p(req)
			if res != nil {
				results <- res
			}
		}(proc)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Merge results
	for res := range results {
		for k, v := range res {
			// Do not overwrite base context variables passed explicitly
			if _, exists := ctxData[k]; !exists {
				ctxData[k] = v
			}
		}
	}

	return godjango.NewContext(ctxData)
}

// RequestContextProcessor adds the request to the context.
func RequestContextProcessor(req *godjangohttp.Request) map[string]any {
	return map[string]any{
		"request": req,
	}
}

// AuthContextProcessor adds the user and permissions to the context.
func AuthContextProcessor(req *godjangohttp.Request) map[string]any {
	var user any
	var perms any

	if req.User != nil {
		user = req.User
		// In a real framework, perms would be a lazy object that checks permissions.
		perms = struct{}{}
	}

	return map[string]any{
		"user":  user,
		"perms": perms,
	}
}

// MessagesContextProcessor adds flash messages to the context.
func MessagesContextProcessor(req *godjangohttp.Request) map[string]any {
	var messages any
	if req.META != nil {
		if msgs, ok := req.META["MESSAGES"]; ok {
			messages = msgs
		}
	}
	return map[string]any{
		"messages": messages,
	}
}

// StaticContextProcessor adds STATIC_URL and MEDIA_URL to the context.
func StaticContextProcessor(req *godjangohttp.Request) map[string]any {
	s := settings.Get()
	return map[string]any{
		"STATIC_URL": s.STATIC_URL,
		"MEDIA_URL":  s.MEDIA_URL,
	}
}

// DebugContextProcessor adds debug info if DEBUG=True and user is staff.
func DebugContextProcessor(req *godjangohttp.Request) map[string]any {
	s := settings.Get()
	if !s.DEBUG {
		return nil
	}

	// Staff check prototype:
	// If User exists, verify they are staff. For prototype, we assume true if authenticated.
	isStaff := false
	if req.User != nil && req.User.IsAuthenticated() {
		isStaff = true
	}

	if isStaff {
		return map[string]any{
			"debug":       true,
			"sql_queries": []string{}, // Mock SQL queries
		}
	}

	return nil
}

// CsrfContextProcessor adds the csrf_token to the context.
func CsrfContextProcessor(req *godjangohttp.Request) map[string]any {
	token := ""
	if req.META != nil {
		if t, ok := req.META["CSRF_TOKEN"]; ok {
			token = t
		}
	}

	// In Django, csrf_token is often a lazy callable.
	// For simplicity, we just provide the string.
	// The `{% csrf_token %}` tag will look for `request.META.CSRF_TOKEN` primarily.
	return map[string]any{
		"csrf_token": token,
	}
}
