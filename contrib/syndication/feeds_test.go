package syndication

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBaseFeed(t *testing.T) {
	feed := &BaseFeed{
		FeedTitle:       "My Blog",
		FeedLink:        "https://example.com",
		FeedDescription: "A test blog",
		FeedItems: []FeedItem{
			{
				Title:       "First Post",
				Link:        "https://example.com/post-1",
				Description: "Hello world",
				Author:      "John",
				PubDate:     time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			},
		},
	}

	if feed.Title() != "My Blog" {
		t.Errorf("unexpected title: %q", feed.Title())
	}
	if feed.FeedType() != RSS2 {
		t.Errorf("expected RSS2 default, got %v", feed.FeedType())
	}
	if len(feed.Items()) != 1 {
		t.Errorf("expected 1 item, got %d", len(feed.Items()))
	}
}

func TestGenerateRSS(t *testing.T) {
	feed := &BaseFeed{
		FeedTitle:       "Test Feed",
		FeedLink:        "https://example.com",
		FeedDescription: "Testing RSS generation",
		FeedItems: []FeedItem{
			{
				Title:       "Item 1",
				Link:        "https://example.com/1",
				Description: "First item",
				PubDate:     time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				Categories:  []string{"tech", "go"},
			},
			{
				Title:       "Item 2",
				Link:        "https://example.com/2",
				Description: "Second item",
			},
		},
	}

	data, err := GenerateRSS(feed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "<rss") {
		t.Error("should contain rss root element")
	}
	if !strings.Contains(output, "Test Feed") {
		t.Error("should contain feed title")
	}
	if !strings.Contains(output, "Item 1") {
		t.Error("should contain item title")
	}
	if !strings.Contains(output, "<category>tech</category>") {
		t.Error("should contain category")
	}
}

func TestGenerateAtom(t *testing.T) {
	feed := &BaseFeed{
		FeedTitle:       "Atom Feed",
		FeedLink:        "https://example.com",
		FeedDescription: "Testing Atom",
		OutputType:      Atom,
		FeedItems: []FeedItem{
			{
				Title:       "Entry 1",
				Link:        "https://example.com/entry-1",
				Description: "First entry",
				Author:      "Alice",
				PubDate:     time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	data, err := GenerateAtom(feed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "<feed") {
		t.Error("should contain feed root element")
	}
	if !strings.Contains(output, "Atom Feed") {
		t.Error("should contain feed title")
	}
	if !strings.Contains(output, "Alice") {
		t.Error("should contain author name")
	}
}

func TestRSSHandler(t *testing.T) {
	feed := &BaseFeed{
		FeedTitle: "Handler Feed",
		FeedLink:  "https://example.com",
		FeedItems: []FeedItem{
			{Title: "Test", Link: "https://example.com/test"},
		},
	}

	handler := Handler(feed)
	req := httptest.NewRequest("GET", "/feed/", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "rss+xml") {
		t.Errorf("expected RSS content type, got %q", ct)
	}
}

func TestAtomHandler(t *testing.T) {
	feed := &BaseFeed{
		FeedTitle:  "Atom Handler",
		FeedLink:   "https://example.com",
		OutputType: Atom,
		FeedItems: []FeedItem{
			{Title: "Test", Link: "https://example.com/test"},
		},
	}

	handler := Handler(feed)
	req := httptest.NewRequest("GET", "/feed/", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "atom+xml") {
		t.Errorf("expected Atom content type, got %q", ct)
	}
}

func TestEmptyFeed(t *testing.T) {
	feed := &BaseFeed{
		FeedTitle: "Empty",
		FeedLink:  "https://example.com",
	}

	data, err := GenerateRSS(feed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(string(data), "Empty") {
		t.Error("should still contain feed title")
	}
}
