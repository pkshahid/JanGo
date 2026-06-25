package response

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkshahid/JanGo/core/settings"
	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/template"
)

func TestResponses(t *testing.T) {
	settings.Configure(settings.Settings{
		DEBUG:        true,
		STATIC_URL:   "/static/",
		MEDIA_URL:    "/media/",
		SECRET_KEY:   "secret",
		ROOT_URLCONF: "test",
	})

	tmpDir, _ := os.MkdirTemp("", "godjango_responses")
	defer os.RemoveAll(tmpDir)

	tmplPath := filepath.Join(tmpDir, "test.html")
	os.WriteFile(tmplPath, []byte(`Hello {{ name }}! URL: {{ STATIC_URL }}`), 0644)

	engine := template.NewEngine([]string{tmpDir}, false)
	rawReq := httptest.NewRequest("GET", "/", nil)
	req := godjangohttp.NewRequest(rawReq)

	ctxData := map[string]any{"name": "World"}

	// Test RenderToString
	res, err := RenderToString(engine, "test.html", ctxData, req)
	if err != nil {
		t.Fatalf("RenderToString error: %v", err)
	}
	if res != "Hello World! URL: /static/" {
		t.Errorf("RenderToString mismatch, got: %s", res)
	}

	// Test Render (HttpResponse)
	httpResp := Render(engine, req, "test.html", ctxData)
	if httpResp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", httpResp.StatusCode)
	}

	// Test TemplateResponse (lazy)
	lazyResp := NewTemplateResponse(engine, req, "test.html", ctxData)
	recorder := httptest.NewRecorder()

	// Nothing rendered yet
	if lazyResp.rendered {
		t.Errorf("Expected not rendered")
	}

	lazyResp.Write(recorder)

	if !lazyResp.rendered {
		t.Errorf("Expected rendered to be true")
	}
	if recorder.Body.String() != "Hello World! URL: /static/" {
		t.Errorf("TemplateResponse body mismatch, got: %s", recorder.Body.String())
	}

	// Test SimpleTemplateResponse (no request / processors)
	simpleResp := NewSimpleTemplateResponse(engine, "test.html", ctxData)
	recorderSimple := httptest.NewRecorder()
	simpleResp.Write(recorderSimple)

	// STATIC_URL shouldn't be resolved because no processors ran
	if recorderSimple.Body.String() != "Hello World! URL: " {
		t.Errorf("SimpleTemplateResponse body mismatch, got: %s", recorderSimple.Body.String())
	}
}
