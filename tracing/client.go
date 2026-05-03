package tracing

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/trace"
)

func New(name string) (trace.Tracer, ShutdownFunc, error) {
	return NewWith(context.Background(), name, NewConfig("tracing"))
}

func NewWith(ctx context.Context, name string, cfg *Config) (trace.Tracer, ShutdownFunc, error) {
	if cfg == nil {
		return noopTracer(name), NoopShutdown, errors.New("tracing config is nil")
	}

	if cfg.Enable {
		return NewTracer(ctx, name, cfg.Endpoint)
	}

	return noopTracer(name), NoopShutdown, nil
}
