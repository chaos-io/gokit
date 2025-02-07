package tracing

import (
	"context"
	"net/http"

	"github.com/chaos-io/chaos/logs"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// HTTPToContext returns an http RequestFunc that tries to join with an
// OpenTelemetry trace found in `req` and starts a new Span called
// `operationName` accordingly. If no trace could be found in `req`, the Span
// will be a trace root. The Span is incorporated in the returned Context and
// can be retrieved with otel.GetSpan(ctx).
func HTTPToContext(tracer trace.Tracer, operationName string) func(ctx context.Context, req *http.Request) context.Context {
	return func(ctx context.Context, req *http.Request) context.Context {
		// Try to join to a trace propagated in `req`.
		// Extract the context from HTTP headers.
		propagator := otel.GetTextMapPropagator()

		// Extract the context from incoming HTTP headers (if exists)
		ctx = propagator.Extract(ctx, RequestCarrier(req.Header))

		// Create a new Span from the extracted context, or root span if none exists
		ctx, span := tracer.Start(ctx, operationName, trace.WithAttributes(
			semconv.HTTPMethod(req.Method),
			semconv.HTTPURL(req.URL.String()),
		))

		// Ensure the span is finished when the context is done
		defer span.End()

		// Inject the trace context into the outgoing request headers (for downstream services)
		propagator.Inject(ctx, RequestCarrier(req.Header))

		// Optional: Log any errors encountered during the trace context extraction
		// If the SpanContext is empty, it means no valid trace context was found in the request
		spanContext := trace.SpanFromContext(ctx).SpanContext()
		if !spanContext.IsValid() {
			logs.Warnw("tracing span context invalid", "traceID", spanContext.TraceID())
		}

		return ctx
	}
}

// RequestCarrier wraps http.Header to implement the TextMapCarrier interface required by OpenTelemetry.
type RequestCarrier http.Header

// Get retrieves a key from the http.Header.
func (c RequestCarrier) Get(key string) string {
	return http.Header(c).Get(key)
}

// Set sets a key-value pair in the http.Header.
func (c RequestCarrier) Set(key, value string) {
	http.Header(c).Set(key, value)
}

// Keys retrieves all keys in the http.Header.
func (c RequestCarrier) Keys() []string {
	keys := make([]string, 0, len(http.Header(c)))
	for key := range http.Header(c) {
		keys = append(keys, key)
	}
	return keys
}
