package monitoring

import (
	"net/http"
	"net/http/pprof"
	"runtime"

	"github.com/godjango/godjango/core/settings"
	godjangohttp "github.com/godjango/godjango/http"
	"github.com/godjango/godjango/http/urls"
)

// ensureDebug returns true if DEBUG is enabled.
func ensureDebug() bool {
	s := settings.Get()
	return s.DEBUG
}

// PprofIndex handler
func PprofIndex(req *godjangohttp.Request) godjangohttp.Response {
	if !ensureDebug() {
		return godjangohttp.HttpResponseForbidden("pprof is disabled in production")
	}
	return bridgePprof(pprof.Index, req.Request)
}

// PprofCmdline handler
func PprofCmdline(req *godjangohttp.Request) godjangohttp.Response {
	if !ensureDebug() {
		return godjangohttp.HttpResponseForbidden("pprof is disabled in production")
	}
	return bridgePprof(pprof.Cmdline, req.Request)
}

// PprofProfile handler
func PprofProfile(req *godjangohttp.Request) godjangohttp.Response {
	if !ensureDebug() {
		return godjangohttp.HttpResponseForbidden("pprof is disabled in production")
	}
	return bridgePprof(pprof.Profile, req.Request)
}

// PprofSymbol handler
func PprofSymbol(req *godjangohttp.Request) godjangohttp.Response {
	if !ensureDebug() {
		return godjangohttp.HttpResponseForbidden("pprof is disabled in production")
	}
	return bridgePprof(pprof.Symbol, req.Request)
}

// PprofTrace handler
func PprofTrace(req *godjangohttp.Request) godjangohttp.Response {
	if !ensureDebug() {
		return godjangohttp.HttpResponseForbidden("pprof is disabled in production")
	}
	return bridgePprof(pprof.Trace, req.Request)
}

// GoroutinesDump handler returns a formatted dump of all goroutines.
func GoroutinesDump(req *godjangohttp.Request) godjangohttp.Response {
	if !ensureDebug() {
		return godjangohttp.HttpResponseForbidden("Goroutine dump is disabled in production")
	}

	buf := make([]byte, 1<<20)
	n := runtime.Stack(buf, true)

	resp := godjangohttp.NewHttpResponse(string(buf[:n]), http.StatusOK)
	resp.Headers.Set("Content-Type", "text/plain; charset=utf-8")
	return resp
}

// MemoryStats handler returns a JSON dump of runtime memory statistics.
func MemoryStats(req *godjangohttp.Request) godjangohttp.Response {
	if !ensureDebug() {
		return godjangohttp.HttpResponseForbidden("Memory stats disabled in production")
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	data := map[string]any{
		"Alloc":      memStats.Alloc,
		"TotalAlloc": memStats.TotalAlloc,
		"Sys":        memStats.Sys,
		"NumGC":      memStats.NumGC,
		"HeapAlloc":  memStats.HeapAlloc,
		"HeapSys":    memStats.HeapSys,
		"HeapIdle":   memStats.HeapIdle,
		"HeapInuse":  memStats.HeapInuse,
		"HeapObjects": memStats.HeapObjects,
	}

	return godjangohttp.NewJsonResponse(data)
}

// bridgePprof converts a standard http.HandlerFunc into a godjango Response.
func bridgePprof(handler http.HandlerFunc, req *http.Request) godjangohttp.Response {
	return &PprofResponse{
		handler: handler,
		req:     req,
	}
}

type PprofResponse struct {
	handler http.HandlerFunc
	req     *http.Request
}

func (p *PprofResponse) Write(w http.ResponseWriter) {
	p.handler(w, p.req)
}

// RegisterRoutes registers all monitoring routes to the provided router.
func RegisterRoutes(router *urls.Router) {
	// Health
	router.Add(urls.Path("/health/", HealthView, "monitoring_health", nil))

	// Metrics
	router.Add(urls.Path("/metrics/", MetricsView, "monitoring_metrics", nil))

	// pprof - requires trailing slash exact matches or dynamic catch-all based on the router
	router.Add(urls.Path("/debug/pprof/", PprofIndex, "monitoring_pprof_index", nil))
	router.Add(urls.Path("/debug/pprof/cmdline/", PprofCmdline, "monitoring_pprof_cmdline", nil))
	router.Add(urls.Path("/debug/pprof/profile/", PprofProfile, "monitoring_pprof_profile", nil))
	router.Add(urls.Path("/debug/pprof/symbol/", PprofSymbol, "monitoring_pprof_symbol", nil))
	router.Add(urls.Path("/debug/pprof/trace/", PprofTrace, "monitoring_pprof_trace", nil))

	// Custom dumps
	router.Add(urls.Path("/debug/goroutines/", GoroutinesDump, "monitoring_goroutines", nil))
	router.Add(urls.Path("/debug/memory/", MemoryStats, "monitoring_memory", nil))
}
