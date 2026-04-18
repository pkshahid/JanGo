package migrations

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkshahid/JanGo/core/settings"
	"github.com/pkshahid/JanGo/orm"
	"github.com/pkshahid/JanGo/orm/backends"
)

func setupTestDB() string {
	tmpDir, _ := os.MkdirTemp("", "godjango_migrate")
	dbPath := filepath.Join(tmpDir, "migrate.sqlite3")

	settings.Configure(settings.Settings{
		DATABASES: map[string]settings.DatabaseConfig{
			"default": {
				Engine: "sqlite",
				Name:   dbPath,
			},
		},
	})
	backends.Init()
	return dbPath
}

func TestMigrationExecutor(t *testing.T) {
	dbPath := setupTestDB()
	defer os.RemoveAll(filepath.Dir(dbPath))
	defer backends.Close()

	// Setup migrations
	globalMigrationRegistry = make(map[string]map[string]*Migration)

	m := &Migration{
		App:  "testapp",
		Name: "0001_initial",
		Operations: []Operation{
			CreateModel{
				Name: "Author",
				Fields: []*orm.Field{
					{Name: "ID", Column: "id", Type: orm.BigAutoField, PrimaryKey: true},
					{Name: "Name", Column: "name", Type: orm.CharField, Options: orm.FieldOptions{MaxLength: 50}},
				},
				Meta: &orm.Meta{DbTable: "test_author"},
			},
		},
	}
	RegisterMigration(m)

	ctx := context.Background()
	executor, err := NewMigrationExecutor("default")
	if err != nil {
		t.Fatalf("Executor init failed: %v", err)
	}

	// 1. Plan
	plan, err := executor.Plan(ctx, "")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	if len(plan) != 1 {
		t.Fatalf("Expected 1 migration to run, got %d", len(plan))
	}

	// 2. Execute
	err = executor.Execute(ctx, plan)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 3. Verify applied
	applied, err := executor.GetAppliedMigrations(ctx)
	if err != nil {
		t.Fatalf("Failed to fetch applied: %v", err)
	}
	if !applied["testapp.0001_initial"] {
		t.Errorf("Migration not recorded as applied")
	}

	// 4. Second Plan (should be empty)
	plan2, _ := executor.Plan(ctx, "")
	if len(plan2) != 0 {
		t.Errorf("Expected 0 migrations after applying, got %d", len(plan2))
	}
}
