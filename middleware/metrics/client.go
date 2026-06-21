package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type HTTPClient struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	inflight *prometheus.GaugeVec
}

func NewHTTPClient(registerer prometheus.Registerer, namespace, name, target string) *HTTPClient {
	registerer = prometheus.WrapRegistererWith(prometheus.Labels{
		"client": name, "target": target,
	}, registerer)
	m := &HTTPClient{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace, Name: "http_client_requests_total", Help: "Outbound HTTP requests completed.",
		}, []string{"method", "result"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace, Name: "http_client_request_duration_seconds",
			Help: "Outbound HTTP request duration.", Buckets: defaultBuckets,
		}, []string{"method"}),
		inflight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace, Name: "http_client_requests_in_flight", Help: "Outbound HTTP requests in flight.",
		}, nil),
	}
	registerer.MustRegister(m.requests, m.duration, m.inflight)
	return m
}

func (m *HTTPClient) Transport(next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		started := time.Now()
		m.inflight.WithLabelValues().Inc()
		defer m.inflight.WithLabelValues().Dec()

		response, err := next.RoundTrip(req)
		m.requests.WithLabelValues(
			req.Method, httpResult(req.Context(), response, err),
		).Inc()
		m.duration.WithLabelValues(
			req.Method,
		).Observe(time.Since(started).Seconds())
		return response, err
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
