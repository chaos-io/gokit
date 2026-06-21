package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

func TestInstrumentationNormalizesNamespace(t *testing.T) {
	instrumentation := newInstrumentation(true, "9-mailgate.v1 MailgateService")
	instrumentation.HTTPMiddleware(func(*http.Request) string { return "/test" })(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/test", nil))

	recorder := httptest.NewRecorder()
	instrumentation.Handler().ServeHTTP(
		recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil),
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "_9_mailgate_v1_MailgateService_http_server_requests_in_flight") {
		t.Fatal("metrics output does not contain normalized namespace")
	}
}

func TestHTTPTransportReusesCollector(t *testing.T) {
	instrumentation := newInstrumentation(true, "test")
	_ = instrumentation.HTTPTransport("registration", "provider", nil)
	_ = instrumentation.HTTPTransport("registration", "provider", nil)
}
