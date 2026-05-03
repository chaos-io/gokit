package tracing

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSDK "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func NewTracer(ctx context.Context, serviceName, otlpEndpoint string) (trace.Tracer, ShutdownFunc, error) {
	provider, err := NewTraceProvider(ctx, serviceName, otlpEndpoint)
	if err != nil {
		return noopTracer(serviceName), NoopShutdown, err
	}

	otel.SetTracerProvider(provider)
	tracer := provider.Tracer(serviceName)
	return tracer, provider.Shutdown, nil
}

func NewTraceProvider(ctx context.Context, serviceName, otlpEndpoint string) (*traceSDK.TracerProvider, error) {
	endpoint := strings.TrimSpace(otlpEndpoint)
	if endpoint == "" {
		endpoint = DefaultOTLPEndpoint
	}

	options := []otlptracehttp.Option{otlptracehttp.WithInsecure()}
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		options = append(options, otlptracehttp.WithEndpointURL(endpoint))
	} else {
		options = append(options, otlptracehttp.WithEndpoint(endpoint))
	}

	exp, err := otlptracehttp.New(ctx, options...)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(serviceName)))
	if err != nil {
		return nil, err
	}

	traceProvider := traceSDK.NewTracerProvider(
		traceSDK.WithResource(res),
		traceSDK.WithSampler(traceSDK.ParentBased(traceSDK.AlwaysSample())),
		traceSDK.WithBatcher(exp, traceSDK.WithBatchTimeout(time.Second)),
	)

	return traceProvider, nil
}
