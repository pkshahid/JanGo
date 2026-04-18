package middleware

import (
	"log/slog"
	"time"
	"fmt"
	"io"
	"bytes"

	godjangohttp "github.com/pkshahid/JanGo/http"
)

// RequestLoggingMiddleware logs request details.
func RequestLoggingMiddleware(next Handler) Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		start := time.Now()

		resp := next(req)

		duration := time.Since(start)

		statusCode := 500
		var bodyLen int64 = 0

		if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
			statusCode = hr.StatusCode
			// Try to estimate bytes length if not streaming
			if hr.Body != nil {
				// In a real framework, you'd wrap the ResponseWriter to count bytes written.
				// Since we return a Response struct, we can't always know length unless it's buffered.
				// For simple HttpResponse we can try to find length.
				// We don't read the whole body just to log it. We check Content-Length header.
				cl := hr.Headers.Get("Content-Length")
				if cl != "" {
					fmt.Sscanf(cl, "%d", &bodyLen)
				} else if b, err := io.ReadAll(hr.Body); err == nil {
					// Fallback: Read body and replace it
					bodyLen = int64(len(b))
					hr.Body = io.NopCloser(bytes.NewReader(b))
				}
			}
		}

		// Color coding (basic ANSI codes)
		statusColor := "\033[0m" // Default reset
		switch {
		case statusCode >= 200 && statusCode < 300:
			statusColor = "\033[32m" // Green
		case statusCode >= 300 && statusCode < 400:
			statusColor = "\033[36m" // Cyan
		case statusCode >= 400 && statusCode < 500:
			statusColor = "\033[33m" // Yellow
		case statusCode >= 500:
			statusColor = "\033[31m" // Red
		}
		resetColor := "\033[0m"

		msg := fmt.Sprintf("[%s%d%s] %s %s %s %d bytes", statusColor, statusCode, resetColor, req.Method, req.Path, duration, bodyLen)

		slog.Info(msg)

		return resp
	}
}
