package monitoring

import (
	"log/slog"
	"os"

	"github.com/godjango/godjango/core/settings"
)

var Logger *slog.Logger

func init() {
	// Initialize default logger
	Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

// InitLogger reconfigures the logger based on settings.
func InitLogger() {
	s := settings.Get()

	level := slog.LevelInfo
	if s.DEBUG {
		level = slog.LevelDebug
	}

	// Can be configured from settings if a LOGGING setting is present
	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if s.DEBUG {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// SlowQueryLogger logs queries that exceed a threshold.
// Note: In GoDjango this would be injected via settings, but we implement the logic here.
func SlowQueryLogger(query string, durationMs int64, thresholdMs int64) {
	if durationMs > thresholdMs {
		Logger.Warn("Slow DB Query",
			slog.String("query", query),
			slog.Int64("duration_ms", durationMs),
		)
	}
}
