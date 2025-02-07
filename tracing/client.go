package tracing

import (
	"context"

	"go.opentelemetry.io/otel"

	"github.com/chaos-io/chaos/logs"
)

func New(name string) (func(context.Context) error, error) {
	return NewWith(name, NewConfig("tracing"))
}

func NewWith(name string, cfg *Config) (func(context.Context) error, error) {
	if cfg == nil {
		logs.Fatal("failed to create the tracer coz of given nil config")
		return nil, nil
	}

	if cfg.Enable {
		ctx := context.Background()
		tracerProvider, err := NewJaegerTraceProvider(ctx, name, cfg.Url)
		if err != nil {
			logs.Errorw("failed to create the tracing", "name", name, "host", cfg.Url, "error", err)
			return nil, err
		}

		otel.SetTracerProvider(tracerProvider)
		return tracerProvider.Shutdown, nil
	}

	return nil, nil
}
