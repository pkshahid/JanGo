package views

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"

	"github.com/godjango/godjango/core/settings"
	godjangohttp "github.com/godjango/godjango/http"
)

func setupTestSettings() {
	s := settings.Settings{
		SECRET_KEY: "secret",
		ROOT_URLCONF: "test",
		DEBUG: true,
		LOGIN_URL: "/login",
	}
	settings.Configure(s)
}

// Mock User
type mockUser struct {
	authenticated bool
	username      string
	perms         []string
}
func (u *mockUser) IsAuthenticated() bool { return u.authenticated }
func (u *mockUser) GetUsername() string   { return u.username }
func (u *mockUser) HasPerm(perm string) bool {
	for _, p := range u.perms {
		if p == perm {
			return true
		}
	}
	return false
}

func TestDecorators(t *testing.T) {
	setupTestSettings()

	baseView := func(req *godjangohttp.Request) godjangohttp.Response {
		return godjangohttp.NewHttpResponse("OK", http.StatusOK)
	}

	// LoginRequired
	loginReqView := LoginRequired(baseView)
	req1 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/secure", nil))
	resp1 := loginReqView(req1)
	if redirect, ok := resp1.(*godjangohttp.RedirectResponse); !ok || redirect.URL != "/login?next=/secure" {
		t.Errorf("Expected redirect to login, got %v", resp1)
	}

	req1.User = &mockUser{authenticated: true}
	if hr := loginReqView(req1).(*godjangohttp.HttpResponse); hr.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for authenticated user")
	}

	// PermissionRequired
	permView := PermissionRequired("edit_post", baseView)
	req2 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/edit", nil))
	req2.User = &mockUser{authenticated: true, perms: []string{"view_post"}}
	if hr := permView(req2).(*godjangohttp.HttpResponse); hr.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for missing perm")
	}

	req2.User = &mockUser{authenticated: true, perms: []string{"edit_post"}}
	if hr := permView(req2).(*godjangohttp.HttpResponse); hr.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for valid perm")
	}

	// RequireHTTPMethods
	postOnlyView := RequirePOST(baseView)
	req3 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/", nil))
	if hr := postOnlyView(req3).(*godjangohttp.HttpResponse); hr.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405 Method Not Allowed")
	}

	req4 := godjangohttp.NewRequest(httptest.NewRequest("POST", "/", nil))
	if hr := postOnlyView(req4).(*godjangohttp.HttpResponse); hr.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for POST")
	}

	// CacheControl
	cacheView := CacheControl(map[string]any{"max-age": 3600, "public": true})(baseView)
	req5 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/", nil))
	resp5 := cacheView(req5).(*godjangohttp.HttpResponse)
	if !strings.Contains(resp5.Headers.Get("Cache-Control"), "max-age=3600") || !strings.Contains(resp5.Headers.Get("Cache-Control"), "public") {
		t.Errorf("Cache-Control header incorrect: %s", resp5.Headers.Get("Cache-Control"))
	}
}

type MyTestView struct {
	BaseView
}
func (v *MyTestView) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	return v.BaseView.Dispatch(req, v)
}
func (v *MyTestView) Get(req *godjangohttp.Request) godjangohttp.Response {
	return godjangohttp.NewHttpResponse("Get Method", http.StatusOK)
}

func TestClassBasedViews(t *testing.T) {
	myView := &MyTestView{}
	viewFunc := AsView(myView)

	// GET method
	req1 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/", nil))
	if hr := viewFunc(req1).(*godjangohttp.HttpResponse); hr.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK")
	}

	// POST method (not implemented)
	req2 := godjangohttp.NewRequest(httptest.NewRequest("POST", "/", nil))
	if hr := viewFunc(req2).(*godjangohttp.HttpResponse); hr.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405 Method Not Allowed")
	}

	// TemplateView
	tmplView := &TemplateView{TemplateName: "test"}
	viewFunc2 := AsView(tmplView)
	req3 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/", nil))
	// In testing without a real engine, Render returns 500 because template "test" doesn't exist
	// But it dispatches correctly.
	resp3 := viewFunc2(req3).(*godjangohttp.HttpResponse)
	if resp3.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected template engine error, got %d", resp3.StatusCode)
	}

	// RedirectView
	redirView := &RedirectView{URL: "/new-page", Permanent: true}
	viewFunc3 := AsView(redirView)
	req4 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/", nil))
	resp4 := viewFunc3(req4)
	if r, ok := resp4.(*godjangohttp.RedirectResponse); !ok || r.URL != "/new-page" || r.Permanent != true {
		t.Errorf("Expected permanent redirect to /new-page, got %v", resp4)
	}
}

func TestExceptions(t *testing.T) {
	setupTestSettings()

	req := godjangohttp.NewRequest(httptest.NewRequest("GET", "/", nil))

	resp1 := PageNotFound(req, nil).(*godjangohttp.HttpResponse)
	if resp1.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404")
	}

	resp2 := ServerError(req).(*godjangohttp.HttpResponse)
	if resp2.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected 500")
	}
}
