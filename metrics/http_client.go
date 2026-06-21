package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type HTTPClientMetrics struct {
	enabled  bool
	name     string
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	inflight *prometheus.GaugeVec
}

func NewHTTPClient(provider *Provider, name string) *HTTPClientMetrics {
	m := &HTTPClientMetrics{enabled: provider.Enabled(), name: name}
	if !m.enabled {
		return m
	}
	m.requests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: provider.Namespace(), Name: "http_client_requests_total", Help: "Outbound HTTP requests completed.",
	}, []string{"client", "host", "method", "result"})
	m.duration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: provider.Namespace(), Name: "http_client_request_duration_seconds",
		Help: "Outbound HTTP request duration.", Buckets: prometheus.DefBuckets,
	}, []string{"client", "host", "method"})
	m.inflight = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: provider.Namespace(), Name: "http_client_requests_in_flight", Help: "Outbound HTTP requests in flight.",
	}, []string{"client"})
	provider.MustRegister(m.requests, m.duration, m.inflight)
	return m
}

func (m *HTTPClientMetrics) Transport(next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	if !m.enabled {
		return next
	}
	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		started := time.Now()
		m.inflight.WithLabelValues(m.name).Inc()
		defer m.inflight.WithLabelValues(m.name).Dec()
		resp, err := next.RoundTrip(req)
		result := clientResult(req.Context(), resp, err)
		m.requests.WithLabelValues(m.name, req.URL.Hostname(), req.Method, result).Inc()
		m.duration.WithLabelValues(m.name, req.URL.Hostname(), req.Method).Observe(time.Since(started).Seconds())
		return resp, err
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func clientResult(ctx context.Context, resp *http.Response, err error) string {
	if ctx.Err() == context.DeadlineExceeded {
		return "timeout"
	}
	if ctx.Err() == context.Canceled {
		return "canceled"
	}
	if err != nil {
		return "failure"
	}
	if resp.StatusCode >= 500 {
		return "server_error"
	}
	if resp.StatusCode >= 400 {
		return "client_error"
	}
	return "success"
}
