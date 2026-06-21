package metrics

import (
	"context"
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

func Disabled() *Instrumentation {
	return newInstrumentation(false, "")
}

func New(service string) (*Instrumentation, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	namespace := cfg.Project
	if namespace == "" {
		namespace = service
	}

	if cfg.Department != "" {
		namespace = cfg.Department + "_" + namespace
	}

	return newInstrumentation(cfg.Enable, namespace), nil
}

func newInstrumentation(enabled bool, namespace string) *Instrumentation {
	registry := prometheus.NewRegistry()
	m := &Instrumentation{
		enabled:   enabled,
		namespace: namespace,
		registry:  registry,
		clients:   make(map[string]*metricmw.HTTPClient),
	}
	if !enabled {
		return m
	}
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	m.http = metricmw.NewHTTPServer(registry, namespace)
	m.grpc = metricmw.NewGRPCServer(registry, namespace)
	return m
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

func (m *Instrumentation) HTTPTransport(name, target string, next http.RoundTripper) http.RoundTripper {
	if m == nil || !m.enabled {
		if next == nil {
			return http.DefaultTransport
		}
		return next
	}
	m.mu.Lock()
	key := name + "\x00" + target
	client := m.clients[key]
	if client == nil {
		client = metricmw.NewHTTPClient(m.registry, m.namespace, name, target)
		m.clients[key] = client
	}
	m.mu.Unlock()
	return client.Transport(next)
}
