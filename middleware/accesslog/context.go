package accesslog

import (
	"context"
	"net"

	"go.opentelemetry.io/otel/trace"
)

func traceFields(ctx context.Context) []any {
	span := trace.SpanContextFromContext(ctx)
	if !span.IsValid() {
		return nil
	}
	return []any{"trace_id", span.TraceID().String(), "span_id", span.SpanID().String()}
}

func remoteHost(address string) string {
	host, _, err := net.SplitHostPort(address)
	if err == nil {
		return host
	}
	return address
}
