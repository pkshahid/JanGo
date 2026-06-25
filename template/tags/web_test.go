package tags

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pkshahid/JanGo/core/settings"
	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/http/urls"
	godjango "github.com/pkshahid/JanGo/template"
)

func TestWebTags(t *testing.T) {
	// Re-initialize settings safely for testing
	settings.Configure(settings.Settings{
		STATIC_URL:   "/static/",
		SECRET_KEY:   "secret",
		ROOT_URLCONF: "test",
	})

	lib := godjango.NewLibrary()
	RegisterWebTags(lib)

	engine := godjango.NewEngine(nil, false)
	engine.AddBuiltin(lib)

	// URL tag setup router
	router := urls.GetGlobalRouter()
	router.Add(urls.Path("test/<int:id>/", func(req *godjangohttp.Request) godjangohttp.Response { return nil }, "test_view", nil))

	inputUrl := `{% url "test_view" id=42 %}`
	tmplUrl, _ := engine.FromString(inputUrl)
	resUrl, _ := tmplUrl.Render(godjango.NewContext(nil))

	if resUrl != "/test/42/" {
		t.Errorf("URL tag mismatch, got: %s", resUrl)
	}

	// Static tag
	inputStatic := `{% static "css/style.css" %}`
	tmplStatic, _ := engine.FromString(inputStatic)
	resStatic, _ := tmplStatic.Render(godjango.NewContext(nil))

	if resStatic != "/static/css/style.css" {
		t.Errorf("Static tag mismatch, got: %s", resStatic)
	}

	// CSRF token tag
	inputCsrf := `{% csrf_token %}`
	tmplCsrf, _ := engine.FromString(inputCsrf)

	req := godjangohttp.NewRequest(httptest.NewRequest("GET", "/", nil))
	req.META["CSRF_TOKEN"] = "abc123token"
	ctx := godjango.NewContext(map[string]any{"request": req})

	resCsrf, _ := tmplCsrf.Render(ctx)
	if !strings.Contains(resCsrf, `value="abc123token"`) {
		t.Errorf("CSRF token tag mismatch, got: %s", resCsrf)
	}
}
