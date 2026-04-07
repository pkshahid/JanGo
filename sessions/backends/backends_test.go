package backends

import (
	"context"
	"testing"
	"time"

	"github.com/godjango/godjango/core/settings"
	"github.com/godjango/godjango/orm"
)

func setupTestSettings() {
	settings.Configure(settings.Settings{
		SECRET_KEY: "test_secret_key",
		SESSION_FILE_PATH: "", // Default to os.TempDir()
	})
}

func TestSignAndUnsign(t *testing.T) {
	setupTestSettings()

	data := map[string]any{"user_id": float64(123), "username": "alice"} // json unmarshals numbers to float64

	signed, err := Sign(data)
	if err != nil {
		t.Fatalf("Failed to sign data: %v", err)
	}

	if signed == "" {
		t.Errorf("Signed data is empty")
	}

	unsigned, err := Unsign(signed)
	if err != nil {
		t.Fatalf("Failed to unsign data: %v", err)
	}

	if unsigned["user_id"] != float64(123) {
		t.Errorf("Data mismatch, expected 123, got %v", unsigned["user_id"])
	}
	if unsigned["username"] != "alice" {
		t.Errorf("Data mismatch, expected alice, got %v", unsigned["username"])
	}

	// Tamper test
	tampered := signed + "x"
	_, err = Unsign(tampered)
	if err == nil {
		t.Errorf("Expected signature verification failure")
	}
}

func TestDatabaseBackend(t *testing.T) {
	// Requires ORM to be initialized to some DB.
	// Since we rely on pure unit testing, we'll ensure it returns DoesNotExist cleanly when empty.
	orm.ClearRegistry()
	orm.Register(&SessionModel{})

	backend := &DatabaseBackend{}
	ctx := context.Background()

	_, err := backend.Load(ctx, "missing_key")
	if err == nil {
		t.Errorf("Expected error for missing session")
	}

	// Save runs async, so testing it without a real DB and without waiting is brittle.
	// We'll skip deep integration test of DB save here to avoid race conditions.
}

func TestFileBackend(t *testing.T) {
	setupTestSettings()

	backend := &FileBackend{}
	ctx := context.Background()

	key := "test_file_session"
	data := map[string]any{"color": "blue"}

	err := backend.Save(ctx, key, data, 10)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Wait for async write
	time.Sleep(100 * time.Millisecond)

	loaded, err := backend.Load(ctx, key)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded["color"] != "blue" {
		t.Errorf("Expected blue, got %v", loaded["color"])
	}

	err = backend.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = backend.Load(ctx, key)
	if err == nil {
		t.Errorf("Expected error after deletion")
	}
}

func TestCacheBackend(t *testing.T) {
	backend := &CacheBackend{}
	ctx := context.Background()

	backend.Save(ctx, "key", map[string]any{"val": 1}, 10)

	loaded, err := backend.Load(ctx, "key")
	if err != nil || loaded["val"] != 1 {
		t.Errorf("Cache backend failed")
	}

	backend.Delete(ctx, "key")
	_, err = backend.Load(ctx, "key")
	if err == nil {
		t.Errorf("Expected error after cache delete")
	}
}
