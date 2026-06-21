package metrics

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestPrometheusProviderDisabled(t *testing.T) {
	provider := NewPrometheusProvider(PrometheusOptions{
		Enabled:  false,
		Registry: prometheus.NewRegistry(),
	})

	recorder := httptest.NewRecorder()
	provider.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestAPICollectorUsesBoundedLabels(t *testing.T) {
	provider := NewPrometheusProvider(PrometheusOptions{
		Enabled:   true,
		Namespace: "service",
		Registry:  prometheus.NewRegistry(),
	})
	collector := provider.NewAPICollector()

	collector.Observe("Create", "endpoint", errors.New("boom"), 0.1)

	if got := counterValue(collector.requests.WithLabelValues("Create", "endpoint", "server_error")); got != 1 {
		t.Fatalf("request count = %v, want 1", got)
	}
}

func counterValue(metric prometheus.Counter) float64 {
	var value dto.Metric
	_ = metric.Write(&value)
	return value.GetCounter().GetValue()
}
