package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc"
)

func TestDisabledProviderReturnsNotFound(t *testing.T) {
	recorder := httptest.NewRecorder()
	(&Instrumentation{}).Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
}

func TestGRPCDisabledInterceptorCallsHandler(t *testing.T) {
	called := false
	_, _ = (&Instrumentation{}).GRPCUnaryInterceptor()(
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
