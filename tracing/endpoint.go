package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-kit/kit/endpoint"
)

// TraceEndpoint returns an endpoint.Middleware that traces the execution of
// an endpoint. It creates a span with the given operation name.
func TraceEndpoint(tracer trace.Tracer, operationName string, opts ...EndpointOption) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			options := traceOptions{}
			for _, opt := range opts {
				opt(&options)
			}

			ctx, span := tracer.Start(ctx, operationName, trace.WithAttributes(options.attrs...))
			defer span.End()

			return next(ctx, request)
		}
	}
}

// TraceServer returns a Middleware that wraps the `next` Endpoint in an
// OpenTelemetry Span called `operationName` with server span.kind tag.
func TraceServer(tracer trace.Tracer, operationName string, opts ...EndpointOption) endpoint.Middleware {
	newOpts := append(opts, WithAttributes(attribute.String("span.kind", "server")))

	return TraceEndpoint(tracer, operationName, newOpts...)
}

// TraceClient returns a Middleware that wraps the `next` Endpoint in an
// OpenTelemetry Span called `operationName` with client span.kind tag.
func TraceClient(tracer trace.Tracer, operationName string, opts ...EndpointOption) endpoint.Middleware {
	newOpts := append(opts, WithAttributes(attribute.String("span.kind", "client")))

	return TraceEndpoint(tracer, operationName, newOpts...)
}

// SpanFromContext retrieves the OpenTelemetry Span stored in the context.
// Returns nil if no Span is found.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// SetAttributes adds attributes to the span in the given context.  If the
// context does not contain a span, this is a no-op.
func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return // No span in context, do nothing
	}
	span.SetAttributes(attrs...)
}
