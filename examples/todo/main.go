// Package main demonstrates a Todo web application built with the JanGo framework.
// It showcases CRUD operations, form handling, URL routing, and template rendering.
//
// Run with:
//
//	go run main.go runserver
package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/http/urls"

	"github.com/pkshahid/JanGo/core/settings"
	"github.com/pkshahid/JanGo/management"

	// Import runserver command so it registers itself
	_ "github.com/pkshahid/JanGo/management/commands/runserver"
)

// Todo represents a single todo item.
type Todo struct {
	ID        uint64
	Title     string
	Completed bool
	CreatedAt time.Time
}

// In-memory store for todos (in a real app, this would use the ORM).
var (
	todos     []Todo
	todoMu    sync.Mutex
	nextID    uint64 = 1
)

func init() {
	// Configure settings
	_ = settings.Configure(settings.Settings{
		DEBUG:        true,
		SECRET_KEY:   "todo-app-insecure-key-for-demo",
		ROOT_URLCONF: "todo.urls",
		DATABASES: map[string]settings.DatabaseConfig{
			"default": {
				Engine: "sqlite3",
				Name:   "file::memory:?cache=shared",
			},
		},
		INSTALLED_APPS: []string{"todo"},
	})

	// Seed some example todos
	todos = []Todo{
		{ID: 1, Title: "Learn JanGo framework", Completed: true, CreatedAt: time.Now().Add(-2 * time.Hour)},
		{ID: 2, Title: "Build a todo app", Completed: false, CreatedAt: time.Now().Add(-1 * time.Hour)},
		{ID: 3, Title: "Deploy to production", Completed: false, CreatedAt: time.Now()},
	}
	nextID = 4

	// Register templates
	if err := godjangohttp.RegisterTemplate("todo_list.html", todoListTemplate); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: failed to register todo_list.html: %v\n", err)
		os.Exit(1)
	}
	if err := godjangohttp.RegisterTemplate("todo_about.html", todoAboutTemplate); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: failed to register todo_about.html: %v\n", err)
		os.Exit(1)
	}

	// Register URL patterns
	r := urls.GetGlobalRouter()
	r.Add(urls.Path("/", todoListView, "todo_list", nil))
	r.Add(urls.Path("/add/", todoAddView, "todo_add", nil))
	r.Add(urls.Path("/toggle/<int:id>/", todoToggleView, "todo_toggle", nil))
	r.Add(urls.Path("/delete/<int:id>/", todoDeleteView, "todo_delete", nil))
	r.Add(urls.Path("/about/", todoAboutView, "todo_about", nil))
}

// todoListView renders the main todo list page.
func todoListView(req *godjangohttp.Request) godjangohttp.Response {
	todoMu.Lock()
	defer todoMu.Unlock()

	// Count stats
	total := len(todos)
	completed := 0
	for _, t := range todos {
		if t.Completed {
			completed++
		}
	}
	pending := total - completed

	return godjangohttp.Render(req, "todo_list.html", map[string]any{
		"todos":      todos,
		"total":      total,
		"completed":  completed,
		"pending":    pending,
		"csrf_token": req.META["CSRF_TOKEN"],
	})
}

// todoAddView handles adding a new todo item via POST.
func todoAddView(req *godjangohttp.Request) godjangohttp.Response {
	if req.Method != http.MethodPost {
		return godjangohttp.HttpResponseBadRequest("Only POST allowed")
	}

	title := ""
	if titles, ok := req.POST["title"]; ok && len(titles) > 0 {
		title = titles[0]
	}

	if title == "" {
		return godjangohttp.NewRedirectResponse("/", false)
	}

	todoMu.Lock()
	todos = append(todos, Todo{
		ID:        nextID,
		Title:     title,
		Completed: false,
		CreatedAt: time.Now(),
	})
	nextID++
	todoMu.Unlock()

	return godjangohttp.NewRedirectResponse("/", false)
}

// todoToggleView toggles the completed status of a todo item.
func todoToggleView(req *godjangohttp.Request) godjangohttp.Response {
	if req.Method != http.MethodPost {
		return godjangohttp.HttpResponseBadRequest("Only POST allowed")
	}

	id := extractID(req)
	if id == 0 {
		return godjangohttp.HttpResponseNotFound("Todo not found")
	}

	todoMu.Lock()
	for i := range todos {
		if todos[i].ID == id {
			todos[i].Completed = !todos[i].Completed
			break
		}
	}
	todoMu.Unlock()

	return godjangohttp.NewRedirectResponse("/", false)
}

// todoDeleteView deletes a todo item.
func todoDeleteView(req *godjangohttp.Request) godjangohttp.Response {
	if req.Method != http.MethodPost {
		return godjangohttp.HttpResponseBadRequest("Only POST allowed")
	}

	id := extractID(req)
	if id == 0 {
		return godjangohttp.HttpResponseNotFound("Todo not found")
	}

	todoMu.Lock()
	for i := range todos {
		if todos[i].ID == id {
			todos = append(todos[:i], todos[i+1:]...)
			break
		}
	}
	todoMu.Unlock()

	return godjangohttp.NewRedirectResponse("/", false)
}

// todoAboutView renders the about page.
func todoAboutView(req *godjangohttp.Request) godjangohttp.Response {
	return godjangohttp.Render(req, "todo_about.html", nil)
}

// extractID gets the URL parameter "id" from the resolver match.
func extractID(req *godjangohttp.Request) uint64 {
	if match, ok := req.ResolverMatch.(*urls.ResolverMatch); ok {
		if idStr, exists := match.Kwargs["id"]; exists {
			if id, err := strconv.ParseUint(fmt.Sprintf("%v", idStr), 10, 64); err == nil {
				return id
			}
		}
	}
	return 0
}

func main() {
	if err := management.Execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
