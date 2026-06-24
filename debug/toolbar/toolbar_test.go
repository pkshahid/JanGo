package toolbar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pkshahid/JanGo/core/settings"
	godjangohttp "github.com/pkshahid/JanGo/http"
)

func TestMiddleware(t *testing.T) {
	settings.Configure(settings.Settings{
		SECRET_KEY:   "test",
		ROOT_URLCONF: "test",
		DEBUG:        true,
		INTERNAL_IPS: []string{"127.0.0.1"},
	})

	nextHandler := func(req *godjangohttp.Request) godjangohttp.Response {
		r := godjangohttp.NewHttpResponse("<html><body>Hello World</body></html>", http.StatusOK)
		r.Headers.Set("Content-Type", "text/html")
		return r
	}

	middleware := DebugToolbarMiddleware(nextHandler)

	rawReq := httptest.NewRequest("GET", "/", nil)
	rawReq.RemoteAddr = "127.0.0.1:12345"
	req := godjangohttp.NewRequest(rawReq)
	req.Context = context.Background()

	resp := middleware(req)

	rr := httptest.NewRecorder()
	resp.Write(rr)

	body := rr.Body.String()
	if !strings.Contains(body, "id=\"djDebug\"") {
		t.Errorf("Expected toolbar injected in response, got %s", body)
	}

	// Test the endpoint
	// Extract the UUID
	idx := strings.Index(body, "data-store-id=\"")
	if idx == -1 {
		t.Fatalf("Store ID not found in body")
	}
	uuidStart := idx + 15
	uuidEnd := strings.Index(body[uuidStart:], "\"")
	storeID := body[uuidStart : uuidStart+uuidEnd]

	rawReq2 := httptest.NewRequest("GET", "/djdt/render_panel/?store_id="+storeID+"&panel_id=Timer", nil)
	req2 := godjangohttp.NewRequest(rawReq2)

	resp2 := RenderPanel(req2)
	rr2 := httptest.NewRecorder()
	resp2.Write(rr2)

	if rr2.Code != 200 {
		t.Errorf("Expected 200 OK from render panel, got %d", rr2.Code)
	}

	if !strings.Contains(rr2.Body.String(), "Request Time") {
		t.Errorf("Expected Timer panel content, got %s", rr2.Body.String())
	}
}
