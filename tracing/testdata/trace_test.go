package testdata

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/chaos-io/chaos/logs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/chaos-io/gokit/tracing"
)

const (
	FullServiceName = "full-service-name"
	operationName   = "operation-name"
)

var (
	prop = propagation.TraceContext{}
)

func Test_trace(t *testing.T) {
	_, shutdown := tracing.New(FullServiceName)
	if shutdown != nil {
		defer func() {
			_ = shutdown
		}()
	}

	errCh := make(chan error)

	go httpServer(errCh)

	fmt.Printf("http server get error: %v", <-errCh)
}

func httpServer(errCh chan error) {
	fmt.Println("http server start")
	http.HandleFunc("/api1", traceHandler1)
	http.HandleFunc("/api2", traceHandler2)
	errCh <- http.ListenAndServe(":8089", nil)
}

func traceHandler1(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("trace-handler1")
	ctx, span := tracer.Start(ctx, "trace-handler-span1")
	defer span.End()

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8089/api2", nil)
	if err != nil {
		fmt.Printf("http.NewRequestWithContext error: %v\n", err)
		return
	}

	logs.Debugw("handler http header", "header", req.Header)
	prop.Inject(ctx, propagation.HeaderCarrier(req.Header))
	logs.Debugw("handler1 context", "ctx", trace.SpanContextFromContext(ctx))

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("client.Do error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("resp: %v\n", resp.Status)
}

func traceHandler2(w http.ResponseWriter, r *http.Request) {
	logs.Debugw("handler2 http header", "header", r.Header)
	ctx := r.Context()
	ctx = prop.Extract(ctx, propagation.HeaderCarrier(r.Header))
	logs.Debugw("handle2 ctx2", "ctx", trace.SpanContextFromContext(ctx))

	tracer := otel.Tracer("trace-handler2")
	ctx, span := tracer.Start(ctx, "trace-handler-span2")
	defer span.End()

	fmt.Fprintf(w, "Hello, %s!", r.URL.Path)
}

// curl -v XGET localhost:8089/api1
