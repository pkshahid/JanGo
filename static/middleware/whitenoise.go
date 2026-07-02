package middleware

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/pkshahid/JanGo/core/settings"
	godjangohttp "github.com/pkshahid/JanGo/http"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type WhiteNoiseMiddleware struct {
	staticRoot string
	staticURL  string
	files      map[string]*StaticFile
	filesMutex sync.RWMutex
}

type StaticFile struct {
	Content      []byte
	GzipContent  []byte
	ContentType  string
	ETag         string
	LastModified time.Time
	Size         int64
}

func NewWhiteNoiseMiddleware() *WhiteNoiseMiddleware {
	s := settings.Get()
	wm := &WhiteNoiseMiddleware{
		staticRoot: s.STATIC_ROOT,
		staticURL:  s.STATIC_URL,
		files:      make(map[string]*StaticFile),
	}

	if wm.staticRoot != "" && wm.staticURL != "" {
		wm.preloadFiles()
	}

	return wm
}

func (m *WhiteNoiseMiddleware) preloadFiles() {
	if _, err := os.Stat(m.staticRoot); os.IsNotExist(err) {
		return
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 20) // Limit concurrency

	filepath.Walk(m.staticRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(m.staticRoot, path)
		// normalize path separators for URL
		relPath = strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

		wg.Add(1)
		semaphore <- struct{}{}
		go func(p, rel string, i os.FileInfo) {
			defer wg.Done()
			defer func() { <-semaphore }()

			content, err := os.ReadFile(p)
			if err != nil {
				return
			}

			hash := md5.Sum(content)
			etag := `"` + hex.EncodeToString(hash[:]) + `"`

			ext := filepath.Ext(p)
			contentType := mime.TypeByExtension(ext)
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			// Gzip compress
			var gzipBuf bytes.Buffer
			gz := gzip.NewWriter(&gzipBuf)
			gz.Write(content)
			gz.Close()
			gzipContent := gzipBuf.Bytes()

			sf := &StaticFile{
				Content:      content,
				GzipContent:  gzipContent,
				ContentType:  contentType,
				ETag:         etag,
				LastModified: i.ModTime(),
				Size:         i.Size(),
			}

			m.filesMutex.Lock()
			m.files[rel] = sf
			m.filesMutex.Unlock()

		}(path, relPath, info)

		return nil
	})

	wg.Wait()
}

func (m *WhiteNoiseMiddleware) Process(next func(*godjangohttp.Request) godjangohttp.Response) func(*godjangohttp.Request) godjangohttp.Response {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		if m.staticURL != "" && strings.HasPrefix(req.URL.Path, m.staticURL) {
			relPath := strings.TrimPrefix(req.URL.Path, m.staticURL)
			relPath = filepath.FromSlash(filepath.Clean("/" + relPath))

			// Remove the leading slash added by Clean
			relPath = strings.TrimPrefix(relPath, string(os.PathSeparator))

			m.filesMutex.RLock()
			file, exists := m.files[relPath]
			m.filesMutex.RUnlock()

			if exists {
				return m.serveFile(req, file)
			}

			// Fallback: try to read from disk if not preloaded (e.g., added later)
			fullPath := filepath.Join(m.staticRoot, relPath)
			absRoot, _ := filepath.Abs(m.staticRoot)
			absFull, _ := filepath.Abs(fullPath)
			if !strings.HasPrefix(absFull, absRoot) {
				return godjangohttp.NewHttpResponse("Forbidden", http.StatusForbidden)
			}

			if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
				content, err := os.ReadFile(fullPath)
				if err == nil {
					hash := md5.Sum(content)
					etag := `"` + hex.EncodeToString(hash[:]) + `"`
					ext := filepath.Ext(fullPath)
					contentType := mime.TypeByExtension(ext)
					if contentType == "" {
						contentType = "application/octet-stream"
					}
					var gzipBuf bytes.Buffer
					gz := gzip.NewWriter(&gzipBuf)
					gz.Write(content)
					gz.Close()

					sf := &StaticFile{
						Content:      content,
						GzipContent:  gzipBuf.Bytes(),
						ContentType:  contentType,
						ETag:         etag,
						LastModified: info.ModTime(),
						Size:         info.Size(),
					}

					m.filesMutex.Lock()
					m.files[relPath] = sf
					m.filesMutex.Unlock()

					return m.serveFile(req, sf)
				}
			}
		}

		return next(req)
	}
}

func (m *WhiteNoiseMiddleware) serveFile(req *godjangohttp.Request, file *StaticFile) godjangohttp.Response {
	// Check If-None-Match
	if match := req.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, file.ETag) {
			resp := godjangohttp.NewHttpResponse("", http.StatusNotModified)
			resp.Headers.Set("ETag", file.ETag)
			return resp
		}
	}

	// Check If-Modified-Since
	if since := req.Header.Get("If-Modified-Since"); since != "" {
		t, err := http.ParseTime(since)
		if err == nil && file.LastModified.Before(t.Add(1*time.Second)) {
			resp := godjangohttp.NewHttpResponse("", http.StatusNotModified)
			resp.Headers.Set("ETag", file.ETag)
			return resp
		}
	}

	resp := godjangohttp.NewHttpResponse("", http.StatusOK)
	resp.Headers.Set("Content-Type", file.ContentType)
	resp.Headers.Set("ETag", file.ETag)
	resp.Headers.Set("Last-Modified", file.LastModified.UTC().Format(http.TimeFormat))
	resp.Headers.Set("Cache-Control", "max-age=31536000, public, immutable")

	// Check if client accepts gzip
	if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		resp.Headers.Set("Content-Encoding", "gzip")
		resp.Headers.Set("Vary", "Accept-Encoding")
		resp.Body = io.NopCloser(bytes.NewBuffer(file.GzipContent))
		resp.Headers.Set("Content-Length", fmt.Sprintf("%d", len(file.GzipContent)))
	} else {
		resp.Body = io.NopCloser(bytes.NewBuffer(file.Content))
		resp.Headers.Set("Content-Length", fmt.Sprintf("%d", file.Size))
	}

	return resp
}
