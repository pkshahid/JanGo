package storage

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestFileSystemStorage(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "storage_test")
	defer os.RemoveAll(tmpDir)

	fs := NewFileSystemStorage(tmpDir, "/media/")

	// Test Save
	content := strings.NewReader("Hello, World!")
	name, err := fs.Save("test.txt", content)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if name != "test.txt" {
		t.Errorf("Expected name 'test.txt', got %q", name)
	}

	// Test Exists
	exists, err := fs.Exists("test.txt")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Expected file to exist")
	}

	// Test Open
	reader, err := fs.Open("test.txt")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	data, _ := io.ReadAll(reader)
	_ = reader.Close()
	if string(data) != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got %q", string(data))
	}

	// Test Size
	size, err := fs.Size("test.txt")
	if err != nil {
		t.Fatalf("Size failed: %v", err)
	}
	if size != 13 {
		t.Errorf("Expected size 13, got %d", size)
	}

	// Test URL
	url := fs.URL("test.txt")
	if url != "/media/test.txt" {
		t.Errorf("Expected '/media/test.txt', got %q", url)
	}

	// Test ModifiedTime
	_, err = fs.ModifiedTime("test.txt")
	if err != nil {
		t.Fatalf("ModifiedTime failed: %v", err)
	}

	// Test duplicate name handling
	content2 := strings.NewReader("Second file")
	name2, err := fs.Save("test.txt", content2)
	if err != nil {
		t.Fatalf("Save duplicate failed: %v", err)
	}
	if name2 == "test.txt" {
		t.Error("Expected different name for duplicate")
	}
	if name2 != "test_1.txt" {
		t.Errorf("Expected 'test_1.txt', got %q", name2)
	}

	// Test ListDir
	dirs, files, err := fs.ListDir("")
	if err != nil {
		t.Fatalf("ListDir failed: %v", err)
	}
	if len(dirs) != 0 {
		t.Errorf("Expected no dirs, got %v", dirs)
	}
	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %v", files)
	}

	// Test Delete
	err = fs.Delete("test.txt")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	exists, _ = fs.Exists("test.txt")
	if exists {
		t.Error("Expected file to not exist after delete")
	}
}

func TestFileSystemStorageSubdirectory(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "storage_test")
	defer os.RemoveAll(tmpDir)

	fs := NewFileSystemStorage(tmpDir, "/media/")

	// Save to subdirectory
	content := strings.NewReader("nested file")
	name, err := fs.Save("sub/dir/file.txt", content)
	if err != nil {
		t.Fatalf("Save to subdir failed: %v", err)
	}
	if name != "sub/dir/file.txt" {
		t.Errorf("Expected 'sub/dir/file.txt', got %q", name)
	}

	exists, _ := fs.Exists("sub/dir/file.txt")
	if !exists {
		t.Error("Expected nested file to exist")
	}
}

func TestMemoryStorage(t *testing.T) {
	ms := NewMemoryStorage("/static/")

	// Test Save
	content := strings.NewReader("memory content")
	name, err := ms.Save("file.txt", content)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if name != "file.txt" {
		t.Errorf("Expected 'file.txt', got %q", name)
	}

	// Test Exists
	exists, _ := ms.Exists("file.txt")
	if !exists {
		t.Error("Expected file to exist")
	}

	// Test Open
	reader, err := ms.Open("file.txt")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	data, _ := io.ReadAll(reader)
	_ = reader.Close()
	if string(data) != "memory content" {
		t.Errorf("Expected 'memory content', got %q", string(data))
	}

	// Test Size
	size, err := ms.Size("file.txt")
	if err != nil {
		t.Fatalf("Size failed: %v", err)
	}
	if size != 14 {
		t.Errorf("Expected size 14, got %d", size)
	}

	// Test URL
	url := ms.URL("file.txt")
	if url != "/static/file.txt" {
		t.Errorf("Expected '/static/file.txt', got %q", url)
	}

	// Test ListDir with nested files
	if _, err := ms.Save("docs/readme.md", strings.NewReader("readme")); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if _, err := ms.Save("docs/guide.md", strings.NewReader("guide")); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if _, err := ms.Save("images/logo.png", strings.NewReader("logo")); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	dirs, files, _ := ms.ListDir("")
	if len(files) != 1 { // file.txt
		t.Errorf("Expected 1 file in root, got %d: %v", len(files), files)
	}
	if len(dirs) != 2 { // docs, images
		t.Errorf("Expected 2 dirs, got %d: %v", len(dirs), dirs)
	}

	// Test Delete
	if err := ms.Delete("file.txt"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	exists, _ = ms.Exists("file.txt")
	if exists {
		t.Error("Expected file to not exist after delete")
	}

	// Test Open nonexistent
	_, err = ms.Open("nonexistent")
	if err == nil {
		t.Error("Expected error opening nonexistent file")
	}
}

func TestDefaultStorage(t *testing.T) {
	// Test that DefaultStorage is set
	if DefaultStorage == nil {
		t.Fatal("DefaultStorage should not be nil")
	}

	// Test SetDefaultStorage
	ms := NewMemoryStorage("/test/")
	SetDefaultStorage(ms)

	if DefaultStorage != ms {
		t.Error("SetDefaultStorage did not update DefaultStorage")
	}

	// Restore default
	SetDefaultStorage(NewFileSystemStorage("./media", "/media/"))
}
