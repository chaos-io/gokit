package tracing

import (
	"context"

	"github.com/chaos-io/chaos/pkg/logs"
	"go.opentelemetry.io/otel/trace"
)

func New(name string) (trace.Tracer, func(context.Context) error) {
	return NewWith(name, NewConfig("tracing"))
}

func NewWith(name string, cfg *Config) (trace.Tracer, func(context.Context) error) {
	if cfg == nil {
		logs.Fatal("failed to create the tracer coz of given nil config")
		return nil, nil
	}

	if cfg.Enable {
		return NewTracer(context.Background(), name, cfg.Url)
	}

	return nil, nil
}
