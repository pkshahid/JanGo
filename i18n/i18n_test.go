package i18n

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	godjangohttp "github.com/godjango/godjango/http"
)

func TestGettext(t *testing.T) {
	AddTranslation("fr", "Hello", "Bonjour")
	AddTranslation("fr", "Car", "Voiture")

	ctx := Activate(context.Background(), "fr")

	if Gettext(ctx, "Hello") != "Bonjour" {
		t.Errorf("Expected 'Bonjour', got %s", Gettext(ctx, "Hello"))
	}
	if Gettext(ctx, "Missing") != "Missing" {
		t.Errorf("Expected fallback 'Missing'")
	}
}

func TestMiddlewareURLPrefix(t *testing.T) {
	Config.Languages = []LanguagePair{{"en", "English"}, {"es", "Spanish"}}

	m := NewLocaleMiddleware()

	nextCalled := false
	var activeLang string

	handler := m.Process(func(req *godjangohttp.Request) godjangohttp.Response {
		nextCalled = true
		activeLang = GetLanguage(req.Context)
		return godjangohttp.NewHttpResponse("OK", 200)
	})

	req := godjangohttp.NewRequest(&http.Request{
		URL: &url.URL{Path: "/es/about/"},
		Header: make(http.Header),
	})

	handler(req)

	if !nextCalled {
		t.Fatal("Next handler not called")
	}
	if activeLang != "es" {
		t.Errorf("Expected active language 'es', got %s", activeLang)
	}
}

func TestMiddlewareAcceptLanguage(t *testing.T) {
	Config.Languages = []LanguagePair{{"en", "English"}, {"fr", "French"}}

	m := NewLocaleMiddleware()

	handler := m.Process(func(req *godjangohttp.Request) godjangohttp.Response {
		return godjangohttp.NewHttpResponse(GetLanguage(req.Context), 200)
	})

	req := godjangohttp.NewRequest(&http.Request{
		URL: &url.URL{Path: "/about/"},
		Header: http.Header{"Accept-Language": []string{"fr-CH, fr;q=0.9, en;q=0.8"}},
	})

	resp := handler(req).(*godjangohttp.HttpResponse)
	// We might parse 'fr-CH' as 'fr' depending on the matcher
	body := make([]byte, 100)
	n, _ := resp.Body.Read(body)
	resLang := string(body[:n])

	if resLang != "fr" {
		t.Errorf("Expected 'fr' from Accept-Language, got %s", resLang)
	}
}
