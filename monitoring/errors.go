package monitoring

import (
	"log/slog"
	"sync"

	godjangohttp "github.com/godjango/godjango/http"
)

// ErrorReporter defines the interface for capturing exceptions.
type ErrorReporter interface {
	Capture(err error, req *godjangohttp.Request)
}

// SlogReporter logs errors using the structured logger.
type SlogReporter struct{}

func (s *SlogReporter) Capture(err error, req *godjangohttp.Request) {
	if req != nil {
		Logger.Error("Captured Error",
			slog.String("error", err.Error()),
			slog.String("path", req.Path),
			slog.String("method", req.Method),
		)
	} else {
		Logger.Error("Captured Error", slog.String("error", err.Error()))
	}
}

var (
	reporterMu sync.RWMutex
	reporters  []ErrorReporter
)

func init() {
	// Register the default slog reporter
	RegisterErrorReporter(&SlogReporter{})
}

// RegisterErrorReporter adds an error reporter to the global list.
func RegisterErrorReporter(r ErrorReporter) {
	reporterMu.Lock()
	defer reporterMu.Unlock()
	reporters = append(reporters, r)
}

// CaptureError broadcasts the error to all registered reporters.
func CaptureError(err error, req *godjangohttp.Request) {
	reporterMu.RLock()
	defer reporterMu.RUnlock()
	for _, r := range reporters {
		r.Capture(err, req)
	}
}
