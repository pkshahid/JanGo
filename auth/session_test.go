package auth

import (
	"context"
	"testing"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"net/http/httptest"
)

// Mock session map for testing
type mockSession struct {
	data     map[string]any
	modified bool
}

func (m *mockSession) Get(key string) (any, bool) {
	v, ok := m.data[key]
	return v, ok
}
func (m *mockSession) Set(key string, val any)         { m.data[key] = val; m.modified = true }
func (m *mockSession) Delete(key string)               { delete(m.data, key); m.modified = true }
func (m *mockSession) Clear()                          { m.data = make(map[string]any); m.modified = true }
func (m *mockSession) SessionKey() string              { return "test-session-key" }
func (m *mockSession) IsModified() bool                { return m.modified }
func (m *mockSession) Flush(_ context.Context) error   { m.Clear(); return nil }
func (m *mockSession) CycleKey(_ context.Context) error { return nil }
func (m *mockSession) Save(_ context.Context) error    { return nil }

func TestSessionAuth(t *testing.T) {
	// 1. Setup Request and Session
	rawReq := httptest.NewRequest("GET", "/", nil)
	req := godjangohttp.NewRequest(rawReq)
	req.Session = &mockSession{data: make(map[string]any)}

	// 2. Setup user mock
	user := &AbstractUser{
		UsernameStr: "alice",
	}
	user.Model.ID = 42 // Assuming Model provides ID

	// 3. Test Login
	err := Login(req, user)
	if err != nil {
		t.Fatalf("Login error: %v", err)
	}

	val, ok := req.Session.Get("_auth_user_id")
	if !ok || val != "42" {
		t.Errorf("Session _auth_user_id not set correctly, got %v", val)
	}

	// 4. Test Logout
	err = Logout(req)
	if err != nil {
		t.Fatalf("Logout error: %v", err)
	}

	val, ok = req.Session.Get("_auth_user_id")
	if ok && val != nil {
		t.Errorf("Session _auth_user_id should be deleted")
	}

	if !req.User.IsAnonymous() {
		t.Errorf("Request user should be anonymous after logout")
	}
}
