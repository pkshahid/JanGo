package backends

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/godjango/godjango/core/settings"
	"github.com/godjango/godjango/orm"
)

type DummyModel struct {
	orm.Model
	Name string `gd:"CharField,max_length=50,unique=true"`
	Age  int    `gd:"IntegerField"`
}

func (d *DummyModel) ModelMeta() *orm.Meta {
	return &orm.Meta{DbTable: "dummy_model"}
}

func setupTestDB(t *testing.T) string {
	tmpDir, _ := os.MkdirTemp("", "godjango_db_test")
	dbPath := filepath.Join(tmpDir, "test.sqlite3")

	settings.Configure(settings.Settings{
		DATABASES: map[string]settings.DatabaseConfig{
			"default": {
				Engine:   "sqlite",
				Name:     dbPath,
				MAX_CONN: 5,
				MAX_IDLE: 2,
				CONN_MAX_LIFETIME: 300,
			},
		},
	})

	err := Init()
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}

	return dbPath
}

func TestSQLiteBackend(t *testing.T) {
	dbPath := setupTestDB(t)
	defer os.Remove(dbPath)
	defer Close()

	backend, err := GetBackend("default")
	if err != nil {
		t.Fatalf("Failed to get backend: %v", err)
	}

	if backend.DatabaseName() != "sqlite" {
		t.Errorf("Expected sqlite database name")
	}

	features := backend.Features()
	if !features.SupportsJSON || !features.SupportsReturning {
		t.Errorf("Features mismatch: %+v", features)
	}

	// Test pooling configs
	stats := backend.DB().Stats()
	if stats.MaxOpenConnections != 5 {
		t.Errorf("Expected MaxOpenConns 5, got %d", stats.MaxOpenConnections)
	}

	orm.ClearRegistry()
	orm.Register(&DummyModel{})
	info, _ := orm.GetModelInfo(&DummyModel{})

	editor := backend.SchemaEditor()

	// Create Table
	if err := editor.CreateTable(info); err != nil {
		t.Fatalf("CreateTable failed: %v", err)
	}

	// Add/Remove column
	newField := &orm.Field{Name: "NewCol", Column: "new_col", Type: orm.CharField, Options: orm.FieldOptions{MaxLength: 10}}
	if err := editor.AddColumn(info, newField); err != nil {
		t.Fatalf("AddColumn failed: %v", err)
	}
	if err := editor.RemoveColumn(info, "new_col"); err != nil {
		t.Fatalf("RemoveColumn failed: %v", err)
	}

	// Create/Delete index
	idx := orm.Index{Name: "test_idx", Fields: []string{"age"}}
	if err := editor.CreateIndex(info, idx); err != nil {
		t.Fatalf("CreateIndex failed: %v", err)
	}
	if err := editor.DeleteIndex(info, "test_idx"); err != nil {
		t.Fatalf("DeleteIndex failed: %v", err)
	}

	// Delete Table
	if err := editor.DeleteTable(info); err != nil {
		t.Fatalf("DeleteTable failed: %v", err)
	}
}

func TestAtomicTransactions(t *testing.T) {
	dbPath := setupTestDB(t)
	defer os.Remove(dbPath)
	defer Close()

	backend, _ := GetBackend("default")
	editor := backend.SchemaEditor()
	orm.ClearRegistry()
	orm.Register(&DummyModel{})
	info, _ := orm.GetModelInfo(&DummyModel{})
	editor.CreateTable(info)

	ctx := context.Background()

	// 1. Success case with OnCommit
	var hookRan bool
	err := Atomic(ctx, "default", func(txCtx context.Context, tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO dummy_model (name, age) VALUES (?, ?)", "Alice", 30)
		if err != nil {
			return err
		}

		OnCommit(txCtx, func() {
			hookRan = true
		})

		return nil
	})

	if err != nil {
		t.Fatalf("Atomic failed: %v", err)
	}
	if !hookRan {
		t.Errorf("OnCommit hook did not run")
	}

	var count int
	backend.DB().QueryRow("SELECT COUNT(*) FROM dummy_model").Scan(&count)
	if count != 1 {
		t.Errorf("Expected 1 row, got %d", count)
	}

	// 2. Rollback case
	hookRan = false
	err = Atomic(ctx, "default", func(txCtx context.Context, tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO dummy_model (name, age) VALUES (?, ?)", "Bob", 25)
		if err != nil {
			return err
		}

		OnCommit(txCtx, func() {
			hookRan = true
		})

		return fmt.Errorf("intentional rollback")
	})

	if err == nil || err.Error() != "intentional rollback" {
		t.Errorf("Expected intentional rollback error")
	}
	if hookRan {
		t.Errorf("OnCommit hook ran despite rollback")
	}

	backend.DB().QueryRow("SELECT COUNT(*) FROM dummy_model").Scan(&count)
	if count != 1 {
		t.Errorf("Expected 1 row after rollback, got %d", count)
	}

	// 3. Nested Savepoints
	err = Atomic(ctx, "default", func(txCtx context.Context, tx *sql.Tx) error {
		tx.Exec("INSERT INTO dummy_model (name, age) VALUES (?, ?)", "Charlie", 40)

		// Nested block
		Atomic(txCtx, "default", func(nestedCtx context.Context, ntx *sql.Tx) error {
			ntx.Exec("INSERT INTO dummy_model (name, age) VALUES (?, ?)", "Dave", 50)
			return fmt.Errorf("nested rollback")
		})

		return nil
	})

	backend.DB().QueryRow("SELECT COUNT(*) FROM dummy_model").Scan(&count)
	if count != 2 {
		t.Errorf("Expected 2 rows (Alice, Charlie), got %d. Dave should be rolled back.", count)
	}

	var name string
	backend.DB().QueryRow("SELECT name FROM dummy_model WHERE age = 40").Scan(&name)
	if name != "Charlie" {
		t.Errorf("Expected Charlie to exist")
	}
}
