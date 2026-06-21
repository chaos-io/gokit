package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Options struct {
	Enabled   bool
	Namespace string
}

type Provider struct {
	enabled    bool
	namespace  string
	registry   *prometheus.Registry
	registerer prometheus.Registerer
}

func New(options Options) *Provider {
	registry := prometheus.NewRegistry()
	provider := &Provider{
		enabled:    options.Enabled,
		namespace:  options.Namespace,
		registry:   registry,
		registerer: registry,
	}
	if provider.enabled {
		provider.MustRegister(
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
	}
	return provider
}

func (p *Provider) Enabled() bool { return p != nil && p.enabled }
func (p *Provider) Namespace() string {
	if p == nil {
		return ""
	}
	return p.namespace
}
func (p *Provider) Registerer() prometheus.Registerer {
	if !p.Enabled() {
		return prometheus.NewRegistry()
	}
	return p.registerer
}
func (p *Provider) Gatherer() prometheus.Gatherer {
	if p == nil {
		return prometheus.NewRegistry()
	}
	return p.registry
}
func (p *Provider) MustRegister(cs ...prometheus.Collector) {
	if p.Enabled() {
		p.registerer.MustRegister(cs...)
	}
}
func (p *Provider) Handler() http.Handler {
	if !p.Enabled() {
		return http.NotFoundHandler()
	}
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}
