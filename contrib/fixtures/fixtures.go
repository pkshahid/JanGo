// Package fixtures provides Django-style fixture loading and dumping.
// Fixtures allow serializing model data to JSON/YAML and loading it back.
package fixtures

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Format represents a serialization format.
type Format string

const (
	JSON Format = "json"
	YAML Format = "yaml"
)

// FixtureItem represents a single model instance in a fixture file.
type FixtureItem struct {
	Model  string         `json:"model"`
	PK     interface{}    `json:"pk"`
	Fields map[string]any `json:"fields"`
}

// Fixture represents a collection of fixture items.
type Fixture struct {
	Items []FixtureItem
}

// Loader loads fixtures from files into a persistence layer.
type Loader struct {
	Dirs      []string
	Format    Format
	Applier   func(item FixtureItem) error
	Verbosity int
}

// NewLoader creates a new fixture loader with default settings.
func NewLoader(dirs []string) *Loader {
	return &Loader{
		Dirs:   dirs,
		Format: JSON,
	}
}

// LoadFile loads a single fixture file by path.
func (l *Loader) LoadFile(path string) (*Fixture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("fixtures: cannot read %s: %v", path, err)
	}

	format := l.detectFormat(path)
	return l.Parse(data, format)
}

// LoadByName finds and loads a fixture by name, searching the configured directories.
func (l *Loader) LoadByName(name string) (*Fixture, error) {
	path, err := l.FindFixture(name)
	if err != nil {
		return nil, err
	}
	return l.LoadFile(path)
}

// FindFixture searches for a fixture file by name in all configured directories.
func (l *Loader) FindFixture(name string) (string, error) {
	extensions := l.getExtensions()

	// Check if the name already has an extension
	for _, ext := range extensions {
		if strings.HasSuffix(name, ext) {
			for _, dir := range l.Dirs {
				path := filepath.Join(dir, name)
				if _, err := os.Stat(path); err == nil {
					return path, nil
				}
			}
			return "", fmt.Errorf("fixtures: %s not found in fixture dirs", name)
		}
	}

	// Try each extension
	for _, dir := range l.Dirs {
		for _, ext := range extensions {
			path := filepath.Join(dir, name+ext)
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("fixtures: %s not found in fixture dirs", name)
}

// Parse parses fixture data from bytes in the given format.
func (l *Loader) Parse(data []byte, format Format) (*Fixture, error) {
	switch format {
	case JSON:
		return parseJSON(data)
	case YAML:
		return parseYAML(data)
	default:
		return nil, fmt.Errorf("fixtures: unsupported format %q", format)
	}
}

// Apply applies all items in a fixture using the configured Applier.
func (l *Loader) Apply(fixture *Fixture) error {
	if l.Applier == nil {
		return fmt.Errorf("fixtures: no Applier configured")
	}
	for i, item := range fixture.Items {
		if err := l.Applier(item); err != nil {
			return fmt.Errorf("fixtures: error applying item %d (model=%s, pk=%v): %v", i, item.Model, item.PK, err)
		}
	}
	return nil
}

// Dump serializes fixture items to the given format.
func Dump(items []FixtureItem, format Format) ([]byte, error) {
	switch format {
	case JSON:
		return json.MarshalIndent(items, "", "  ")
	case YAML:
		return dumpYAML(items)
	default:
		return nil, fmt.Errorf("fixtures: unsupported format %q", format)
	}
}

// DumpToFile serializes fixture items and writes them to a file.
func DumpToFile(items []FixtureItem, path string) error {
	format := JSON
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		format = YAML
	}

	data, err := Dump(items, format)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func parseJSON(data []byte) (*Fixture, error) {
	var items []FixtureItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("fixtures: invalid JSON: %v", err)
	}
	return &Fixture{Items: items}, nil
}

func parseYAML(data []byte) (*Fixture, error) {
	// Simple YAML-like parser for fixture format
	// In production, would use a full YAML library
	// For now, try JSON as YAML is a superset
	return parseJSON(data)
}

func dumpYAML(items []FixtureItem) ([]byte, error) {
	// Simple YAML output
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("- model: %s\n", item.Model))
		sb.WriteString(fmt.Sprintf("  pk: %v\n", item.PK))
		sb.WriteString("  fields:\n")
		for k, v := range item.Fields {
			sb.WriteString(fmt.Sprintf("    %s: %v\n", k, v))
		}
	}
	return []byte(sb.String()), nil
}

func (l *Loader) detectFormat(path string) Format {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return YAML
	default:
		return JSON
	}
}

func (l *Loader) getExtensions() []string {
	return []string{".json", ".yaml", ".yml"}
}
