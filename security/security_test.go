package security

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/godjango/godjango/core/settings"
	godjangohttp "github.com/godjango/godjango/http"
)

func TestGenerateSecretKey(t *testing.T) {
	key, err := GenerateSecretKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	if len(key) != 50 {
		t.Errorf("Expected length 50, got %d", len(key))
	}

	key2, _ := GenerateSecretKey()
	if key == key2 {
		t.Errorf("Keys should be random")
	}
}

func TestCSPMiddleware(t *testing.T) {
	m := NewCSPMiddleware()

	handler := m.Process(func(req *godjangohttp.Request) godjangohttp.Response {
		if req.META["CSP_NONCE"] == "" {
			t.Errorf("Nonce missing from META")
		}
		return godjangohttp.NewHttpResponse("OK", 200)
	})

	r, _ := http.NewRequest("GET", "/", nil)
	req := godjangohttp.NewRequest(r)
	resp := handler(req).(*godjangohttp.HttpResponse)

	csp := resp.Headers.Get("Content-Security-Policy")
	if csp == "" {
		t.Errorf("Expected CSP header")
	}
	if !strings.Contains(csp, "default-src 'self'") {
		t.Errorf("Missing default-src")
	}
	if !strings.Contains(csp, "'nonce-") {
		t.Errorf("Missing nonce in script-src")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	// 2 tokens per 100ms
	m := NewRateLimitMiddleware(2, 2, 100*time.Millisecond)

	handler := m.Process(func(req *godjangohttp.Request) godjangohttp.Response {
		return godjangohttp.NewHttpResponse("OK", 200)
	})

	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "127.0.0.1:1234"
	req := godjangohttp.NewRequest(r)

	resp1 := handler(req).(*godjangohttp.HttpResponse)
	if resp1.StatusCode != 200 {
		t.Errorf("Expected 200 for 1st request, got %d", resp1.StatusCode)
	}

	resp2 := handler(req).(*godjangohttp.HttpResponse)
	if resp2.StatusCode != 200 {
		t.Errorf("Expected 200 for 2nd request, got %d", resp2.StatusCode)
	}

	resp3 := handler(req).(*godjangohttp.HttpResponse)
	if resp3.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected 429 for 3rd request, got %d", resp3.StatusCode)
	}

	// Wait for refill
	time.Sleep(150 * time.Millisecond)

	resp4 := handler(req).(*godjangohttp.HttpResponse)
	if resp4.StatusCode != 200 {
		t.Errorf("Expected 200 after refill, got %d", resp4.StatusCode)
	}
}

func TestAllowedHostsMiddleware(t *testing.T) {
	m := NewAllowedHostsMiddleware()

	err := settings.Configure(settings.Settings{
		SECRET_KEY:    "test-secret-key-1234567890",
		ROOT_URLCONF:  "test",
		DEBUG:         false,
		ALLOWED_HOSTS: []string{"example.com", ".example.net"},
	})
	if err != nil && err.Error() != "settings are already configured" && settings.Get() == nil {
		t.Fatalf("Failed to configure settings: %v", err)
	}

	handler := m.Process(func(req *godjangohttp.Request) godjangohttp.Response {
		return godjangohttp.NewHttpResponse("OK", 200)
	})

	tests := []struct {
		host string
		code int
	}{
		{"example.com", 200},
		{"example.com:8080", 200}, // test port stripping
		{"sub.example.net", 200},
		{"example.net", 200},
		{"invalid.com", 400},
	}

	for _, tt := range tests {
		r, _ := http.NewRequest("GET", "/", nil)
		r.Host = tt.host
		req := godjangohttp.NewRequest(r)

		resp := handler(req).(*godjangohttp.HttpResponse)
		if resp.StatusCode != tt.code {
			t.Errorf("Host %s expected %d, got %d", tt.host, tt.code, resp.StatusCode)
		}
	}
}
