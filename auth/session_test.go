package auth

import (
	"testing"

	godjangohttp "github.com/godjango/godjango/http"
	"net/http/httptest"
)

// Mock session map for testing
type mockSession struct {
	data map[string]any
}
func (m *mockSession) Get(key string) any { return m.data[key] }
func (m *mockSession) Set(key string, val any) { m.data[key] = val }
func (m *mockSession) Delete(key string) { delete(m.data, key) }

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

	if req.Session.Get("_auth_user_id") != "42" {
		t.Errorf("Session _auth_user_id not set correctly")
	}

	// 4. Test Logout
	err = Logout(req)
	if err != nil {
		t.Fatalf("Logout error: %v", err)
	}

	if req.Session.Get("_auth_user_id") != nil {
		t.Errorf("Session _auth_user_id should be deleted")
	}

	if !req.User.IsAnonymous() {
		t.Errorf("Request user should be anonymous after logout")
	}
}
