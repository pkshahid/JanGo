package admin

import (
	"strings"
	"testing"
	"net/http"
	"net/http/httptest"
	"io"

	godjangohttp "github.com/godjango/godjango/http"
	"github.com/godjango/godjango/auth"
	"github.com/godjango/godjango/orm"
)

type TestAdminModel struct {
	orm.Model
	Name string `gd:"CharField"`
}

type mockUser struct {
	authenticated bool
	staff         bool
	username      string
}

func (u *mockUser) ID() uint64 { return 1 }
func (u *mockUser) Username() string { return u.username }
func (u *mockUser) Email() string { return "" }
func (u *mockUser) IsAuthenticated() bool { return u.authenticated }
func (u *mockUser) IsAnonymous() bool { return !u.authenticated }
func (u *mockUser) IsActive() bool { return true }
func (u *mockUser) IsStaff() bool { return u.staff }
func (u *mockUser) IsSuperuser() bool { return u.staff }
func (u *mockUser) HasPerm(perm string) bool { return true }
func (u *mockUser) HasPerms(perms []string) bool { return true }
func (u *mockUser) HasModulePerm(appLabel string) bool { return true }

func TestAdminSite(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&TestAdminModel{})

	site := NewAdminSite("testadmin")

	err := site.Register(&TestAdminModel{}, nil)
	if err != nil {
		t.Fatalf("Failed to register model: %v", err)
	}

	urls := site.URLs()
	if len(urls.Patterns) < 5 {
		t.Fatalf("Expected URL patterns to be generated")
	}

	// Setup mock request
	rawReq := httptest.NewRequest("GET", "/admin/", nil)
	req := godjangohttp.NewRequest(rawReq)

	// Test Unauthorized access
	req.User = &mockUser{authenticated: false}

	resp := site.index(req)
	// Since we called `site.index` directly, we bypass the decorator in tests.
	// We need to test the decorator.

	wrappedView := site.adminView(site.index)
	resp2 := wrappedView(req)
	if rr, ok := resp2.(*godjangohttp.RedirectResponse); !ok || !strings.Contains(rr.URL, "/login/") {
		t.Errorf("Expected login redirect, got %T", resp2)
	}

	// Test Non-staff access
	req.User = &mockUser{authenticated: true, staff: false}
	resp3 := wrappedView(req)
	hr3 := resp3.(*godjangohttp.HttpResponse)
	if hr3.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden, got %d", hr3.StatusCode)
	}

	// Test Staff access
	req.User = &mockUser{authenticated: true, staff: true, username: "admin_test"}
	resp4 := wrappedView(req)
	hr4 := resp4.(*godjangohttp.HttpResponse)

	if hr4.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", hr4.StatusCode)
	}

	bodyBytes, _ := io.ReadAll(hr4.Body)
	bodyStr := string(bodyBytes)

	if !strings.Contains(bodyStr, "GoDjango administration") {
		t.Errorf("Missing header in index: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "TestAdminModel") {
		t.Errorf("Missing registered model in index: %s", bodyStr)
	}

	// Test static serving
	reqStatic := godjangohttp.NewRequest(httptest.NewRequest("GET", "/admin/static/css/base.css", nil))
	reqStatic.ResolverMatch = &godjangohttp.ResolverMatch{Kwargs: map[string]any{"path": "css/base.css"}}

	staticResp := site.ServeStatic(reqStatic).(*godjangohttp.HttpResponse)
	if staticResp.Headers.Get("Content-Type") != "text/css" {
		t.Errorf("Expected static CSS, got type %s", staticResp.Headers.Get("Content-Type"))
	}

	staticBody, _ := io.ReadAll(staticResp.Body)
	if !strings.Contains(string(staticBody), "body { font-family:") {
		t.Errorf("Static body incorrect")
	}
}
