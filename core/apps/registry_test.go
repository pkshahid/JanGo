package apps

import (
	"strings"
	"testing"
)

type mockApp struct {
	name  string
	label string
	path  string
	ready bool
}

func (m *mockApp) Name() string        { return m.name }
func (m *mockApp) Label() string       { return m.label }
func (m *mockApp) Path() string        { return m.path }
func (m *mockApp) Models() []ModelInfo { return nil }
func (m *mockApp) Ready()              { m.ready = true }

func TestAppRegistry(t *testing.T) {
	t.Cleanup(Reset)

	app1 := &mockApp{name: "github.com/my/app1", label: "app1"}
	app2 := &mockApp{name: "github.com/my/app2"} // label defaults to name

	Register(app1)
	Register(app2)

	err := Setup([]string{"github.com/my/app1", "github.com/my/app2"})
	if err != nil {
		t.Fatalf("unexpected error during setup: %v", err)
	}

	if !app1.ready {
		t.Errorf("app1.Ready() was not called")
	}
	if !app2.ready {
		t.Errorf("app2.Ready() was not called")
	}

	// Test Get by label
	gotApp1, err := Get("app1")
	if err != nil || gotApp1.Name() != "github.com/my/app1" {
		t.Errorf("Get(app1) failed: err=%v, app=%v", err, gotApp1)
	}

	gotApp2, err := Get("github.com/my/app2") // defaults to name
	if err != nil || gotApp2.Name() != "github.com/my/app2" {
		t.Errorf("Get(github.com/my/app2) failed: err=%v, app=%v", err, gotApp2)
	}

	// Test All
	allApps := All()
	if len(allApps) != 2 || allApps[0].Name() != "github.com/my/app1" || allApps[1].Name() != "github.com/my/app2" {
		t.Errorf("All() returned incorrect order or apps: %v", allApps)
	}

	// Test IsInstalled
	if !IsInstalled("github.com/my/app1") {
		t.Errorf("IsInstalled(github.com/my/app1) should be true")
	}
	if IsInstalled("github.com/my/app3") {
		t.Errorf("IsInstalled(github.com/my/app3) should be false")
	}
}

func TestSetupErrors(t *testing.T) {
	t.Run("missing app", func(t *testing.T) {
		Reset()
		err := Setup([]string{"missing.app"})
		if err == nil || !strings.Contains(err.Error(), "has not been registered") {
			t.Errorf("expected error for missing app, got: %v", err)
		}
	})

	t.Run("duplicate label", func(t *testing.T) {
		Reset()
		Register(&mockApp{name: "app1", label: "dup"})
		Register(&mockApp{name: "app2", label: "dup"})
		err := Setup([]string{"app1", "app2"})
		if err == nil || !strings.Contains(err.Error(), "already in use") {
			t.Errorf("expected error for duplicate label, got: %v", err)
		}
	})

	t.Run("already ready", func(t *testing.T) {
		Reset()
		Register(&mockApp{name: "app1"})
		Setup([]string{"app1"})
		err := Setup([]string{"app1"})
		if err == nil || !strings.Contains(err.Error(), "already populated") {
			t.Errorf("expected error for calling Setup twice, got: %v", err)
		}
	})
}

func TestRegisterPanics(t *testing.T) {
	t.Cleanup(Reset)

	t.Run("empty name", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Register did not panic on empty name")
			}
		}()
		Reset()
		Register(&mockApp{name: ""})
	})

	t.Run("duplicate registration", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Register did not panic on duplicate registration")
			}
		}()
		Reset()
		app := &mockApp{name: "app1"}
		Register(app)
		Register(app) // Should panic
	})
}

func TestUnreadyAccess(t *testing.T) {
	Reset()

	if _, err := Get("app1"); err == nil {
		t.Errorf("Get should fail if registry is not ready")
	}

	if All() != nil {
		t.Errorf("All should return nil if registry is not ready")
	}

	if IsInstalled("app1") != false {
		t.Errorf("IsInstalled should return false if registry is not ready")
	}
}
