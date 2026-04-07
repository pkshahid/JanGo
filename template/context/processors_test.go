package context

import (
	"net/http/httptest"
	"testing"

	"github.com/godjango/godjango/core/settings"
	godjangohttp "github.com/godjango/godjango/http"
)

type mockUser struct {
	authenticated bool
	username      string
}

func (u *mockUser) IsAuthenticated() bool { return u.authenticated }
func (u *mockUser) GetUsername() string   { return u.username }
func (u *mockUser) HasPerm(perm string) bool { return true }

func TestContextProcessors(t *testing.T) {
	settings.Configure(settings.Settings{
		DEBUG:      true,
		STATIC_URL: "/static/",
		MEDIA_URL:  "/media/",
		SECRET_KEY: "secret",
		ROOT_URLCONF: "test",
	})

	rawReq := httptest.NewRequest("GET", "/", nil)
	req := godjangohttp.NewRequest(rawReq)
	req.User = &mockUser{authenticated: true, username: "admin"}
	req.META["MESSAGES"] = "Hello Flash"
	req.META["CSRF_TOKEN"] = "testtoken"

	processors := []ProcessorFunc{
		RequestContextProcessor,
		AuthContextProcessor,
		MessagesContextProcessor,
		StaticContextProcessor,
		DebugContextProcessor,
		CsrfContextProcessor,
	}

	base := map[string]any{
		"custom_var": "custom_val",
		"user":       "should_not_overwrite", // base overrides processor
	}

	ctx := BuildRequestContext(req, processors, base)
	flat := ctx.Flatten()

	// Verify Base
	if flat["custom_var"] != "custom_val" {
		t.Errorf("Expected custom_var=custom_val")
	}

	// Verify Overwrite protection
	if flat["user"] != "should_not_overwrite" {
		t.Errorf("Expected base user var to take precedence, got %v", flat["user"])
	}

	// Verify Request
	if flat["request"] != req {
		t.Errorf("Expected request object")
	}

	// Verify Messages
	if flat["messages"] != "Hello Flash" {
		t.Errorf("Expected Hello Flash messages")
	}

	// Verify Static
	if flat["STATIC_URL"] != "/static/" {
		t.Errorf("Expected /static/")
	}

	// Verify Debug
	if flat["debug"] != true {
		t.Errorf("Expected debug=true")
	}

	// Verify CSRF
	if flat["csrf_token"] != "testtoken" {
		t.Errorf("Expected testtoken csrf")
	}
}
