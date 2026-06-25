// Package syndication implements Django's syndication framework.
// It provides RSS and Atom feed generation.
package syndication

import (
	"encoding/xml"
	"net/http"
	"time"
)

// FeedType defines the output format.
type FeedType string

const (
	RSS2 FeedType = "rss"
	Atom FeedType = "atom"
)

// FeedItem represents a single item in a feed.
type FeedItem struct {
	Title       string
	Link        string
	Description string
	Author      string
	PubDate     time.Time
	GUID        string
	Categories  []string
}

// Feed defines the interface for generating feeds.
// Equivalent to Django's Feed class.
type Feed interface {
	// Title returns the feed title.
	Title() string
	// Link returns the feed link.
	Link() string
	// Description returns the feed description.
	Description() string
	// Items returns all items in the feed.
	Items() []FeedItem
	// FeedType returns the output format (RSS2 or Atom).
	FeedType() FeedType
}

// BaseFeed provides default implementations for a feed.
type BaseFeed struct {
	FeedTitle       string
	FeedLink        string
	FeedDescription string
	FeedItems       []FeedItem
	OutputType      FeedType
}

func (f *BaseFeed) Title() string       { return f.FeedTitle }
func (f *BaseFeed) Link() string        { return f.FeedLink }
func (f *BaseFeed) Description() string { return f.FeedDescription }
func (f *BaseFeed) Items() []FeedItem   { return f.FeedItems }
func (f *BaseFeed) FeedType() FeedType {
	if f.OutputType != "" {
		return f.OutputType
	}
	return RSS2
}

// RSS2 XML structures
type rssRoot struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	Author      string   `xml:"author,omitempty"`
	PubDate     string   `xml:"pubDate,omitempty"`
	GUID        string   `xml:"guid,omitempty"`
	Categories  []string `xml:"category,omitempty"`
}

// Atom XML structures
type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	XMLNS   string      `xml:"xmlns,attr"`
	Title   string      `xml:"title"`
	Link    atomLink    `xml:"link"`
	ID      string      `xml:"id"`
	Updated string      `xml:"updated"`
	Entries []atomEntry `xml:"entry"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
}

type atomEntry struct {
	Title   string      `xml:"title"`
	Link    atomLink    `xml:"link"`
	ID      string      `xml:"id"`
	Updated string      `xml:"updated"`
	Summary string      `xml:"summary"`
	Author  *atomAuthor `xml:"author,omitempty"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

// GenerateRSS generates RSS 2.0 XML from a feed.
func GenerateRSS(feed Feed) ([]byte, error) {
	root := rssRoot{
		Version: "2.0",
		Channel: rssChannel{
			Title:       feed.Title(),
			Link:        feed.Link(),
			Description: feed.Description(),
		},
	}

	for _, item := range feed.Items() {
		rItem := rssItem{
			Title:       item.Title,
			Link:        item.Link,
			Description: item.Description,
			Author:      item.Author,
			GUID:        item.GUID,
			Categories:  item.Categories,
		}
		if !item.PubDate.IsZero() {
			rItem.PubDate = item.PubDate.Format(time.RFC1123Z)
		}
		if rItem.GUID == "" {
			rItem.GUID = item.Link
		}
		root.Channel.Items = append(root.Channel.Items, rItem)
	}

	return xml.MarshalIndent(root, "", "  ")
}

// GenerateAtom generates Atom XML from a feed.
func GenerateAtom(feed Feed) ([]byte, error) {
	af := atomFeed{
		XMLNS:   "http://www.w3.org/2005/Atom",
		Title:   feed.Title(),
		Link:    atomLink{Href: feed.Link(), Rel: "alternate"},
		ID:      feed.Link(),
		Updated: time.Now().UTC().Format(time.RFC3339),
	}

	for _, item := range feed.Items() {
		entry := atomEntry{
			Title:   item.Title,
			Link:    atomLink{Href: item.Link},
			ID:      item.Link,
			Summary: item.Description,
		}
		if !item.PubDate.IsZero() {
			entry.Updated = item.PubDate.Format(time.RFC3339)
		}
		if item.Author != "" {
			entry.Author = &atomAuthor{Name: item.Author}
		}
		af.Entries = append(af.Entries, entry)
	}

	return xml.MarshalIndent(af, "", "  ")
}

// Handler returns an HTTP handler that serves the feed.
func Handler(feed Feed) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data []byte
		var err error

		switch feed.FeedType() {
		case Atom:
			w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
			data, err = GenerateAtom(feed)
		default:
			w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
			data, err = GenerateRSS(feed)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte(xml.Header))
		w.Write(data)
	}
}
