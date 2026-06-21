package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type OperationCollector struct {
	enabled  bool
	labels   []string
	attempts *prometheus.CounterVec
	duration *prometheus.HistogramVec
	inflight *prometheus.GaugeVec
}

func NewOperation(provider *Provider, subsystem string, labels []string, buckets []float64) *OperationCollector {
	m := &OperationCollector{enabled: provider.Enabled(), labels: append([]string(nil), labels...)}
	if !m.enabled {
		return m
	}
	resultLabels := append(append([]string(nil), labels...), "result")
	m.attempts = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: provider.Namespace(), Subsystem: subsystem, Name: "attempts_total", Help: "Operations completed.",
	}, resultLabels)
	m.duration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: provider.Namespace(), Subsystem: subsystem, Name: "duration_seconds",
		Help: "Operation duration.", Buckets: buckets,
	}, resultLabels)
	m.inflight = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: provider.Namespace(), Subsystem: subsystem, Name: "in_flight", Help: "Operations in flight.",
	}, labels)
	provider.MustRegister(m.attempts, m.duration, m.inflight)
	return m
}

func (m *OperationCollector) Start(ctx context.Context, values ...string) *Operation {
	op := &Operation{collector: m, ctx: ctx, values: append([]string(nil), values...), started: time.Now()}
	if m.enabled {
		m.inflight.WithLabelValues(values...).Inc()
	}
	return op
}

type Operation struct {
	once      sync.Once
	collector *OperationCollector
	ctx       context.Context
	values    []string
	started   time.Time
}

func (o *Operation) Finish(err error) {
	o.once.Do(func() {
		if !o.collector.enabled {
			return
		}
		result := Result(o.ctx, err)
		o.collector.inflight.WithLabelValues(o.values...).Dec()
		values := append(append([]string(nil), o.values...), result)
		o.collector.attempts.WithLabelValues(values...).Inc()
		o.collector.duration.WithLabelValues(values...).Observe(time.Since(o.started).Seconds())
	})
}

func Result(ctx context.Context, err error) string {
	if ctx != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "timeout"
		}
		if ctx.Err() == context.Canceled {
			return "canceled"
		}
	}
	if err != nil {
		return "failure"
	}
	return "success"
}
