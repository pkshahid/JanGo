package media_test

import (
	"bytes"
	"github.com/pkshahid/JanGo/core/settings"
	"github.com/pkshahid/JanGo/static/media"
	"io/ioutil"
	"os"
	"testing"
)

func setupSettings() {
	s := settings.Settings{
		MEDIA_URL:    "/media/",
		MEDIA_ROOT:   "test_media_root",
		SECRET_KEY:   "secret",
		ROOT_URLCONF: "urls",
	}
	settings.Configure(s)
	os.MkdirAll("test_media_root", 0755)
}

func cleanupSettings() {
	os.RemoveAll("test_media_root")
}

func TestFileSystemStorage(t *testing.T) {
	setupSettings()
	defer cleanupSettings()

	storage := media.NewFileSystemStorage()

	// Test Save
	content := bytes.NewBufferString("test content")
	name, err := storage.Save("test.txt", content)
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	if name != "test.txt" {
		t.Errorf("Expected test.txt, got %s", name)
	}

	if !storage.Exists("test.txt") {
		t.Errorf("Expected file to exist")
	}

	// Test Open
	file, err := storage.Open("test.txt")
	if err != nil {
		t.Fatalf("Failed to open: %v", err)
	}
	defer file.Close()

	data, _ := ioutil.ReadAll(file)
	if string(data) != "test content" {
		t.Errorf("Expected 'test content', got %q", string(data))
	}

	// Test Size
	size, err := storage.Size("test.txt")
	if err != nil || size != 12 {
		t.Errorf("Expected size 12, got %d (err: %v)", size, err)
	}

	// Test GetAvailableName
	content2 := bytes.NewBufferString("test content 2")
	name2, _ := storage.Save("test.txt", content2)
	if name2 != "test_1.txt" {
		t.Errorf("Expected test_1.txt, got %s", name2)
	}

	// Test URL
	url := storage.URL("test.txt")
	if url != "/media/test.txt" {
		t.Errorf("Expected /media/test.txt, got %s", url)
	}

	// Test Delete
	err = storage.Delete("test.txt")
	if err != nil {
		t.Errorf("Failed to delete: %v", err)
	}
	if storage.Exists("test.txt") {
		t.Errorf("File should not exist after deletion")
	}
}

func TestFieldFile(t *testing.T) {
	setupSettings()
	defer cleanupSettings()

	storage := media.NewFileSystemStorage()
	ff := media.NewFieldFile("test.txt", storage)

	content := bytes.NewBufferString("hello")
	err := ff.Save("test.txt", content)
	if err != nil {
		t.Fatalf("FieldFile.Save failed: %v", err)
	}

	if ff.Name != "test.txt" {
		t.Errorf("FieldFile name mismatch")
	}

	if ff.URL() != "/media/test.txt" {
		t.Errorf("FieldFile URL mismatch: %s", ff.URL())
	}

	ff.Delete()
	if storage.Exists("test.txt") {
		t.Errorf("FieldFile.Delete failed")
	}
}
