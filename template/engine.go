package template

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	godjangohttp "github.com/godjango/godjango/http"
)

// Engine is responsible for loading and rendering templates.
type Engine struct {
	DIRS       []string
	APP_DIRS   bool
	Autoescape bool
	Libraries  map[string]*Library
	Builtins   []*Library

	cache      map[string]*Template
	mu         sync.RWMutex
}

// NewEngine creates a new template engine.
func NewEngine(dirs []string, appDirs bool) *Engine {
	return &Engine{
		DIRS:       dirs,
		APP_DIRS:   appDirs,
		Autoescape: true,
		Libraries:  make(map[string]*Library),
		Builtins:   []*Library{},
		cache:      make(map[string]*Template),
	}
}

// RegisterLibrary makes a library available for `{% load %}`
func (e *Engine) RegisterLibrary(name string, lib *Library) {
	e.Libraries[name] = lib
}

// AddBuiltin adds a library that is always loaded
func (e *Engine) AddBuiltin(lib *Library) {
	e.Builtins = append(e.Builtins, lib)
}

// Template represents a parsed template.
type Template struct {
	name   string
	engine *Engine
	nodes  NodeList
}

func (t *Template) Name() string {
	return t.name
}

// Render renders the template with a given context.
func (t *Template) Render(ctx *Context) (string, error) {
	if ctx == nil {
		ctx = NewContext(nil)
	}
	ctx.SetRenderState("engine", t.engine)
	return t.nodes.Render(ctx)
}

// RenderToResponse renders the template and returns an HttpResponse.
func (t *Template) RenderToResponse(req *godjangohttp.Request, ctx *Context) godjangohttp.Response {
	if ctx == nil {
		ctx = NewContext(nil)
	}
	// Add request to context globally for the template
	ctx.Set("request", req)

	content, err := t.Render(ctx)
	if err != nil {
		return godjangohttp.NewHttpResponse(fmt.Sprintf("Template Error: %v", err), 500)
	}

	return godjangohttp.NewHttpResponse(content, 200)
}

// GetTemplate loads a template by name, optionally from cache.
func (e *Engine) GetTemplate(name string) (*Template, error) {
	e.mu.RLock()
	if t, ok := e.cache[name]; ok {
		e.mu.RUnlock()
		return t, nil
	}
	e.mu.RUnlock()

	// Try to find the file
	var filePath string
	var found bool

	// 1. Check DIRS
	for _, dir := range e.DIRS {
		fullPath := filepath.Join(dir, name)
		if _, err := os.Stat(fullPath); err == nil {
			filePath = fullPath
			found = true
			break
		}
	}

	// 2. Check APP_DIRS (Simulated for this implementation)
	// In Django, it iterates through INSTALLED_APPS and checks app_dir/templates/
	if !found && e.APP_DIRS {
		// Mock logic: we assume apps are top-level directories
		// We would iterate over core/apps.All() here
	}

	if !found {
		// Just for testing, see if it exists as a relative path directly
		if _, err := os.Stat(name); err == nil {
			filePath = name
			found = true
		}
	}

	if !found {
		return nil, fmt.Errorf("template %s not found", name)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tmpl, err := e.FromString(string(content))
	if err != nil {
		return nil, err
	}
	tmpl.name = name

	e.mu.Lock()
	e.cache[name] = tmpl
	e.mu.Unlock()

	return tmpl, nil
}

// FromString compiles a template directly from a string.
func (e *Engine) FromString(content string) (*Template, error) {
	lexer := NewLexer(content)
	tokens := lexer.Lex()

	parser := NewParser(tokens, e)
	nodes, err := parser.Parse(nil)
	if err != nil {
		return nil, err
	}

	return &Template{
		name:   "<string>",
		engine: e,
		nodes:  nodes,
	}, nil
}
