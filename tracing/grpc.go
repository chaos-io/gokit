package tracing

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/metadata"
)

var grpcPropagator = propagation.NewCompositeTextMapPropagator(
	propagation.TraceContext{},
	propagation.Baggage{},
)

// GRPCToContext extracts OpenTelemetry propagation metadata from md.
func GRPCToContext(ctx context.Context, md metadata.MD) context.Context {
	return grpcPropagator.Extract(ctx, metadataTextMap(md))
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
