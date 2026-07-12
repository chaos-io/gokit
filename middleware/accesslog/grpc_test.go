package accesslog

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func TestUnaryServerInterceptorLogsStructuredServerError(t *testing.T) {
	t.Parallel()

	logger := &recordingLogger{}
	interceptor := UnaryServerInterceptor(Config{SampleEvery: 0}, WithLogFunc(logger.Log))
	ctx := metadata.NewIncomingContext(contextWithTrace(context.Background()), metadata.Pairs("x-request-id", "req-grpc"))
	ctx = peer.NewContext(ctx, &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP("203.0.113.11"), Port: 4321}})
	wantErr := status.Error(codes.Unavailable, "unavailable")

	response, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{FullMethod: "/user.v1.User/GetUser"}, func(context.Context, any) (any, error) {
		return "response", wantErr
	})
	if response != "response" || err != wantErr {
		t.Fatalf("response, err = %v, %v; want response, %v", response, err, wantErr)
	}

	entry := logger.single(t)
	if entry.Level != LevelWarn {
		t.Fatalf("level = %v, want %v", entry.Level, LevelWarn)
	}
	assertField(t, entry.Fields, "protocol", "grpc")
	assertField(t, entry.Fields, "method", "/user.v1.User/GetUser")
	assertField(t, entry.Fields, "code", codes.Unavailable.String())
	assertField(t, entry.Fields, "remote_ip", "203.0.113.11")
	assertField(t, entry.Fields, "request_id", "req-grpc")
	assertField(t, entry.Fields, "trace_id", testTraceID.String())
}

func TestUnaryServerInterceptorSkipsSuccessfulHealthMethod(t *testing.T) {
	t.Parallel()

	logger := &recordingLogger{}
	interceptor := UnaryServerInterceptor(Config{
		SampleEvery: 1,
		GRPC:        GRPCConfig{SkipMethods: []string{"/grpc.health.v1.Health/Check"}},
	}, WithLogFunc(logger.Log))

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/grpc.health.v1.Health/Check"}, func(context.Context, any) (any, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := logger.count(); got != 0 {
		t.Fatalf("log count = %d, want 0", got)
	}
}

func TestUnaryServerInterceptorLogsSlowSuccessfulCall(t *testing.T) {
	t.Parallel()

	logger := &recordingLogger{}
	interceptor := UnaryServerInterceptor(Config{SlowThreshold: time.Nanosecond}, WithLogFunc(logger.Log))
	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/user.v1.User/ListUsers"}, func(context.Context, any) (any, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if entry := logger.single(t); entry.Level != LevelInfo {
		t.Fatalf("level = %v, want %v", entry.Level, LevelInfo)
	}
}

func TestUnaryServerInterceptorLogsAndPropagatesPanic(t *testing.T) {
	t.Parallel()

	logger := &recordingLogger{}
	interceptor := UnaryServerInterceptor(Config{SampleEvery: 0}, WithLogFunc(logger.Log))

	defer func() {
		if recovered := recover(); recovered != "boom" {
			t.Fatalf("recovered = %v, want boom", recovered)
		}
		if entry := logger.single(t); entry.Level != LevelError {
			t.Fatalf("level = %v, want %v", entry.Level, LevelError)
		}
	}()
	_, _ = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/user.v1.User/GetUser"}, func(context.Context, any) (any, error) {
		panic("boom")
	})
}
