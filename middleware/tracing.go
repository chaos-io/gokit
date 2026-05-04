package middleware

import (
	"context"
	"net/http"

	gokittracing "github.com/chaos-io/gokit/tracing"
	"github.com/go-kit/kit/endpoint"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	httptransport "github.com/go-kit/kit/transport/http"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

// TraceServer starts and ends a server span around the endpoint invocation.
func TraceServer(tracer trace.Tracer, operationName string, opts ...gokittracing.EndpointOption) endpoint.Middleware {
	return gokittracing.TraceServer(tracer, operationName, opts...)
}

// TraceClient starts and ends a client span around the endpoint invocation.
func TraceClient(tracer trace.Tracer, operationName string, opts ...gokittracing.EndpointOption) endpoint.Middleware {
	return gokittracing.TraceClient(tracer, operationName, opts...)
}

// HTTPTraceContext extracts incoming HTTP propagation headers into context.
func HTTPTraceContext() httptransport.RequestFunc {
	return func(ctx context.Context, req *http.Request) context.Context {
		return gokittracing.HTTPToContext(ctx, req)
	}
}

// HTTPServerTraceOption wires HTTP propagation extraction into a go-kit HTTP server.
func HTTPServerTraceOption() httptransport.ServerOption {
	return httptransport.ServerBefore(HTTPTraceContext())
}

// GRPCTraceContext extracts incoming gRPC propagation metadata into context.
func GRPCTraceContext() grpctransport.ServerRequestFunc {
	return func(ctx context.Context, md metadata.MD) context.Context {
		return gokittracing.GRPCToContext(ctx, md)
	}
}

// GRPCServerTraceOption wires gRPC propagation extraction into a go-kit gRPC server.
func GRPCServerTraceOption() grpctransport.ServerOption {
	return grpctransport.ServerBefore(GRPCTraceContext())
}
