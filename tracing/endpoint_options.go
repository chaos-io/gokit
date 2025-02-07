package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
)

// EndpointOptions holds the options for tracing an endpoint
type EndpointOptions struct {
	// IgnoreBusinessError if set to true will not treat a business error
	// identified through the endpoint.Failer interface as a span error.
	IgnoreBusinessError bool

	// GetOperationName is an optional function that can set the span operation name based on the existing one
	// for the endpoint and information in the context.
	//
	// If the function is nil, or the returned name is empty, the existing name for the endpoint is used.
	GetOperationName func(ctx context.Context, name string) string

	// Attributes holds the default attributes which will be set on span
	// creation by our Endpoint middleware.
	Attributes []attribute.KeyValue

	// GetAttributes is an optional function that can extract attributes
	// from the context and add them to the span.
	GetAttributes func(ctx context.Context) []attribute.KeyValue
}

// EndpointOption allows for functional options to endpoint tracing middleware.
type EndpointOption func(*EndpointOptions)

// WithOptions sets all configuration options at once by use of the EndpointOptions struct.
func WithOptions(options EndpointOptions) EndpointOption {
	return func(o *EndpointOptions) {
		*o = options
	}
}

// WithIgnoreBusinessError if set to true will not treat a business error
// identified through the endpoint.Failer interface as a span error.
func WithIgnoreBusinessError(ignoreBusinessError bool) EndpointOption {
	return func(o *EndpointOptions) {
		o.IgnoreBusinessError = ignoreBusinessError
	}
}

// WithOperationNameFunc allows to set function that can set the span operation name based on the existing one
// for the endpoint and information in the context.
func WithOperationNameFunc(getOperationName func(ctx context.Context, name string) string) EndpointOption {
	return func(o *EndpointOptions) {
		o.GetOperationName = getOperationName
	}
}

// WithAttributes adds default attributes for the spans created by the Endpoint tracer.
func WithAttributes(attributes ...attribute.KeyValue) EndpointOption {
	return func(o *EndpointOptions) {
		o.Attributes = append(o.Attributes, attributes...)
	}
}

// WithAttributesFunc set the func to extracts additional attributes from the context.
func WithAttributesFunc(getAttributes func(ctx context.Context) []attribute.KeyValue) EndpointOption {
	return func(o *EndpointOptions) {
		o.GetAttributes = getAttributes
	}
}
