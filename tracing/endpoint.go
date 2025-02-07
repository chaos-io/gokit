package tracing

import (
	"context"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd/lb"
)

// TraceEndpoint returns a Middleware that wraps the `next` Endpoint in an
// OpenTelemetry Span called `operationName`.
//
// If `ctx` already has a Span, child span is created from it.
// If `ctx` doesn't yet have a Span, the new one is created.
func TraceEndpoint(tracer trace.Tracer, operationName string, opts ...EndpointOption) endpoint.Middleware {
	cfg := &EndpointOptions{
		Attributes: make([]attribute.KeyValue, 0),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if cfg.GetOperationName != nil {
				if newOperationName := cfg.GetOperationName(ctx, operationName); newOperationName != "" {
					operationName = newOperationName
				}
			}

			ctx, span := tracer.Start(ctx, operationName)
			defer span.End()

			applyAttributes(span, cfg.Attributes)
			if cfg.GetAttributes != nil {
				extraAttributes := cfg.GetAttributes(ctx)
				applyAttributes(span, extraAttributes)
			}

			defer func() {
				if err != nil {
					span.SetStatus(codes.Error, err.Error())
					if lbErr, ok := err.(lb.RetryError); ok {
						// handle errors originating from lb.Retry
						for idx, rawErr := range lbErr.RawErrors {
							span.SetAttributes(attribute.String("gokit.retry.error."+strconv.Itoa(idx+1), rawErr.Error()))
						}

						return
					}

					// generic error
					return
				}

				// test for business error
				if res, ok := response.(endpoint.Failer); ok && res.Failed() != nil {
					span.SetAttributes(attribute.String("gokit.business.error", res.Failed().Error()))

					if cfg.IgnoreBusinessError {
						return
					}

					// treating business error as real error in span.
					span.SetStatus(codes.Error, res.Failed().Error())

					return
				}
			}()

			return next(ctx, request)
		}
	}
}

// TraceServer returns a Middleware that wraps the `next` Endpoint in an
// OpenTelemetry Span called `operationName` with server span.kind tag..
func TraceServer(tracer trace.Tracer, operationName string, opts ...EndpointOption) endpoint.Middleware {
	opts = append(opts, WithAttributes(attribute.String("span.kind", "server")))

	return TraceEndpoint(tracer, operationName, opts...)
}

// TraceClient returns a Middleware that wraps the `next` Endpoint in an
// OpenTelemetry Span called `operationName` with client span.kind tag.
func TraceClient(tracer trace.Tracer, operationName string, opts ...EndpointOption) endpoint.Middleware {
	opts = append(opts, WithAttributes(attribute.String("span.kind", "client")))

	return TraceEndpoint(tracer, operationName, opts...)
}

func applyAttributes(span trace.Span, attributes []attribute.KeyValue) {
	span.SetAttributes(attributes...)
}
