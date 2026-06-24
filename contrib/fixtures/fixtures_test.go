package fixtures

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseJSON(t *testing.T) {
	data := `[
		{"model": "auth.user", "pk": 1, "fields": {"username": "admin", "email": "admin@example.com"}},
		{"model": "auth.user", "pk": 2, "fields": {"username": "user1", "email": "user1@example.com"}}
	]`

	loader := NewLoader(nil)
	fixture, err := loader.Parse([]byte(data), JSON)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(fixture.Items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(fixture.Items))
	}

	if fixture.Items[0].Model != "auth.user" {
		t.Errorf("Expected model 'auth.user', got %q", fixture.Items[0].Model)
	}
	if fixture.Items[0].PK != float64(1) {
		t.Errorf("Expected PK 1, got %v", fixture.Items[0].PK)
	}
	if fixture.Items[0].Fields["username"] != "admin" {
		t.Errorf("Expected username 'admin', got %v", fixture.Items[0].Fields["username"])
	}
}

func TestLoadFile(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fixtures_test")
	defer os.RemoveAll(tmpDir)

	items := []FixtureItem{
		{Model: "blog.post", PK: 1, Fields: map[string]any{"title": "Hello", "content": "World"}},
		{Model: "blog.post", PK: 2, Fields: map[string]any{"title": "Foo", "content": "Bar"}},
	}

	data, _ := json.MarshalIndent(items, "", "  ")
	fixturePath := filepath.Join(tmpDir, "posts.json")
	os.WriteFile(fixturePath, data, 0644)

	loader := NewLoader([]string{tmpDir})
	fixture, err := loader.LoadFile(fixturePath)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	if len(fixture.Items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(fixture.Items))
	}
	if fixture.Items[0].Fields["title"] != "Hello" {
		t.Errorf("Expected title 'Hello', got %v", fixture.Items[0].Fields["title"])
	}
}

func TestLoadByName(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fixtures_test")
	defer os.RemoveAll(tmpDir)

	items := []FixtureItem{
		{Model: "auth.user", PK: 1, Fields: map[string]any{"username": "admin"}},
	}
	data, _ := json.MarshalIndent(items, "", "  ")
	os.WriteFile(filepath.Join(tmpDir, "users.json"), data, 0644)

	loader := NewLoader([]string{tmpDir})

	// Test finding by name without extension
	fixture, err := loader.LoadByName("users")
	if err != nil {
		t.Fatalf("LoadByName failed: %v", err)
	}
	if len(fixture.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(fixture.Items))
	}

	// Test finding by name with extension
	fixture, err = loader.LoadByName("users.json")
	if err != nil {
		t.Fatalf("LoadByName with ext failed: %v", err)
	}
	if len(fixture.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(fixture.Items))
	}

	// Test not found
	_, err = loader.LoadByName("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent fixture")
	}
}

func TestApply(t *testing.T) {
	fixture := &Fixture{
		Items: []FixtureItem{
			{Model: "auth.user", PK: 1, Fields: map[string]any{"username": "admin"}},
			{Model: "auth.user", PK: 2, Fields: map[string]any{"username": "user1"}},
		},
	}

	var applied []FixtureItem
	loader := NewLoader(nil)
	loader.Applier = func(item FixtureItem) error {
		applied = append(applied, item)
		return nil
	}

	err := loader.Apply(fixture)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if len(applied) != 2 {
		t.Fatalf("Expected 2 applied items, got %d", len(applied))
	}
	if applied[0].Fields["username"] != "admin" {
		t.Errorf("Expected username 'admin', got %v", applied[0].Fields["username"])
	}
}

func TestDump(t *testing.T) {
	items := []FixtureItem{
		{Model: "blog.post", PK: 1, Fields: map[string]any{"title": "Test"}},
	}

	// Test JSON dump
	data, err := Dump(items, JSON)
	if err != nil {
		t.Fatalf("Dump JSON failed: %v", err)
	}

	var parsed []FixtureItem
	json.Unmarshal(data, &parsed)
	if len(parsed) != 1 || parsed[0].Model != "blog.post" {
		t.Error("JSON dump did not round-trip correctly")
	}

	// Test YAML dump
	data, err = Dump(items, YAML)
	if err != nil {
		t.Fatalf("Dump YAML failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty YAML output")
	}
}

func TestDumpToFile(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fixtures_dump_test")
	defer os.RemoveAll(tmpDir)

	items := []FixtureItem{
		{Model: "blog.post", PK: 1, Fields: map[string]any{"title": "Test"}},
	}

	path := filepath.Join(tmpDir, "output.json")
	err := DumpToFile(items, path)
	if err != nil {
		t.Fatalf("DumpToFile failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	var parsed []FixtureItem
	json.Unmarshal(data, &parsed)
	if len(parsed) != 1 {
		t.Error("Expected 1 item in dumped file")
	}
}

func TestLoaderNoApplier(t *testing.T) {
	loader := NewLoader(nil)
	fixture := &Fixture{Items: []FixtureItem{{Model: "test", PK: 1, Fields: nil}}}

	err := loader.Apply(fixture)
	if err == nil {
		t.Error("Expected error when no Applier configured")
	}
}
