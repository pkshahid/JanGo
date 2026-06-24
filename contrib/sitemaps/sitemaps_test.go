package sitemaps

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStaticSitemap(t *testing.T) {
	sm := &StaticSitemap{
		URLs:            []string{"/", "/about", "/contact"},
		ChangeFrequency: Monthly,
		PriorityValue:   0.8,
	}

	items := sm.Items()
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	loc := sm.Location(items[0])
	if loc != "/" {
		t.Errorf("expected '/', got %q", loc)
	}

	cf := sm.Changefreq(items[0])
	if cf != Monthly {
		t.Errorf("expected Monthly, got %v", cf)
	}

	p := sm.Priority(items[0])
	if p != 0.8 {
		t.Errorf("expected 0.8, got %f", p)
	}
}

func TestRegistry(t *testing.T) {
	reg := NewRegistry("https://example.com")
	reg.Register("static", &StaticSitemap{
		URLs: []string{"/", "/about"},
	})

	urlSet, err := reg.GenerateURLSet("static")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(urlSet.URLs) != 2 {
		t.Fatalf("expected 2 URLs, got %d", len(urlSet.URLs))
	}

	if urlSet.URLs[0].Loc != "https://example.com/" {
		t.Errorf("expected full URL, got %q", urlSet.URLs[0].Loc)
	}
}

func TestGenerateIndex(t *testing.T) {
	reg := NewRegistry("https://example.com")
	reg.Register("pages", &StaticSitemap{URLs: []string{"/"}})
	reg.Register("blog", &StaticSitemap{URLs: []string{"/blog"}})

	index := reg.GenerateIndex()
	if len(index.Sitemaps) != 2 {
		t.Fatalf("expected 2 sitemaps in index, got %d", len(index.Sitemaps))
	}
}

func TestHandler(t *testing.T) {
	reg := NewRegistry("https://example.com")
	reg.Register("pages", &StaticSitemap{URLs: []string{"/", "/about"}})

	handler := reg.Handler()
	req := httptest.NewRequest("GET", "/sitemap.xml", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "<urlset") {
		t.Error("response should contain urlset element")
	}
	if !strings.Contains(body, "https://example.com/") {
		t.Error("response should contain full URL")
	}
}

func TestSectionHandler(t *testing.T) {
	reg := NewRegistry("https://example.com")
	reg.Register("blog", &StaticSitemap{URLs: []string{"/blog/post-1", "/blog/post-2"}})

	handler := reg.SectionHandler("blog")
	req := httptest.NewRequest("GET", "/sitemap-blog.xml", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "/blog/post-1") {
		t.Error("response should contain blog post URL")
	}
}

func TestSectionHandlerNotFound(t *testing.T) {
	reg := NewRegistry("https://example.com")
	handler := reg.SectionHandler("nonexistent")
	req := httptest.NewRequest("GET", "/sitemap-nonexistent.xml", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestURLXMLOutput(t *testing.T) {
	now := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	urlSet := &URLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs: []URL{
			{
				Loc:        "https://example.com/",
				LastMod:    &now,
				ChangeFreq: Daily,
				Priority:   1.0,
			},
		},
	}

	data, err := xml.MarshalIndent(urlSet, "", "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "https://example.com/") {
		t.Error("XML should contain the URL")
	}
	if !strings.Contains(output, "daily") {
		t.Error("XML should contain changefreq")
	}
}

func TestMultipleSitemapsIndex(t *testing.T) {
	reg := NewRegistry("https://example.com")
	reg.Register("pages", &StaticSitemap{URLs: []string{"/"}})
	reg.Register("blog", &StaticSitemap{URLs: []string{"/blog"}})

	handler := reg.Handler()
	req := httptest.NewRequest("GET", "/sitemap.xml", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "<sitemapindex") {
		t.Error("multiple sitemaps should produce an index")
	}
}
