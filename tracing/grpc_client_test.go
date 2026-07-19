package tracing

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestGRPCUnaryClientInterceptorInjectsTraceAndRecordsError(t *testing.T) {
	tracer, recorder := testTracer()
	wantErr := errors.New("unavailable")
	interceptor := GRPCUnaryClientInterceptor(tracer)

	err := interceptor(
		context.Background(),
		"/user.v1.UserService/GetUser",
		nil,
		nil,
		nil,
		func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok || len(md.Get("traceparent")) == 0 {
				t.Fatal("expected traceparent in outgoing metadata")
			}
			return wantErr
		},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}

	spans := endedSpans(t, recorder, 1)
	if spans[0].Name() != "/user.v1.UserService/GetUser" {
		t.Fatalf("unexpected span name %q", spans[0].Name())
	}
	if spans[0].SpanKind() != trace.SpanKindClient {
		t.Fatalf("expected client span, got %s", spans[0].SpanKind())
	}
	if spans[0].Status().Code != codes.Error {
		t.Fatalf("expected error status, got %s", spans[0].Status().Code)
	}
}
