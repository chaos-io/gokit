package metrics

import (
	"context"
	"database/sql"
	"net/http"
	"sync"

	metricmw "github.com/chaos-io/gokit/middleware/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

type Instrumentation struct {
	enabled   bool
	namespace string
	registry  *prometheus.Registry
	http      *metricmw.HTTPServer
	grpc      *metricmw.GRPCServer
	mu        sync.Mutex
	clients   map[string]*metricmw.HTTPClient
}

func New(service string) (*Instrumentation, error) {
	cfg := NewConfig("metrics")
	enabled, namespace := false, service
	if cfg != nil {
		enabled = cfg.Enabled()
		namespace = cfg.Project
		if cfg.Department != "" {
			namespace = cfg.Department + "_" + cfg.Project
		}
	}
	registry := prometheus.NewRegistry()
	m := &Instrumentation{
		enabled: enabled, namespace: namespace, registry: registry,
		clients: make(map[string]*metricmw.HTTPClient),
	}
	if !enabled {
		return m, nil
	}
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	m.http = metricmw.NewHTTPServer(registry, namespace)
	m.grpc = metricmw.NewGRPCServer(registry, namespace)
	return m, nil
}

func (m *Instrumentation) Handler() http.Handler {
	if m == nil || !m.enabled {
		return http.NotFoundHandler()
	}
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *Instrumentation) HTTPMiddleware(resolve metricmw.RouteResolver) func(http.Handler) http.Handler {
	if m == nil || !m.enabled {
		return func(next http.Handler) http.Handler { return next }
	}
	return m.http.Middleware(resolve)
}

func (m *Instrumentation) GRPCUnaryInterceptor() grpc.UnaryServerInterceptor {
	if m == nil || !m.enabled {
		return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, req)
		}
	}
	return m.grpc.UnaryInterceptor()
}

func (m *Instrumentation) HTTPTransport(name string, next http.RoundTripper) http.RoundTripper {
	if m == nil || !m.enabled {
		if next == nil {
			return http.DefaultTransport
		}
		return next
	}
	m.mu.Lock()
	client := m.clients[name]
	if client == nil {
		client = metricmw.NewHTTPClient(m.registry, m.namespace, name)
		m.clients[name] = client
	}
	m.mu.Unlock()
	return client.Transport(next)
}

func (m *Instrumentation) RegisterDB(name string, db *sql.DB) {
	if m != nil && m.enabled && db != nil {
		m.registry.MustRegister(collectors.NewDBStatsCollector(db, name))
	}
}
