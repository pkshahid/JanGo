// Package main demonstrates a simple JanGo web application.
// It registers several routes showing different response types:
// function-based views, JSON responses, URL parameters, and templates.
package main

import (
	"fmt"
	"net/http"
	"os"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/http/urls"

	"github.com/pkshahid/JanGo/core/settings"
	"github.com/pkshahid/JanGo/management"

	// Import runserver command so it registers itself
	_ "github.com/pkshahid/JanGo/management/commands/runserver"
)

func init() {
	// Configure settings (like Django's settings.py)
	_ = settings.Configure(settings.Settings{
		DEBUG:        false, // Disable file watcher for test
		SECRET_KEY:   "hello-world-insecure-key-for-demo",
		ROOT_URLCONF: "hello.urls",
		DATABASES: map[string]settings.DatabaseConfig{
			"default": {
				Engine: "sqlite3",
				Name:   "file::memory:?cache=shared",
			},
		},
		INSTALLED_APPS: []string{"hello"},
	})

	// Register templates
	godjangohttp.RegisterTemplate("home.html", `<!DOCTYPE html>
<html>
<head><title>JanGo Hello World</title></head>
<body>
  <h1>Welcome to JanGo!</h1>
  <p>A Django-like web framework written in pure Go.</p>
  <ul>
    <li><a href="/about/">About</a></li>
    <li><a href="/api/status/">API Status (JSON)</a></li>
    <li><a href="/greet/World/">Greet Example</a></li>
  </ul>
</body>
</html>`)

	godjangohttp.RegisterTemplate("about.html", `<!DOCTYPE html>
<html>
<head><title>About - JanGo</title></head>
<body>
  <h1>About JanGo</h1>
  <p>JanGo is a Go web framework inspired by Django.</p>
  <p>It provides URL routing, middleware, templates, ORM, admin, and more.</p>
  <a href="/">Back to Home</a>
</body>
</html>`)

	godjangohttp.RegisterTemplate("greet.html", `<!DOCTYPE html>
<html>
<head><title>Greeting - JanGo</title></head>
<body>
  <h1>Hello, {{.name}}!</h1>
  <p>This page demonstrates URL parameters (path converters).</p>
  <a href="/">Back to Home</a>
</body>
</html>`)

	// Register URL patterns (like Django's urls.py)
	r := urls.GetGlobalRouter()
	r.Add(urls.Path("/", homeView, "home", nil))
	r.Add(urls.Path("/about/", aboutView, "about", nil))
	r.Add(urls.Path("/api/status/", apiStatusView, "api_status", nil))
	r.Add(urls.Path("/greet/<str:name>/", greetView, "greet", nil))
}

// homeView renders the home page template.
func homeView(req *godjangohttp.Request) godjangohttp.Response {
	return godjangohttp.Render(req, "home.html", nil)
}

// aboutView renders the about page template.
func aboutView(req *godjangohttp.Request) godjangohttp.Response {
	return godjangohttp.Render(req, "about.html", nil)
}

// apiStatusView returns a JSON response (like a Django REST endpoint).
func apiStatusView(req *godjangohttp.Request) godjangohttp.Response {
	return godjangohttp.NewJsonResponse(map[string]any{
		"framework": "JanGo",
		"version":   "0.0.1",
		"status":    "running",
		"go_like":   true,
	})
}

// greetView demonstrates URL path parameters (converters).
func greetView(req *godjangohttp.Request) godjangohttp.Response {
	// Extract the URL parameter from ResolverMatch
	name := "World"
	if match, ok := req.ResolverMatch.(*urls.ResolverMatch); ok {
		if n, exists := match.Kwargs["name"]; exists {
			name = fmt.Sprintf("%v", n)
		}
	}
	return godjangohttp.Render(req, "greet.html", map[string]any{
		"name": name,
	})
}

func main() {
	if err := management.Execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// If no command was provided, start the server directly
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go runserver [addr]")
		fmt.Println("Starting server on :8000...")
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req := godjangohttp.NewRequest(r)
			router := urls.GetGlobalRouter()
			match, err := router.Match(req.Path)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			req.ResolverMatch = match
			resp := match.Func(req)
			resp.Write(w)
		})
		http.ListenAndServe(":8000", handler)
	}
}
