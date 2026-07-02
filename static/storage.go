package static

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pkshahid/JanGo/core/settings"
	"io"

	"os"
	"path/filepath"
	"strings"
	"sync"
)

// StaticFilesStorage interface
type StaticFilesStorage interface {
	Path(name string) string
	URL(name string) string
	Exists(name string) bool
	ListDir(path string) ([]string, error)
	Save(name string, content io.Reader) error
}

// FileSystemStorage is the base storage using local file system.
type FileSystemStorage struct {
	Location string
	BaseURL  string
}

func NewFileSystemStorage() *FileSystemStorage {
	s := settings.Get()
	return &FileSystemStorage{
		Location: s.STATIC_ROOT,
		BaseURL:  s.STATIC_URL,
	}
}

func (s *FileSystemStorage) Path(name string) string {
	return filepath.Join(s.Location, name)
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

func (s *FileSystemStorage) Exists(name string) bool {
	_, err := os.Stat(s.Path(name))
	return err == nil
}

func (s *FileSystemStorage) ListDir(path string) ([]string, error) {
	var files []string
	baseDir := filepath.Join(s.Location, path)
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return files, nil
	}
	err := filepath.Walk(baseDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			rel, err := filepath.Rel(s.Location, p)
			if err == nil {
				files = append(files, rel)
			}
		}
		return nil
	})
	return files, err
}

func (s *FileSystemStorage) Save(name string, content io.Reader) error {
	fullPath := s.Path(name)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	outFile, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, content)
	return err
}

// ManifestStaticFilesStorage hashes filenames and maintains a mapping.
type ManifestStaticFilesStorage struct {
	*FileSystemStorage
	Manifest     map[string]string
	manifestLock sync.RWMutex
}

func NewManifestStaticFilesStorage() *ManifestStaticFilesStorage {
	fs := NewFileSystemStorage()
	m := &ManifestStaticFilesStorage{
		FileSystemStorage: fs,
		Manifest:          make(map[string]string),
	}
	m.loadManifest()
	return m
}

func (m *ManifestStaticFilesStorage) loadManifest() {
	manifestPath := filepath.Join(m.Location, "staticfiles.json")
	data, err := os.ReadFile(manifestPath)
	if err == nil {
		m.manifestLock.Lock()
		defer m.manifestLock.Unlock()
		json.Unmarshal(data, &m.Manifest)
	}
}

func (m *ManifestStaticFilesStorage) saveManifest() error {
	m.manifestLock.RLock()
	data, err := json.MarshalIndent(m.Manifest, "", "  ")
	m.manifestLock.RUnlock()
	if err != nil {
		return err
	}
	manifestPath := filepath.Join(m.Location, "staticfiles.json")
	return os.WriteFile(manifestPath, data, 0644)
}

func (m *ManifestStaticFilesStorage) URL(name string) string {
	m.manifestLock.RLock()
	hashedName, ok := m.Manifest[name]
	m.manifestLock.RUnlock()

	if ok {
		return m.FileSystemStorage.URL(hashedName)
	}
	return m.FileSystemStorage.URL(name)
}

// HashFile calculates MD5 hash of file content.
func HashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	// Take first 12 chars of hex encoded hash
	return hex.EncodeToString(hash.Sum(nil))[:12], nil
}

// PostProcess is used by collectstatic to hash files concurrently.
func (m *ManifestStaticFilesStorage) PostProcess() error {
	files, err := m.ListDir("")
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(files))
	semaphore := make(chan struct{}, 10) // Limit concurrent hashing

	var mtx sync.Mutex // Protects m.Manifest modifications

	for _, name := range files {
		if name == "staticfiles.json" {
			continue
		}
		wg.Add(1)
		go func(fileName string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			path := m.Path(fileName)
			hash, err := HashFile(path)
			if err != nil {
				errChan <- fmt.Errorf("error hashing %s: %w", fileName, err)
				return
			}

			// Generate hashed name
			ext := filepath.Ext(fileName)
			base := fileName[:len(fileName)-len(ext)]
			hashedName := fmt.Sprintf("%s.%s%s", base, hash, ext)

			// Copy to hashed name
			hashedPath := m.Path(hashedName)
			src, err := os.Open(path)
			if err != nil {
				errChan <- fmt.Errorf("error opening %s: %w", fileName, err)
				return
			}
			defer src.Close()

			dst, err := os.Create(hashedPath)
			if err != nil {
				errChan <- fmt.Errorf("error creating %s: %w", hashedName, err)
				return
			}
			defer dst.Close()

			if _, err := io.Copy(dst, src); err != nil {
				errChan <- fmt.Errorf("error copying to %s: %w", hashedName, err)
				return
			}

			// Update manifest
			mtx.Lock()
			m.Manifest[fileName] = hashedName
			mtx.Unlock()
		}(name)
	}

	wg.Wait()
	close(errChan)

	var errs []string
	for err := range errChan {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return m.saveManifest()
}

// GetStorage returns the configured StaticFilesStorage.
func GetStorage() StaticFilesStorage {
	s := settings.Get()
	storageType := s.STATICFILES_STORAGE

	if storageType == "ManifestStaticFilesStorage" || storageType == "static.ManifestStaticFilesStorage" {
		return NewManifestStaticFilesStorage()
	}

	return NewFileSystemStorage()
}
