package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"strconv"
	"strings"

	godjangohttp "github.com/pkshahid/JanGo/http"
)

// GZipMiddleware compresses responses above a threshold length.
func GZipMiddleware(next Handler) Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		resp := next(req)

		// Respect Accept-Encoding
		if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			return resp
		}

		if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
			// Don't compress if already compressed or short
			if hr.Headers.Get("Content-Encoding") != "" {
				return resp
			}

			// We need to read the body to compress it.
			// Streaming responses should not be buffered this way in a real implementation,
			// but for HttpResponse this is acceptable.
			if hr.Body != nil {
				bodyBytes, err := io.ReadAll(hr.Body)
				if err == nil && len(bodyBytes) > 200 {
					var buf bytes.Buffer
					gz := gzip.NewWriter(&buf)
					gz.Write(bodyBytes)
					gz.Close()

					hr.Body = io.NopCloser(&buf)
					hr.Headers.Set("Content-Encoding", "gzip")
					hr.Headers.Set("Content-Length", strconv.Itoa(buf.Len()))
				} else {
					// Put the unread bytes back if we didn't compress
					hr.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				}
			}
		}

		return resp
	}
}
