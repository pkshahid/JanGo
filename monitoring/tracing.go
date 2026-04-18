package monitoring

import (
	"context"

	godjangohttp "github.com/godjango/godjango/http"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func init() {
	// Initialize a noop tracer by default
	tracer = otel.Tracer("github.com/godjango/godjango")
}

// InitTracing sets up OpenTelemetry tracing.
// For now, it configures a stdout exporter.
func InitTracing(exporterType string) error {
	var exporter sdktrace.SpanExporter
	var err error

	switch exporterType {
	case "stdout":
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	// In the future:
	// case "otlp":
	//     exporter, err = otlptracegrpc.New(...)
	default:
		// noop
		return nil
	}

	if err != nil {
		return err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tracer = tp.Tracer("github.com/godjango/godjango")
	return nil
}

// TracingMiddleware extracts trace context from headers and wraps the request in a span.
func TracingMiddleware(next func(*godjangohttp.Request) godjangohttp.Response) func(*godjangohttp.Request) godjangohttp.Response {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		ctx := otel.GetTextMapPropagator().Extract(req.Context, propagation.HeaderCarrier(req.Header))

		spanName := req.Method + " " + req.Path
		if req.ResolverMatch != nil {
			// Try to use a less high-cardinality name if possible
			spanName = "HTTP " + req.Method
		}

		ctx, span := tracer.Start(ctx, spanName)
		defer span.End()

		req.Context = ctx

		return next(req)
	}
}

// StartDBSpan is a helper to wrap database queries in a span.
func StartDBSpan(ctx context.Context, operation string) (context.Context, trace.Span) {
	return tracer.Start(ctx, "db."+operation)
}

// StartCacheSpan is a helper to wrap cache operations in a span.
func StartCacheSpan(ctx context.Context, operation string) (context.Context, trace.Span) {
	return tracer.Start(ctx, "cache."+operation)
}
