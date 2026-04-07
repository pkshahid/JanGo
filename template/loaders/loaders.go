package loaders

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"io/fs"
)

// Loader defines the interface for all template loaders.
type Loader interface {
	Load(name string) (string, error)
}

// FilesystemLoader searches a list of directories in order.
type FilesystemLoader struct {
	Dirs []string
}

func NewFilesystemLoader(dirs []string) *FilesystemLoader {
	return &FilesystemLoader{Dirs: dirs}
}

func (l *FilesystemLoader) Load(name string) (string, error) {
	for _, dir := range l.Dirs {
		fullPath := filepath.Join(dir, name)
		if content, err := os.ReadFile(fullPath); err == nil {
			return string(content), nil
		}
	}
	return "", fmt.Errorf("template %s not found on filesystem", name)
}

// AppDirectoriesLoader searches `<app>/templates/` in each installed app.
// For this prototype, we pass in the app directories explicitly.
// A full implementation would ask the app registry for paths.
type AppDirectoriesLoader struct {
	AppPaths []string
}

func NewAppDirectoriesLoader(appPaths []string) *AppDirectoriesLoader {
	return &AppDirectoriesLoader{AppPaths: appPaths}
}

func (l *AppDirectoriesLoader) Load(name string) (string, error) {
	for _, appPath := range l.AppPaths {
		fullPath := filepath.Join(appPath, "templates", name)
		if content, err := os.ReadFile(fullPath); err == nil {
			return string(content), nil
		}
	}
	return "", fmt.Errorf("template %s not found in app directories", name)
}

// CachedLoader wraps another loader with a sync.Map cache.
// Note: The template Engine also has a cache of parsed templates.
// This cache stores the raw string, which can be useful if reparsing is needed or engine cache is off.
type CachedLoader struct {
	Loader Loader
	cache  sync.Map
}

func NewCachedLoader(loader Loader) *CachedLoader {
	return &CachedLoader{Loader: loader}
}

func (l *CachedLoader) Load(name string) (string, error) {
	if content, ok := l.cache.Load(name); ok {
		return content.(string), nil
	}
	content, err := l.Loader.Load(name)
	if err != nil {
		return "", err
	}
	l.cache.Store(name, content)
	return content, nil
}

// EmbedLoader loads templates from an embed.FS.
type EmbedLoader struct {
	FS   embed.FS
	Base string
}

func NewEmbedLoader(fs embed.FS, base string) *EmbedLoader {
	return &EmbedLoader{FS: fs, Base: base}
}

func (l *EmbedLoader) Load(name string) (string, error) {
	path := name
	if l.Base != "" {
		path = filepath.Join(l.Base, name)
	}
	// For cross-platform embed path compatibility, ensure forward slashes
	path = filepath.ToSlash(path)

	content, err := fs.ReadFile(l.FS, path)
	if err != nil {
		return "", fmt.Errorf("template %s not found in embed FS: %w", name, err)
	}
	return string(content), nil
}

// LocmemLoader is an in-memory loader for testing.
type LocmemLoader struct {
	Templates map[string]string
}

func NewLocmemLoader(templates map[string]string) *LocmemLoader {
	return &LocmemLoader{Templates: templates}
}

func (l *LocmemLoader) Load(name string) (string, error) {
	if content, ok := l.Templates[name]; ok {
		return content, nil
	}
	return "", fmt.Errorf("template %s not found in locmem", name)
}
