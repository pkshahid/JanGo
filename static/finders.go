package static

import (
	"errors"
	"github.com/pkshahid/JanGo/core/apps"
	"github.com/pkshahid/JanGo/core/settings"
	"os"
	"path/filepath"
)

// Finder interface for finding static files.
type Finder interface {
	Find(name string) (string, error)
	List() ([]string, error)
}

// FileSystemFinder looks for static files in STATICFILES_DIRS.
type FileSystemFinder struct{}

func (f *FileSystemFinder) Find(name string) (string, error) {
	s := settings.Get()
	for _, dir := range s.STATICFILES_DIRS {
		fullPath := filepath.Join(dir, name)
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			return fullPath, nil
		}
	}
	return "", errors.New("file not found in STATICFILES_DIRS")
}

func (f *FileSystemFinder) List() ([]string, error) {
	s := settings.Get()
	var files []string
	for _, dir := range s.STATICFILES_DIRS {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				rel, err := filepath.Rel(dir, path)
				if err == nil {
					files = append(files, rel)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

// AppDirectoriesFinder looks for static files in each app's "static" directory.
type AppDirectoriesFinder struct{}

func (f *AppDirectoriesFinder) Find(name string) (string, error) {
	for _, appConfig := range apps.All() {
		dir := filepath.Join(appConfig.Path(), "static")
		fullPath := filepath.Join(dir, name)
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			return fullPath, nil
		}
	}
	return "", errors.New("file not found in app static directories")
}

func (f *AppDirectoriesFinder) List() ([]string, error) {
	var files []string
	for _, appConfig := range apps.All() {
		dir := filepath.Join(appConfig.Path(), "static")
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				rel, err := filepath.Rel(dir, path)
				if err == nil {
					files = append(files, rel)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

// DefaultStorageFinder looks for static files in the default STATICFILES_STORAGE.
// To avoid cyclical dependencies, it's implemented using Storage getter.
type DefaultStorageFinder struct{}

func (f *DefaultStorageFinder) Find(name string) (string, error) {
	storage := GetStorage()
	if storage.Exists(name) {
		return storage.Path(name), nil
	}
	return "", errors.New("file not found in default storage")
}

func (f *DefaultStorageFinder) List() ([]string, error) {
	storage := GetStorage()
	return storage.ListDir("")
}

// GetFinders returns instances of all configured finders.
func GetFinders() []Finder {
	s := settings.Get()
	var finders []Finder
	finderNames := s.STATICFILES_FINDERS
	if len(finderNames) == 0 {
		finderNames = []string{
			"FileSystemFinder",
			"AppDirectoriesFinder",
		}
	}

	for _, name := range finderNames {
		switch name {
		case "FileSystemFinder", "static.FileSystemFinder":
			finders = append(finders, &FileSystemFinder{})
		case "AppDirectoriesFinder", "static.AppDirectoriesFinder":
			finders = append(finders, &AppDirectoriesFinder{})
		case "DefaultStorageFinder", "static.DefaultStorageFinder":
			finders = append(finders, &DefaultStorageFinder{})
		}
	}
	return finders
}

// Find looks for a file using all configured finders and returns the absolute path.
func Find(name string) (string, error) {
	for _, finder := range GetFinders() {
		if path, err := finder.Find(name); err == nil {
			return path, nil
		}
	}
	return "", errors.New("file not found: " + name)
}
