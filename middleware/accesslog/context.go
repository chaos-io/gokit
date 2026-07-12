package accesslog

import (
	"context"
	"net"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func traceFields(ctx context.Context) []any {
	span := trace.SpanContextFromContext(ctx)
	if !span.IsValid() {
		return nil
	}
	return []any{"trace_id", span.TraceID().String(), "span_id", span.SpanID().String()}
}

func httpRequestID(r *http.Request) string {
	if r == nil {
		return ""
	}
	return r.Header.Get("X-Request-ID")
}

func grpcRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	if values := md.Get("x-request-id"); len(values) > 0 {
		return values[0]
	}
	return ""
}

func httpRemoteIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	return host(r.RemoteAddr)
}

func grpcRemoteIP(ctx context.Context) string {
	remote, ok := peer.FromContext(ctx)
	if !ok || remote.Addr == nil {
		return ""
	}
	if tcp, ok := remote.Addr.(*net.TCPAddr); ok {
		return tcp.IP.String()
	}
	return host(remote.Addr.String())
}

func host(address string) string {
	value, _, err := net.SplitHostPort(address)
	if err == nil {
		return value
	}
	return address
}
