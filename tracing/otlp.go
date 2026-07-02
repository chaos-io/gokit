package tracing

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSDK "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func NewTracer(ctx context.Context, serviceName, otlpEndpoint string) (trace.Tracer, ShutdownFunc, error) {
	return NewTracerWithConfig(ctx, serviceName, &Config{
		Enable:      true,
		Endpoint:    otlpEndpoint,
		SampleRatio: 1,
	})
}

func NewTracerWithConfig(ctx context.Context, serviceName string, cfg *Config) (trace.Tracer, ShutdownFunc, error) {
	provider, err := NewTraceProviderWithConfig(ctx, serviceName, cfg)
	if err != nil {
		return noopTracer(serviceName), NoopShutdown, err
	}

	otel.SetTracerProvider(provider)
	tracer := provider.Tracer(serviceName)
	return tracer, provider.Shutdown, nil
}

func NewTraceProvider(ctx context.Context, serviceName, otlpEndpoint string) (*traceSDK.TracerProvider, error) {
	return NewTraceProviderWithConfig(ctx, serviceName, &Config{
		Enable:      true,
		Endpoint:    otlpEndpoint,
		SampleRatio: 1,
	})
}

func NewTraceProviderWithConfig(ctx context.Context, serviceName string, cfg *Config) (*traceSDK.TracerProvider, error) {
	if cfg == nil {
		cfg = &Config{SampleRatio: 1}
	}
	endpoint := strings.TrimSpace(cfg.Endpoint)
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

	res, err := newResource(ctx, serviceName, cfg)
	if err != nil {
		return nil, err
	}

	traceProvider := traceSDK.NewTracerProvider(
		traceSDK.WithResource(res),
		traceSDK.WithSampler(samplerFromRatio(sampleRatio(cfg))),
		traceSDK.WithBatcher(exp, traceSDK.WithBatchTimeout(time.Second)),
	)

	return traceProvider, nil
}

func sampleRatio(cfg *Config) float64 {
	if cfg == nil || cfg.SampleRatio == 0 {
		return 1
	}
	return cfg.SampleRatio
}

func samplerFromRatio(ratio float64) traceSDK.Sampler {
	if ratio <= 0 {
		return traceSDK.ParentBased(traceSDK.NeverSample())
	}
	if ratio >= 1 {
		return traceSDK.ParentBased(traceSDK.AlwaysSample())
	}
	return traceSDK.ParentBased(traceSDK.TraceIDRatioBased(ratio))
}

func newResource(ctx context.Context, serviceName string, cfg *Config) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(serviceName),
	}
	if cfg != nil {
		if cfg.Environment != "" {
			attrs = append(attrs, attribute.String("deployment.environment", cfg.Environment))
		}
		if cfg.ServiceVersion != "" {
			attrs = append(attrs, attribute.String("service.version", cfg.ServiceVersion))
		}
		if cfg.ServiceInstanceID != "" {
			attrs = append(attrs, attribute.String("service.instance.id", cfg.ServiceInstanceID))
		}
	}
	return resource.New(ctx, resource.WithAttributes(attrs...))
}
