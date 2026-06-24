// Package sitemaps implements Django's sitemaps framework.
// It provides automatic XML sitemap generation for search engine optimization.
package sitemaps

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

// ChangeFreq represents how frequently a page is likely to change.
type ChangeFreq string

const (
	Always  ChangeFreq = "always"
	Hourly  ChangeFreq = "hourly"
	Daily   ChangeFreq = "daily"
	Weekly  ChangeFreq = "weekly"
	Monthly ChangeFreq = "monthly"
	Yearly  ChangeFreq = "yearly"
	Never   ChangeFreq = "never"
)

// URL represents a single URL entry in the sitemap.
type URL struct {
	XMLName    xml.Name   `xml:"url"`
	Loc        string     `xml:"loc"`
	LastMod    *time.Time `xml:"lastmod,omitempty"`
	ChangeFreq ChangeFreq `xml:"changefreq,omitempty"`
	Priority   float64    `xml:"priority,omitempty"`
}

// URLSet represents the root element of a sitemap XML file.
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

// SitemapIndex represents an index of multiple sitemaps.
type SitemapIndex struct {
	XMLName  xml.Name       `xml:"sitemapindex"`
	XMLNS    string         `xml:"xmlns,attr"`
	Sitemaps []SitemapEntry `xml:"sitemap"`
}

// SitemapEntry represents a single sitemap in the index.
type SitemapEntry struct {
	XMLName xml.Name   `xml:"sitemap"`
	Loc     string     `xml:"loc"`
	LastMod *time.Time `xml:"lastmod,omitempty"`
}

// Item represents an object that can produce sitemap URLs.
type Item interface{}

// Sitemap defines the interface for generating sitemap entries.
// Equivalent to Django's Sitemap class.
type Sitemap interface {
	// Items returns all objects to include in this sitemap section.
	Items() []Item
	// Location returns the URL path for a given item.
	Location(item Item) string
	// LastModified returns the last modification time for an item.
	LastModified(item Item) *time.Time
	// Changefreq returns how frequently the item changes.
	Changefreq(item Item) ChangeFreq
	// Priority returns the priority of this URL relative to others (0.0 to 1.0).
	Priority(item Item) float64
}

// StaticSitemap is a simple sitemap for static URLs.
type StaticSitemap struct {
	URLs            []string
	ChangeFrequency ChangeFreq
	PriorityValue   float64
}

func (s *StaticSitemap) Items() []Item {
	items := make([]Item, len(s.URLs))
	for i, u := range s.URLs {
		items[i] = u
	}
	return items
}

func (s *StaticSitemap) Location(item Item) string {
	return item.(string)
}

func (s *StaticSitemap) LastModified(item Item) *time.Time {
	return nil
}

func (s *StaticSitemap) Changefreq(item Item) ChangeFreq {
	if s.ChangeFrequency != "" {
		return s.ChangeFrequency
	}
	return Weekly
}

func (s *StaticSitemap) Priority(item Item) float64 {
	if s.PriorityValue != 0 {
		return s.PriorityValue
	}
	return 0.5
}

// Registry holds all registered sitemaps.
type Registry struct {
	sitemaps map[string]Sitemap
	baseURL  string
}

// NewRegistry creates a new sitemap registry.
func NewRegistry(baseURL string) *Registry {
	return &Registry{
		sitemaps: make(map[string]Sitemap),
		baseURL:  baseURL,
	}
}

// Register adds a named sitemap to the registry.
func (r *Registry) Register(name string, sitemap Sitemap) {
	r.sitemaps[name] = sitemap
}

// GenerateURLSet generates the URL set for a specific sitemap.
func (r *Registry) GenerateURLSet(name string) (*URLSet, error) {
	sitemap, exists := r.sitemaps[name]
	if !exists {
		return nil, fmt.Errorf("sitemaps: no sitemap named %q", name)
	}

	urlSet := &URLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
	}

	for _, item := range sitemap.Items() {
		url := URL{
			Loc:        r.baseURL + sitemap.Location(item),
			LastMod:    sitemap.LastModified(item),
			ChangeFreq: sitemap.Changefreq(item),
			Priority:   sitemap.Priority(item),
		}
		urlSet.URLs = append(urlSet.URLs, url)
	}

	return urlSet, nil
}

// GenerateIndex generates a sitemap index referencing all registered sitemaps.
func (r *Registry) GenerateIndex() *SitemapIndex {
	index := &SitemapIndex{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
	}

	for name := range r.sitemaps {
		entry := SitemapEntry{
			Loc: r.baseURL + "/sitemap-" + name + ".xml",
		}
		index.Sitemaps = append(index.Sitemaps, entry)
	}

	return index
}

// Handler returns an HTTP handler that serves sitemap XML.
func (r *Registry) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")

		// If there's only one sitemap, serve it directly
		if len(r.sitemaps) == 1 {
			for name := range r.sitemaps {
				urlSet, err := r.GenerateURLSet(name)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Write([]byte(xml.Header))
				enc := xml.NewEncoder(w)
				enc.Indent("", "  ")
				enc.Encode(urlSet)
				return
			}
		}

		// Multiple sitemaps: serve index
		index := r.GenerateIndex()
		w.Write([]byte(xml.Header))
		enc := xml.NewEncoder(w)
		enc.Indent("", "  ")
		enc.Encode(index)
	}
}

// SectionHandler returns an HTTP handler for a named sitemap section.
func (r *Registry) SectionHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")

		urlSet, err := r.GenerateURLSet(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Write([]byte(xml.Header))
		enc := xml.NewEncoder(w)
		enc.Indent("", "  ")
		enc.Encode(urlSet)
	}
}
