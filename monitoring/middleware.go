package monitoring

import (
	"log/slog"
	"strconv"
	"time"

	godjangohttp "github.com/godjango/godjango/http"
)

// RequestLogMiddleware logs each HTTP request with structured fields.
func RequestLogMiddleware(next func(*godjangohttp.Request) godjangohttp.Response) func(*godjangohttp.Request) godjangohttp.Response {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		start := time.Now()

		resp := next(req)

		duration := time.Since(start)
		status := 200

		if hr, ok := resp.(interface{ StatusCode() int }); ok {
			status = hr.StatusCode()
		} else {
			switch r := resp.(type) {
			case *godjangohttp.HttpResponse:
				status = r.StatusCode
			case *godjangohttp.JsonResponse:
				status = r.StatusCode
			case *godjangohttp.RedirectResponse:
				status = r.StatusCode
			}
		}

		// Extract User ID if available
		userID := "anonymous"
		if req.User != nil && !req.User.IsAnonymous() {
			userID = strconv.FormatUint(req.User.ID(), 10)
		}

		Logger.Info("HTTP Request",
			slog.String("method", req.Method),
			slog.String("path", req.Path),
			slog.Int("status", status),
			slog.Duration("duration", duration),
			slog.String("ip", req.RemoteAddr),
			slog.String("user_id", userID),
		)

		return resp
	}
}
