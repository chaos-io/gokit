package accesslog

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor(cfg Config, options ...Option) grpc.UnaryServerInterceptor {
	policy := newPolicy(cfg)
	opts := buildOptions(options)

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response any, err error) {
		started := time.Now()
		defer func() {
			duration := time.Since(started)
			recovered := recover()
			code := status.Code(err)
			if recovered != nil {
				code = codes.Internal
			}
			logGRPC(ctx, info, code, duration, policy, opts)
			if recovered != nil {
				panic(recovered)
			}
		}()
		return handler(ctx, req)
	}
}

func logGRPC(ctx context.Context, info *grpc.UnaryServerInfo, code codes.Code, duration time.Duration, policy *policy, opts options) {
	method := ""
	if info != nil {
		method = info.FullMethod
	}
	requestID := grpcRequestID(ctx)
	if !policy.ShouldLog(Event{
		Protocol:  ProtocolGRPC,
		Operation: method,
		RequestID: requestID,
		Duration:  duration,
		Important: importantGRPCCode(code),
	}) {
		return
	}

	fields := []any{
		"protocol", ProtocolGRPC,
		"method", method,
		"code", code.String(),
		"duration", duration.String(),
		"duration_ms", float64(duration.Microseconds()) / 1000,
		"remote_ip", grpcRemoteIP(ctx),
		"request_id", requestID,
	}
	fields = append(fields, traceFields(ctx)...)
	opts.log(ctx, grpcLevel(code), "grpc access", fields...)
}

func importantGRPCCode(code codes.Code) bool {
	switch code {
	case codes.Unknown,
		codes.DeadlineExceeded,
		codes.PermissionDenied,
		codes.ResourceExhausted,
		codes.Unavailable,
		codes.Unauthenticated,
		codes.Internal,
		codes.DataLoss:
		return true
	default:
		return false
	}
}

func grpcLevel(code codes.Code) Level {
	switch code {
	case codes.Unknown, codes.Internal, codes.DataLoss:
		return LevelError
	case codes.DeadlineExceeded,
		codes.PermissionDenied,
		codes.ResourceExhausted,
		codes.FailedPrecondition,
		codes.Aborted,
		codes.Unavailable,
		codes.Unauthenticated:
		return LevelWarn
	default:
		return LevelInfo
	}
}
