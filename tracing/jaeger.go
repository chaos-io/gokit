package tracing

import (
	"context"
	"time"

	"github.com/chaos-io/chaos/pkg/logs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSDK "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func NewTracer(ctx context.Context, serviceName, jaegerEndpoint string) (trace.Tracer, func(ctx context.Context) error) {
	provider, err := NewTraceProvider(ctx, serviceName, jaegerEndpoint)
	if err != nil {
		logs.Fatalw("failed to create trace provider", "error", err)
	}

	otel.SetTracerProvider(provider)
	tracer := provider.Tracer(serviceName)
	return tracer, provider.Shutdown
}

func NewTraceProvider(ctx context.Context, serviceName, jaegerEndpoint string) (*traceSDK.TracerProvider, error) {
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(jaegerEndpoint),
		otlptracehttp.WithInsecure())
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(serviceName)))
	if err != nil {
		return nil, err
	}

	traceProvider := traceSDK.NewTracerProvider(
		traceSDK.WithResource(res),
		traceSDK.WithSampler(traceSDK.AlwaysSample()),
		traceSDK.WithBatcher(exp, traceSDK.WithBatchTimeout(time.Second)),
	)

	return traceProvider, nil
}
