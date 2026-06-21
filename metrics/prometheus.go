package metrics

import (
	"errors"
	"net/http"

	"github.com/chaos-io/core/go/chaos/core"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusOptions struct {
	Enabled   bool
	Namespace string
	Registry  *prometheus.Registry
}

type PrometheusProvider struct {
	enabled   bool
	namespace string
	registry  *prometheus.Registry
}

func NewPrometheusProvider(options PrometheusOptions) *PrometheusProvider {
	if options.Registry == nil {
		options.Registry = prometheus.NewRegistry()
	}
	return &PrometheusProvider{
		enabled:   options.Enabled,
		namespace: options.Namespace,
		registry:  options.Registry,
	}
}

func (p *PrometheusProvider) Enabled() bool {
	return p != nil && p.enabled
}

func (p *PrometheusProvider) Namespace() string {
	if p == nil {
		return ""
	}
	return p.namespace
}

func (p *PrometheusProvider) Register(collectors ...prometheus.Collector) {
	if !p.Enabled() {
		return
	}
	p.registry.MustRegister(collectors...)
}

func (p *PrometheusProvider) Handler() http.Handler {
	if !p.Enabled() {
		return http.NotFoundHandler()
	}
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}

type APICollector struct {
	enabled  bool
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func (p *PrometheusProvider) NewAPICollector() *APICollector {
	collector := &APICollector{
		enabled: p.Enabled(),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: p.Namespace(),
			Name:      "api_requests_total",
			Help:      "API requests completed.",
		}, []string{"method", "layer", "code_class"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: p.Namespace(),
			Name:      "api_request_duration_seconds",
			Help:      "API request duration in seconds.",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		}, []string{"method", "layer", "code_class"}),
	}

	p.Register(collector.requests, collector.duration)

	return collector
}

func (c *APICollector) Observe(method, layer string, err error, seconds float64) {
	if c == nil || !c.enabled {
		return
	}

	class := ErrorClass(err)
	c.requests.WithLabelValues(method, layer, class).Inc()
	c.duration.WithLabelValues(method, layer, class).Observe(seconds)
}

func ErrorClass(err error) string {
	if err == nil {
		return "success"
	}

	var e *core.Error
	if errors.As(err, &e) && e.StatusCode() >= 400 && e.StatusCode() < 500 {
		return "client_error"
	}

	return "server_error"
}
