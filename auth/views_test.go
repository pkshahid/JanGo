package auth

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"strings"

	godjangohttp "github.com/godjango/godjango/http"
	"github.com/godjango/godjango/orm"
)

func TestAuthViews(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&AbstractUser{}, &Group{}, &Permission{})

	// Test LoginView GET Redirect (already logged in)
	loginView := NewLoginView()
	req := godjangohttp.NewRequest(httptest.NewRequest("GET", "/login", nil))
	req.User = &AbstractUser{UsernameStr: "admin", IsActiveVal: true}

	resp := loginView.Dispatch(req)
	if redirect, ok := resp.(*godjangohttp.RedirectResponse); !ok || redirect.URL != "/" {
		t.Errorf("Expected redirect to /, got %v", resp)
	}

	// Test LogoutView
	req2 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/logout", nil))
	req2.Session = &mockSession{data: make(map[string]any)}

	resp2 := LogoutView(req2)
	hr2 := resp2.(*godjangohttp.HttpResponse)
	if hr2.StatusCode != http.StatusOK {
		t.Errorf("LogoutView should render 200 OK by default")
	}

	// Test PasswordChangeView Unauthorized
	pwView := NewPasswordChangeView()
	req3 := godjangohttp.NewRequest(httptest.NewRequest("GET", "/password_change", nil))
	resp3 := pwView.Dispatch(req3)

	if redirect, ok := resp3.(*godjangohttp.RedirectResponse); !ok || !strings.Contains(redirect.URL, "/login/") {
		t.Errorf("PasswordChangeView should redirect unauthorized users, got %v", resp3)
	}
}
