package media

import (
	"io"
)

// FieldFile acts as a proxy to a file stored using a Storage backend.
// It is returned by FileField and ImageField on model instances.
type FieldFile struct {
	Name    string
	Storage Storage
}

// NewFieldFile creates a FieldFile wrapping the default or given storage.
func NewFieldFile(name string, storage Storage) *FieldFile {
	if storage == nil {
		storage = GetStorage()
	}
	return &FieldFile{
		Name:    name,
		Storage: storage,
	}
}

// URL returns the URL for this file.
func (f *FieldFile) URL() string {
	if f.Name == "" || f.Storage == nil {
		return ""
	}
	return f.Storage.URL(f.Name)
}

// Open opens the file from the storage.
func (f *FieldFile) Open() (io.ReadCloser, error) {
	if f.Storage == nil {
		return nil, io.ErrClosedPipe
	}
	return f.Storage.Open(f.Name)
}

// Save saves new content to the file and updates its Name.
func (f *FieldFile) Save(name string, content io.Reader) error {
	if f.Storage == nil {
		f.Storage = GetStorage()
	}
	savedName, err := f.Storage.Save(name, content)
	if err != nil {
		return err
	}
	f.Name = savedName
	return nil
}

// Delete deletes the file from storage.
func (f *FieldFile) Delete() error {
	if f.Name == "" || f.Storage == nil {
		return nil
	}
	err := f.Storage.Delete(f.Name)
	if err == nil {
		f.Name = ""
	}
	return err
}

// Size returns the file size.
func (f *FieldFile) Size() (int64, error) {
	if f.Name == "" || f.Storage == nil {
		return 0, nil
	}
	return f.Storage.Size(f.Name)
}
