package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc"
)

func TestDisabledProviderReturnsNotFound(t *testing.T) {
	recorder := httptest.NewRecorder()
	newInstrumentation(false, "test").Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
}

func TestEnabledProviderExposesMetrics(t *testing.T) {
	recorder := httptest.NewRecorder()
	newInstrumentation(true, "test").Handler().ServeHTTP(
		recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil),
	)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "go_goroutines") {
		t.Fatal("metrics output does not contain Go runtime metrics")
	}
}

func TestHTTPTransportReusesCollector(t *testing.T) {
	instrumentation := newInstrumentation(true, "test")
	_ = instrumentation.HTTPTransport("registration", "provider", nil)
	_ = instrumentation.HTTPTransport("registration", "provider", nil)
}

func TestGRPCDisabledInterceptorCallsHandler(t *testing.T) {
	called := false
	_, _ = newInstrumentation(false, "test").GRPCUnaryInterceptor()(
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
