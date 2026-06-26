package queryset

import (
	"testing"

	"github.com/pkshahid/JanGo/core/settings"
	"github.com/pkshahid/JanGo/orm"
	"github.com/pkshahid/JanGo/orm/backends"
)

// setupTestDB creates a SQLite in-memory database, registers it as the
// "default" backend, creates the table schema for the given model, and
// registers cleanup with t.Cleanup.
func setupTestDB(t *testing.T, model any) *orm.ModelInfo {
	t.Helper()

	backend := &backends.SQLiteBackend{}
	if err := backend.Connect(settings.DatabaseConfig{Name: ":memory:"}); err != nil {
		t.Fatalf("failed to connect test DB: %v", err)
	}
	// Ensure a single connection so :memory: is shared.
	backend.DB().SetMaxOpenConns(1)

	backends.ClearBackends()
	backends.RegisterBackendInstance("default", backend)

	info, err := orm.GetModelInfo(model)
	if err != nil {
		t.Fatalf("failed to get model info: %v", err)
	}

	if err := backend.SchemaEditor().CreateTable(info); err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	t.Cleanup(func() {
		backend.Close()
		backends.ClearBackends()
	})

	return info
}
