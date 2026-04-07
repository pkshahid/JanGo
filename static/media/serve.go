package media

import (
	godjangohttp "github.com/godjango/godjango/http"
	"github.com/godjango/godjango/core/settings"
	"path/filepath"
	"strings"
	"os"
	"net/http"
	"strconv"
)

// ServeMedia serves files from MEDIA_ROOT. It should only be used in development.
func ServeMedia(req *godjangohttp.Request) godjangohttp.Response {
	s := settings.Get()

	if s.MEDIA_URL == "" || s.MEDIA_ROOT == "" {
		return godjangohttp.NewHttpResponse("Media not configured", http.StatusNotFound)
	}

	relPath := strings.TrimPrefix(req.URL.Path, s.MEDIA_URL)
	relPath = strings.TrimPrefix(relPath, "/")

	if strings.Contains(relPath, "..") {
		return godjangohttp.NewHttpResponse("Forbidden", http.StatusForbidden)
	}

	fullPath := filepath.Join(s.MEDIA_ROOT, relPath)

	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		return godjangohttp.NewHttpResponse("File not found", http.StatusNotFound)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return godjangohttp.NewHttpResponse("Forbidden", http.StatusForbidden)
	}
	// We don't defer file.Close() here because we pass it to the response body which closes it.

	resp := godjangohttp.NewHttpResponse("", http.StatusOK)
	resp.Body = file

	// Detect content type simply
	ext := filepath.Ext(fullPath)
	var contentType string
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	case ".svg":
		contentType = "image/svg+xml"
	case ".pdf":
		contentType = "application/pdf"
	case ".txt":
		contentType = "text/plain"
	default:
		contentType = "application/octet-stream"
	}

	resp.Headers.Set("Content-Type", contentType)
	resp.Headers.Set("Content-Length", strconv.FormatInt(info.Size(), 10))

	return resp
}

// MediaMiddleware intercepts requests starting with MEDIA_URL and serves them in DEBUG mode.
func MediaMiddleware(next func(*godjangohttp.Request) godjangohttp.Response) func(*godjangohttp.Request) godjangohttp.Response {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		s := settings.Get()
		if s.DEBUG && s.MEDIA_URL != "" && strings.HasPrefix(req.URL.Path, s.MEDIA_URL) {
			return ServeMedia(req)
		}
		return next(req)
	}
}
