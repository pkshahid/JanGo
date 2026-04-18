package forms

import (
	"strings"

	"github.com/pkshahid/JanGo/template"
)

// Media handles CSS and JS files for a widget or form.
type Media struct {
	CSS map[string][]string // Keyed by medium, e.g., "all", "screen"
	JS  []string
}

// NewMedia creates an empty Media object.
func NewMedia() *Media {
	return &Media{
		CSS: make(map[string][]string),
		JS:  []string{},
	}
}

// Merge combines two Media objects, deduplicating paths.
func (m *Media) Merge(other *Media) {
	if other == nil {
		return
	}

	// Merge CSS
	for medium, paths := range other.CSS {
		existing := m.CSS[medium]
		for _, path := range paths {
			if !contains(existing, path) {
				existing = append(existing, path)
			}
		}
		m.CSS[medium] = existing
	}

	// Merge JS
	for _, path := range other.JS {
		if !contains(m.JS, path) {
			m.JS = append(m.JS, path)
		}
	}
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// RenderCSS outputs <link> tags.
func (m *Media) RenderCSS() template.SafeString {
	var buf strings.Builder
	for medium, paths := range m.CSS {
		for _, path := range paths {
			buf.WriteString(`<link href="`)
			buf.WriteString(path)
			buf.WriteString(`" type="text/css" media="`)
			buf.WriteString(medium)
			buf.WriteString(`" rel="stylesheet">` + "\n")
		}
	}
	return template.SafeString(strings.TrimSuffix(buf.String(), "\n"))
}

// RenderJS outputs <script> tags.
func (m *Media) RenderJS() template.SafeString {
	var buf strings.Builder
	for _, path := range m.JS {
		buf.WriteString(`<script src="`)
		buf.WriteString(path)
		buf.WriteString(`"></script>` + "\n")
	}
	return template.SafeString(strings.TrimSuffix(buf.String(), "\n"))
}

// Render outputs all tags.
func (m *Media) Render() template.SafeString {
	css := m.RenderCSS()
	js := m.RenderJS()

	if css == "" {
		return js
	}
	if js == "" {
		return css
	}
	return template.SafeString(string(css) + "\n" + string(js))
}
