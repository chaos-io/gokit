package tracing

import (
	"context"
	"net/http"
	"net/url"

	"github.com/chaos-io/chaos/logs"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// var (
// 	traceparent = http.CanonicalHeaderKey("traceparent")
// 	tracestate  = http.CanonicalHeaderKey("tracestate")
// )

// HTTPToContext returns an http RequestFunc that tries to join with an
// OpenTelemetry trace found in `req` and starts a new Span called
// `operationName` accordingly. If no trace could be found in `req`, the Span
// will be a trace root. The Span is incorporated in the returned Context and
// can be retrieved with otel.GetSpan(ctx).
func HTTPToContext(tracer trace.Tracer, operationName string) func(ctx context.Context, req *http.Request) context.Context {
	return func(ctx context.Context, req *http.Request) context.Context {
		prop := propagation.TraceContext{}
		ctx = prop.Extract(ctx, propagation.HeaderCarrier(req.Header))

		ctx, span := tracer.Start(ctx, operationName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		// Set attributes on the span
		span.SetAttributes(attribute.String("http.method", req.Method))
		if req.URL != nil {
			span.SetAttributes(attribute.String("http.url", req.URL.String()))
			span.SetAttributes(attribute.String("http.scheme", req.URL.String()))

			// Optional: Break down the URL into more attributes (path, query)
			if parsedURL, err := url.Parse(req.URL.String()); err == nil {
				span.SetAttributes(attribute.String("http.path", parsedURL.Path))
				span.SetAttributes(attribute.String("http.query", parsedURL.RawQuery))
			} else {
				logs.Warnw("failed to parse url", "url", req.URL.String(), "error", err)
			}
		} else {
			logs.Warn("request URL is nil")
		}

		return ctx
	}
}

func InjectHTTPHeader(ctx context.Context, header http.Header) {
	prop := propagation.TraceContext{}
	prop.Inject(ctx, propagation.HeaderCarrier(header))
}
