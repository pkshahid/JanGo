package toolbar

import (
	"net"
	"bytes"
	"embed"
	"html/template"
	"context"
	"io"
	"net/http/httptest"
	"strings"

	"github.com/pkshahid/JanGo/core/settings"
	godjangohttp "github.com/pkshahid/JanGo/http"
)

type contextKey string

const toolbarKey contextKey = "djdt_toolbar"

//go:embed templates/*
var templatesFS embed.FS

// GetToolbarFromContext retrieves the toolbar from the request context.
func GetToolbarFromContext(ctx context.Context) *Toolbar {
	if tb, ok := ctx.Value(toolbarKey).(*Toolbar); ok {
		return tb
	}
	return nil
}

func isInternalIP(ip string, internalIPs []string) bool {
	// Safely strip port
	host, _, err := net.SplitHostPort(ip)
	if err == nil {
		ip = host
	}
	// For dev purposes, if INTERNAL_IPS is empty but DEBUG is true, we might allow 127.0.0.1
	if len(internalIPs) == 0 {
		return ip == "127.0.0.1" || ip == "::1" || ip == "localhost"
	}
	for _, allowed := range internalIPs {
		if ip == allowed {
			return true
		}
	}
	return false
}

// DebugToolbarMiddleware injects the debug toolbar.
func DebugToolbarMiddleware(next func(*godjangohttp.Request) godjangohttp.Response) func(*godjangohttp.Request) godjangohttp.Response {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		s := settings.Get()
		if !s.DEBUG {
			return next(req)
		}

		// Check INTERNAL_IPS based on actual RemoteAddr
		// (Do not blindly trust X-Forwarded-For to prevent IP spoofing)
		ip := req.RemoteAddr
		if !isInternalIP(ip, s.INTERNAL_IPS) {
			return next(req)
		}

		// Don't intercept toolbar's own requests
		if strings.HasPrefix(req.Path, "/djdt/") {
			return next(req)
		}

		tb := NewToolbar()

		// Initialize panels here...
		tb.AddPanel(NewTimerPanel())
		tb.AddPanel(NewRequestPanel())
		tb.AddPanel(NewHeadersPanel())
		tb.AddPanel(NewRoutingPanel())
		tb.AddPanel(NewSQLPanel())
		tb.AddPanel(NewTemplatesPanel())
		tb.AddPanel(NewCachePanel())
		tb.AddPanel(NewLoggingPanel())
		tb.AddPanel(NewSignalsPanel())
		tb.AddPanel(NewProfilerPanel())
		// ... etc ...

		// Attach to context
		req.Context = context.WithValue(req.Context, toolbarKey, tb)

		// Process Request for panels
		for _, name := range tb.Ordered {
			p := tb.Panels[name]
			if p.Enable() {
				p.ProcessRequest(req)
			}
		}

		// Execute the view
		resp := next(req)

		// Process Response for panels
		for _, name := range tb.Ordered {
			p := tb.Panels[name]
			if p.Enable() {
				p.ProcessResponse(req, resp)
			}
		}

		tb.Save()

		// Inject HTML if response is HTML
		if resp != nil {
			// We need to inspect the response and inject before </body>
			// godjangohttp.Response has Write(w). We can use an httptest.ResponseRecorder to capture it.
			// Or check if it's an HttpResponse and content-type is HTML.

			rr := httptest.NewRecorder()
			resp.Write(rr)

			res := rr.Result()
			contentType := res.Header.Get("Content-Type")

			body, _ := io.ReadAll(res.Body)

			if strings.Contains(contentType, "text/html") {
				bodyStr := string(body)

				// Basic template for the base toolbar container
				// The Javascript will load panel contents asynchronously

				// Use embedded template
				tmpl, err := template.ParseFS(templatesFS, "templates/toolbar.html")
				if err != nil {
					// Fallback if template is broken
					return godjangohttp.NewHttpResponse(string(body), rr.Code)
				}

				type TemplateData struct {
					ID string
					Panels []Panel
				}

				var orderedPanels []Panel
				for _, name := range tb.Ordered {
					orderedPanels = append(orderedPanels, tb.Panels[name])
				}

				var tplBuf bytes.Buffer
				err = tmpl.Execute(&tplBuf, TemplateData{
					ID: tb.ID,
					Panels: orderedPanels,
				})

				injection := ""
				if err == nil {
					injection = tplBuf.String()
				}

// Insert before </body>
				idx := strings.LastIndex(strings.ToLower(bodyStr), "</body>")
				if idx != -1 {
					bodyStr = bodyStr[:idx] + injection + bodyStr[idx:]
				} else {
					bodyStr += injection
				}

				hr := godjangohttp.NewHttpResponse(bodyStr, rr.Code)
				for k, v := range res.Header {
					for _, vv := range v {
						hr.Headers.Add(k, vv)
					}
				}
				return hr
			}

			// If not HTML, reconstruct the response
			hr := godjangohttp.NewHttpResponse(string(body), rr.Code)
			for k, v := range res.Header {
				for _, vv := range v {
					hr.Headers.Add(k, vv)
				}
			}
			// In case it was Streaming or File, capturing in rr buffers it fully.
			// Ideally we don't buffer if not HTML, but to know Content-Type we usually check headers.
			// Since we want to check Content-Type, we might only buffer if we can determine it,
			// but Response interface doesn't expose Headers directly without writing.
			// Let's stick to this for simplicity in this task.
			return hr
		}

		return resp
	}
}
