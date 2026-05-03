package tracing

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func New(name string) (trace.Tracer, func(context.Context) error, error) {
	return NewWith(context.Background(), name, NewConfig("tracing"))
}

func NewWith(ctx context.Context, name string, cfg *Config) (trace.Tracer, func(context.Context) error, error) {
	if cfg == nil {
		return nil, nil, errors.New("tracing config is nil")
	}

	if cfg.Enable {
		return NewTracer(ctx, name, cfg.Endpoint)
	}

	return otel.Tracer(name), func(context.Context) error { return nil }, nil
}
