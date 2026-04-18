package template

import (
	"fmt"
	"os"
	"sync"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/template/loaders"
)

// Engine is responsible for loading and rendering templates.
type Engine struct {
	DIRS       []string
	APP_DIRS   bool
	Autoescape bool
	Libraries  map[string]*Library
	Builtins   []*Library
	Loaders    []loaders.Loader

	cache      map[string]*Template
	mu         sync.RWMutex
}

// NewEngine creates a new template engine.
func NewEngine(dirs []string, appDirs bool) *Engine {
	engine := &Engine{
		DIRS:       dirs,
		APP_DIRS:   appDirs,
		Autoescape: true,
		Libraries:  make(map[string]*Library),
		Builtins:   []*Library{},
		cache:      make(map[string]*Template),
		Loaders:    []loaders.Loader{},
	}

	// Add default loaders
	if len(dirs) > 0 {
		engine.Loaders = append(engine.Loaders, loaders.NewFilesystemLoader(dirs))
	}
	// Simulated AppDirectoriesLoader based on APP_DIRS setting
	if appDirs {
		// Mock empty string list, in a real framework this queries the core/apps registry
		engine.Loaders = append(engine.Loaders, loaders.NewAppDirectoriesLoader([]string{}))
	}

	return engine
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

	// Iterate through registered loaders to find template content
	var content string
	var err error
	var loaded bool

	for _, loader := range e.Loaders {
		content, err = loader.Load(name)
		if err == nil {
			loaded = true
			break
		}
	}

	if !loaded {
		// Fallback for tests if loaders are not explicitly set but file is locally accessible
		if _, statErr := os.Stat(name); statErr == nil {
			if fileContent, readErr := os.ReadFile(name); readErr == nil {
				content = string(fileContent)
				loaded = true
			}
		}
	}

	if !loaded {
		return nil, fmt.Errorf("template %s not found by any loader", name)
	}

	tmpl, err := e.FromString(content)
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
