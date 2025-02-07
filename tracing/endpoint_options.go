package tracing

import (
	"fmt"

	"github.com/chaos-io/chaos/logs"
	"go.opentelemetry.io/otel/attribute"
)

// EndpointOption allows you to configure customized options for the tracing middleware.
type EndpointOption func(*traceOptions)

// WithAttributes allows you to configure customized attributes to be added to the span.
func WithAttributes(attrs ...attribute.KeyValue) EndpointOption {
	return func(o *traceOptions) {
		o.attrs = append(o.attrs, attrs...)
	}
}

// traceOptions holds the options for the tracing middleware.
type traceOptions struct {
	attrs []attribute.KeyValue
}

// WithTags adapts the old tags format to attributes.
func WithTags(tags map[string]interface{}) EndpointOption {
	attrs := make([]attribute.KeyValue, 0, len(tags))
	for key, value := range tags {
		// Handle common types.  You might need more sophisticated type handling.
		switch v := value.(type) {
		case string:
			attrs = append(attrs, attribute.String(key, v))
		case int:
			attrs = append(attrs, attribute.Int(key, v))
		case bool:
			attrs = append(attrs, attribute.Bool(key, v))
		case float64:
			attrs = append(attrs, attribute.Float64(key, v))
		case float32:
			attrs = append(attrs, attribute.Float64(key, float64(v)))
		default:
			// Handle other types or log an error
			// For example:
			logs.Warnw("Unsupported tag type", "key", key, "type", fmt.Sprintf("%T", value))
			attrs = append(attrs, attribute.String(key, "unsupported type"))

		}
	}
	return WithAttributes(attrs...)
}
