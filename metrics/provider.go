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
	enabled     bool
	namespace   string
	registerer  prometheus.Registerer
	gatherer    prometheus.Gatherer
	http        *metricmw.HTTPServer
	grpc        *metricmw.GRPCServer
	mu          sync.Mutex
	httpClients map[string]*metricmw.HTTPClient
	grpcClients map[string]*metricmw.GRPCClient
}

func Disabled() *Instrumentation {
	return newInstrumentation(false, "")
}

func New(service string) (*Instrumentation, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	return newInstrumentation(cfg.Enable, metricNamespace(*cfg, service)), nil
}

func metricNamespace(cfg Config, service string) string {
	if cfg.Namespace != "" {
		return cfg.Namespace
	}
	return service
}

func newInstrumentation(enabled bool, namespace string) *Instrumentation {
	namespace = normalizeNamespace(namespace)
	registry := prometheus.NewRegistry()
	return newInstrumentationWithRegistry(enabled, namespace, registry, registry, true)
}

func NewWithRegistry(
	namespace string,
	registerer prometheus.Registerer,
	gatherer prometheus.Gatherer,
) *Instrumentation {
	if registerer == nil {
		registerer = prometheus.NewRegistry()
	}
	if gatherer == nil {
		if value, ok := registerer.(prometheus.Gatherer); ok {
			gatherer = value
		} else {
			gatherer = prometheus.DefaultGatherer
		}
	}
	return newInstrumentationWithRegistry(true, normalizeNamespace(namespace), registerer, gatherer, false)
}

func newInstrumentationWithRegistry(
	enabled bool,
	namespace string,
	registerer prometheus.Registerer,
	gatherer prometheus.Gatherer,
	registerRuntime bool,
) *Instrumentation {
	m := &Instrumentation{
		enabled:     enabled,
		namespace:   namespace,
		registerer:  registerer,
		gatherer:    gatherer,
		httpClients: make(map[string]*metricmw.HTTPClient),
		grpcClients: make(map[string]*metricmw.GRPCClient),
	}
	if !enabled {
		return m
	}
	if registerRuntime {
		registerer.MustRegister(
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
	}
	m.http = metricmw.NewHTTPServer(registerer, namespace)
	m.grpc = metricmw.NewGRPCServer(registerer, namespace)
	return m
}

func (m *Instrumentation) Handler() http.Handler {
	if m == nil || !m.enabled {
		return http.NotFoundHandler()
	}
	return promhttp.HandlerFor(m.gatherer, promhttp.HandlerOpts{})
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
	client := m.httpClients[key]
	if client == nil {
		client = metricmw.NewHTTPClient(m.registerer, m.namespace, name, target)
		m.httpClients[key] = client
	}
	m.mu.Unlock()
	return client.Transport(next)
}

func (m *Instrumentation) GRPCUnaryClientInterceptor(name, target string) grpc.UnaryClientInterceptor {
	if m == nil || !m.enabled {
		return func(
			ctx context.Context,
			method string,
			req, reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
	}
	m.mu.Lock()
	key := name + "\x00" + target
	client := m.grpcClients[key]
	if client == nil {
		client = metricmw.NewGRPCClient(m.registerer, m.namespace, name, target)
		m.grpcClients[key] = client
	}
	m.mu.Unlock()
	return client.UnaryInterceptor()
}
