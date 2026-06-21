package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"google.golang.org/grpc"
)

func TestDisabledProviderReturnsNotFound(t *testing.T) {
	recorder := httptest.NewRecorder()
	New(Options{}).Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
}

func TestOperationFinishIsIdempotent(t *testing.T) {
	provider := New(Options{Enabled: true, Namespace: "test"})
	collector := NewOperation(provider, "job", []string{"kind"}, prometheus.DefBuckets)
	operation := collector.Start(context.Background(), "sync")
	operation.Finish(nil)
	operation.Finish(nil)

	families, err := provider.Gatherer().Gather()
	if err != nil {
		t.Fatal(err)
	}
	if value := counterMetricValue(families, "test_job_attempts_total"); value != 1 {
		t.Fatalf("attempts = %v, want 1", value)
	}
}

func TestGRPCMethodSplit(t *testing.T) {
	service, method := splitGRPCMethod("/mailgate.v1.MailgateService/CreateTask")
	if service != "mailgate.v1.MailgateService" || method != "CreateTask" {
		t.Fatalf("split = %q %q", service, method)
	}
}

func TestGRPCDisabledInterceptorCallsHandler(t *testing.T) {
	called := false
	_, _ = NewGRPCServer(New(Options{})).UnaryInterceptor()(
		context.Background(), nil, &grpc.UnaryServerInfo{},
		func(context.Context, any) (any, error) {
			called = true
			return nil, nil
		},
	)
	if !called {
		t.Fatal("handler was not called")
	}
}

func counterMetricValue(families []*dto.MetricFamily, name string) float64 {
	for _, family := range families {
		if family.GetName() == name && len(family.Metric) > 0 {
			return family.Metric[0].GetCounter().GetValue()
		}
	}
	return 0
}
