package monitoring

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/orm/backends"
)

var (
	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "godjango_requests_total",
		Help: "Total number of HTTP requests processed, partitioned by method, path, and status code.",
	}, []string{"method", "path", "status"})

	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "godjango_request_duration_seconds",
		Help:    "Histogram of request latencies in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	dbQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "godjango_db_query_duration_seconds",
		Help:    "Histogram of database query latencies in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"backend", "operation"})

	dbConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "godjango_db_connections",
		Help: "Current number of database connections.",
	}, []string{"backend", "state"})

	cacheHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "godjango_cache_hits_total",
		Help: "Total cache hits.",
	}, []string{"backend"})

	cacheMisses = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "godjango_cache_misses_total",
		Help: "Total cache misses.",
	}, []string{"backend"})

	goroutinesGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "godjango_goroutines",
		Help: "Current number of goroutines.",
	})

	memoryGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "godjango_memory_bytes",
		Help: "Current memory usage in bytes (Alloc).",
	})
)

func init() {
	// Start background goroutine to update runtime metrics
	go updateRuntimeMetrics()
}

func updateRuntimeMetrics() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var memStats runtime.MemStats
	for range ticker.C {
		goroutinesGauge.Set(float64(runtime.NumGoroutine()))

		runtime.ReadMemStats(&memStats)
		memoryGauge.Set(float64(memStats.Alloc))

		// Also update DB connection stats if possible
		updateDBStats()
	}
}

func updateDBStats() {
	// In a real scenario we might need to iterate over all configured databases
	// For now we check "default" if available.
	if backend, err := backends.GetBackend("default"); err == nil {
		if db := backend.DB(); db != nil {
			stats := db.Stats()
			dbConnections.WithLabelValues("default", "open").Set(float64(stats.OpenConnections))
			dbConnections.WithLabelValues("default", "idle").Set(float64(stats.Idle))
		}
	}
}

// MetricsMiddleware tracks requests for Prometheus.
func MetricsMiddleware(next func(*godjangohttp.Request) godjangohttp.Response) func(*godjangohttp.Request) godjangohttp.Response {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		start := time.Now()

		// If it's a metrics request itself, maybe skip recording to avoid noise
		if req.Path == "/metrics" {
			return next(req)
		}

		resp := next(req)

		duration := time.Since(start).Seconds()
		status := 200

		// Try to extract status code if the response implements a specific interface
		// Response in GoDjango might not directly expose StatusCode, but let's assume
		// we can access it through casting.
		if hr, ok := resp.(interface{ StatusCode() int }); ok {
			status = hr.StatusCode()
		} else {
			// In GoDjango, concrete types like *HttpResponse expose it as a public field
			// but we are returning interfaces. Let's do a switch type.
			switch r := resp.(type) {
			case *godjangohttp.HttpResponse:
				status = r.StatusCode
			case *godjangohttp.JsonResponse:
				status = r.StatusCode
			case *godjangohttp.RedirectResponse:
				status = r.StatusCode
			}
		}

		statusStr := strconv.Itoa(status)

		// Note: Using req.Path directly might cause high cardinality if paths have IDs in them.
		// In a real implementation we would use req.ResolverMatch.Route or similar.
		pathStr := req.Path
		if req.ResolverMatch != nil {
			// Usually ResolverMatch has a pattern or name we can use
			// We'll just stick to path for now or maybe view name if available.
			pathStr = "matched_route" // simplification
		}

		requestsTotal.WithLabelValues(req.Method, pathStr, statusStr).Inc()
		requestDuration.WithLabelValues(req.Method, pathStr).Observe(duration)

		return resp
	}
}

// MetricsView serves the Prometheus metrics endpoint.
func MetricsView(req *godjangohttp.Request) godjangohttp.Response {
	// We need to bridge promhttp.Handler() which returns an http.Handler
	// to a GoDjango Response.
	// We can use a custom response that just delegates to the promhttp.Handler.
	return &PrometheusResponse{
		handler: promhttp.Handler(),
		req:     req.Request, // The underlying *http.Request
	}
}

// PrometheusResponse bridges standard http.Handler to godjangohttp.Response
type PrometheusResponse struct {
	handler http.Handler
	req     *http.Request
}

func (p *PrometheusResponse) Write(w http.ResponseWriter) {
	p.handler.ServeHTTP(w, p.req)
}

// RecordDBQuery duration and stats
func RecordDBQuery(backend, operation string, duration time.Duration) {
	dbQueryDuration.WithLabelValues(backend, operation).Observe(duration.Seconds())
}

// RecordCacheHit misses/hits
func RecordCacheHit(backend string) {
	cacheHits.WithLabelValues(backend).Inc()
}

func RecordCacheMiss(backend string) {
	cacheMisses.WithLabelValues(backend).Inc()
}
