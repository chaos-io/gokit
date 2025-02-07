package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

// GRPCToContext returns a grpc RequestFunc that tries to join with an
// OpenTelemetry trace found in `md` and starts a new Span called
// `operationName` accordingly. If no trace could be found in `md`, the Span
// will be a trace root. The Span is incorporated in the returned Context.
func GRPCToContext(tracer trace.Tracer, operationName string) func(ctx context.Context, md metadata.MD) context.Context {
	return func(ctx context.Context, md metadata.MD) context.Context {
		propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
		ctx = propagator.Extract(ctx, metadataTextMap(md))

		ctx, span := tracer.Start(ctx, operationName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		// Add gRPC metadata as attributes to the span. This is useful for debugging
		// and analysis. Consider making this configurable if you have sensitive data
		// in your metadata.
		for key, values := range md {
			for _, value := range values {
				span.SetAttributes(attribute.String("grpc.metadata."+key, value))
			}
		}

		return ctx
	}
}

// metadataTextMap implements the TextMapCarrier interface for gRPC metadata.
type metadataTextMap metadata.MD

// Get retrieves a single value for a given key.
func (m metadataTextMap) Get(key string) string {
	if values := m[key]; len(values) > 0 {
		return values[0]
	}
	return ""
}

// Set sets a value for a given key.
func (m metadataTextMap) Set(key string, value string) {
	m[key] = []string{value}
}

// Keys lists the keys stored in this carrier.
func (m metadataTextMap) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
