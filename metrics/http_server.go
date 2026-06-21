package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type RouteResolver func(*http.Request) string

type HTTPServerMetrics struct {
	enabled  bool
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	inflight *prometheus.GaugeVec
}

func NewHTTPServer(provider *Provider) *HTTPServerMetrics {
	m := &HTTPServerMetrics{enabled: provider.Enabled()}
	if !m.enabled {
		return m
	}
	m.requests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: provider.Namespace(), Name: "http_server_requests_total", Help: "HTTP requests completed.",
	}, []string{"method", "route", "status_class"})
	m.duration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: provider.Namespace(), Name: "http_server_request_duration_seconds",
		Help: "HTTP request duration.", Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
	m.inflight = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: provider.Namespace(), Name: "http_server_requests_in_flight", Help: "HTTP requests in flight.",
	}, []string{"method"})
	provider.MustRegister(m.requests, m.duration, m.inflight)
	return m
}

func (m *HTTPServerMetrics) Middleware(resolve RouteResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if !m.enabled {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			rw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			m.inflight.WithLabelValues(r.Method).Inc()
			defer m.inflight.WithLabelValues(r.Method).Dec()
			next.ServeHTTP(rw, r)
			route := "unknown"
			if resolve != nil {
				if value := resolve(r); value != "" {
					route = value
				}
			}
			m.requests.WithLabelValues(r.Method, route, strconv.Itoa(rw.status/100)+"xx").Inc()
			m.duration.WithLabelValues(r.Method, route).Observe(time.Since(started).Seconds())
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
