// Package storage provides Django-style file storage backends.
// It supports local filesystem, in-memory, and can be extended with
// cloud storage backends (S3, GCS, Azure Blob).
package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Storage is the interface all storage backends must implement.
// It mirrors Django's Storage class API.
type Storage interface {
	// Open opens a file by name and returns a reader.
	Open(name string) (io.ReadCloser, error)

	// Save stores content under the given name, returning the actual name used.
	Save(name string, content io.Reader) (string, error)

	// Delete removes a file by name.
	Delete(name string) error

	// Exists checks if a file exists.
	Exists(name string) (bool, error)

	// ListDir returns directories and files at the given path.
	ListDir(path string) (dirs []string, files []string, err error)

	// Size returns the file size in bytes.
	Size(name string) (int64, error)

	// URL returns the URL for the given file name.
	URL(name string) string

	// ModifiedTime returns the last modified time of the file.
	ModifiedTime(name string) (time.Time, error)
}

// FileSystemStorage implements Storage using the local filesystem.
// Equivalent to Django's FileSystemStorage.
type FileSystemStorage struct {
	Location string // Base directory for file storage
	BaseURL  string // URL prefix for serving files
}

// NewFileSystemStorage creates a new filesystem storage backend.
func NewFileSystemStorage(location, baseURL string) *FileSystemStorage {
	return &FileSystemStorage{
		Location: location,
		BaseURL:  baseURL,
	}
}

func (fs *FileSystemStorage) path(name string) string {
	return filepath.Join(fs.Location, name)
}

// Open opens a file from the filesystem.
func (fs *FileSystemStorage) Open(name string) (io.ReadCloser, error) {
	return os.Open(fs.path(name))
}

// Save stores content to the filesystem. If a file with the same name exists,
// a unique name is generated.
func (fs *FileSystemStorage) Save(name string, content io.Reader) (string, error) {
	name = fs.getAvailableName(name)
	fullPath := fs.path(name)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("storage: cannot create directory: %v", err)
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("storage: cannot create file: %v", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, content); err != nil {
		return "", fmt.Errorf("storage: cannot write file: %v", err)
	}

	return name, nil
}

// Delete removes a file from the filesystem.
func (fs *FileSystemStorage) Delete(name string) error {
	return os.Remove(fs.path(name))
}

// Exists checks if a file exists on the filesystem.
func (fs *FileSystemStorage) Exists(name string) (bool, error) {
	_, err := os.Stat(fs.path(name))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// ListDir lists directories and files at the given path.
func (fs *FileSystemStorage) ListDir(path string) ([]string, []string, error) {
	fullPath := fs.path(path)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, nil, err
	}

	var dirs, files []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		} else {
			files = append(files, entry.Name())
		}
	}
	return dirs, files, nil
}

// Size returns the file size in bytes.
func (fs *FileSystemStorage) Size(name string) (int64, error) {
	info, err := os.Stat(fs.path(name))
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// URL returns the URL for the given file name.
func (fs *FileSystemStorage) URL(name string) string {
	return strings.TrimRight(fs.BaseURL, "/") + "/" + strings.TrimLeft(name, "/")
}

// ModifiedTime returns the last modified time of the file.
func (fs *FileSystemStorage) ModifiedTime(name string) (time.Time, error) {
	info, err := os.Stat(fs.path(name))
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

func (fs *FileSystemStorage) getAvailableName(name string) string {
	exists, _ := fs.Exists(name)
	if !exists {
		return name
	}

	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s_%d%s", base, i, ext)
		exists, _ := fs.Exists(candidate)
		if !exists {
			return candidate
		}
	}
}

// MemoryStorage implements Storage using in-memory maps (useful for testing).
type MemoryStorage struct {
	mu      sync.RWMutex
	files   map[string][]byte
	baseURL string
}

// NewMemoryStorage creates a new in-memory storage backend.
func NewMemoryStorage(baseURL string) *MemoryStorage {
	return &MemoryStorage{
		files:   make(map[string][]byte),
		baseURL: baseURL,
	}
}

// Open returns a reader for the named file.
func (ms *MemoryStorage) Open(name string) (io.ReadCloser, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	data, ok := ms.files[name]
	if !ok {
		return nil, fmt.Errorf("storage: file %q not found", name)
	}
	return io.NopCloser(strings.NewReader(string(data))), nil
}

// Save stores content in memory.
func (ms *MemoryStorage) Save(name string, content io.Reader) (string, error) {
	data, err := io.ReadAll(content)
	if err != nil {
		return "", err
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.files[name] = data
	return name, nil
}

// Delete removes a file from memory.
func (ms *MemoryStorage) Delete(name string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.files, name)
	return nil
}

// Exists checks if a file exists in memory.
func (ms *MemoryStorage) Exists(name string) (bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	_, ok := ms.files[name]
	return ok, nil
}

// ListDir lists files at the given path prefix.
func (ms *MemoryStorage) ListDir(path string) ([]string, []string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	prefix := path
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	dirSet := make(map[string]bool)
	var files []string
	for name := range ms.files {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		rel := strings.TrimPrefix(name, prefix)
		parts := strings.SplitN(rel, "/", 2)
		if len(parts) == 1 {
			files = append(files, parts[0])
		} else {
			dirSet[parts[0]] = true
		}
	}

	dirs := make([]string, 0, len(dirSet))
	for d := range dirSet {
		dirs = append(dirs, d)
	}
	return dirs, files, nil
}

// Size returns the size of a file in memory.
func (ms *MemoryStorage) Size(name string) (int64, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	data, ok := ms.files[name]
	if !ok {
		return 0, fmt.Errorf("storage: file %q not found", name)
	}
	return int64(len(data)), nil
}

// URL returns the URL for the given file name.
func (ms *MemoryStorage) URL(name string) string {
	return strings.TrimRight(ms.baseURL, "/") + "/" + strings.TrimLeft(name, "/")
}

// ModifiedTime returns a zero time for memory storage.
func (ms *MemoryStorage) ModifiedTime(name string) (time.Time, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if _, ok := ms.files[name]; !ok {
		return time.Time{}, fmt.Errorf("storage: file %q not found", name)
	}
	return time.Now(), nil
}

// DefaultStorage is the global default storage backend.
var DefaultStorage Storage = NewFileSystemStorage("./media", "/media/")

// SetDefaultStorage sets the global default storage backend.
func SetDefaultStorage(s Storage) {
	DefaultStorage = s
}
