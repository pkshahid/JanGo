package auth

import (
	"testing"

	"github.com/pkshahid/JanGo/auth/hashers"
	"github.com/pkshahid/JanGo/core/settings"
	"github.com/pkshahid/JanGo/orm"
	"github.com/pkshahid/JanGo/orm/backends"
	"github.com/pkshahid/JanGo/orm/queryset"
)

func TestModelBackend(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&AbstractUser{}, &Group{}, &Permission{})

	// Set up a real SQLite in-memory database so QuerySet.Get() works.
	backend := &backends.SQLiteBackend{}
	if err := backend.Connect(settings.DatabaseConfig{Name: ":memory:"}); err != nil {
		t.Fatalf("failed to connect test DB: %v", err)
	}
	backend.DB().SetMaxOpenConns(1)
	backends.ClearBackends()
	backends.RegisterBackendInstance("default", backend)
	t.Cleanup(func() {
		backend.Close()
		backends.ClearBackends()
	})

	// Create tables for all registered models.
	for _, m := range orm.AllModels() {
		if err := backend.SchemaEditor().CreateTable(m); err != nil {
			t.Fatalf("failed to create table for %s: %v", m.Name, err)
		}
	}

	password := "my_secure_pass"
	encoded := hashers.MakePassword(password)

	mb := &ModelBackend{}

	// User doesn't exist in the DB, so Authenticate should return DoesNotExist.
	_, err := mb.Authenticate("admin", password)
	if err == nil || err.Error() != "user not found or inactive: orm: DoesNotExist" {
		t.Errorf("Expected DoesNotExist error because user is not in DB, got: %v", err)
	}

	// Insert a user and verify authentication works.
	user := &AbstractUser{
		UsernameStr: "admin",
		Password:    encoded,
		IsActiveVal: true,
	}
	qs := queryset.NewQuerySet[AbstractUser]()
	if err := qs.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	authed, err := mb.Authenticate("admin", password)
	if err != nil {
		t.Fatalf("Authenticate error: %v", err)
	}
	if authed == nil {
		t.Fatal("Expected authenticated user, got nil")
	}

	// Wrong password should fail.
	_, err = mb.Authenticate("admin", "wrong_password")
	if err == nil {
		t.Error("Expected error for wrong password")
	}
}
