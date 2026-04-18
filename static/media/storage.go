package media

import (
	"github.com/pkshahid/JanGo/core/settings"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"crypto/rand"
	"encoding/hex"
)

// Storage defines the interface for file storage backends.
type Storage interface {
	Save(name string, content io.Reader) (string, error)
	Open(name string) (io.ReadCloser, error)
	Delete(name string) error
	Exists(name string) bool
	URL(name string) string
	Size(name string) (int64, error)
	ListDir(path string) ([]string, []string, error)
	GetAvailableName(name string) string
}

// FileSystemStorage implements Storage for the local file system.
type FileSystemStorage struct {
	Location string
	BaseURL  string
}

// NewFileSystemStorage creates a new FileSystemStorage instance.
func NewFileSystemStorage() *FileSystemStorage {
	s := settings.Get()
	return &FileSystemStorage{
		Location: s.MEDIA_ROOT,
		BaseURL:  s.MEDIA_URL,
	}
}

// Path returns the absolute path on the file system.
func (s *FileSystemStorage) Path(name string) string {
	return filepath.Join(s.Location, name)
}

func (s *FileSystemStorage) Save(name string, content io.Reader) (string, error) {
	// Sanitize name to prevent path traversal
	name = filepath.Clean(name)
	name = filepath.ToSlash(name)
	if strings.Contains(name, "..") || strings.HasPrefix(name, "/") || filepath.IsAbs(name) {
		return "", fmt.Errorf("invalid file name")
	}

	name = s.GetAvailableName(name)
	fullPath := s.Path(name)

	// Double check it's inside Location
	absLocation, _ := filepath.Abs(s.Location)
	absFullPath, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absFullPath, absLocation) {
		return "", fmt.Errorf("invalid file path escapes media root")
	}

	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	outFile, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, content); err != nil {
		return "", err
	}

	return name, nil
}

func (s *FileSystemStorage) Open(name string) (io.ReadCloser, error) {
	return os.Open(s.Path(name))
}

func (s *FileSystemStorage) Delete(name string) error {
	return os.Remove(s.Path(name))
}

func (s *FileSystemStorage) Exists(name string) bool {
	_, err := os.Stat(s.Path(name))
	return err == nil
}

func (s *FileSystemStorage) URL(name string) string {
	baseURL := s.BaseURL
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	// For URL, normalize separators
	name = strings.ReplaceAll(name, string(os.PathSeparator), "/")
	return baseURL + name
}

func (s *FileSystemStorage) Size(name string) (int64, error) {
	info, err := os.Stat(s.Path(name))
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (s *FileSystemStorage) ListDir(path string) ([]string, []string, error) {
	var dirs []string
	var files []string

	baseDir := filepath.Join(s.Location, path)
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		} else {
			files = append(files, entry.Name())
		}
	}

	return dirs, files, nil
}

func (s *FileSystemStorage) GetAvailableName(name string) string {
	if !s.Exists(name) {
		return name
	}

	ext := filepath.Ext(name)
	base := name[:len(name)-len(ext)]

	counter := 1
	for {
		// e.g. file_1.jpg
		newName := fmt.Sprintf("%s_%d%s", base, counter, ext)
		if !s.Exists(newName) {
			return newName
		}
		counter++

		if counter > 1000 {
			// fallback to random hex
			bytes := make([]byte, 4)
			rand.Read(bytes)
			return fmt.Sprintf("%s_%s%s", base, hex.EncodeToString(bytes), ext)
		}
	}
}

// GetStorage returns the configured Storage instance.
func GetStorage() Storage {
	s := settings.Get()
	storageType := s.DEFAULT_FILE_STORAGE

	if storageType == "S3Storage" || storageType == "media.S3Storage" {
		// return NewS3Storage()
	}

	return NewFileSystemStorage()
}
