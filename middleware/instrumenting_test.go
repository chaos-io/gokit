package middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/go-kit/kit/endpoint"
)

func TestInstrumentingAPIReportsEndpointResult(t *testing.T) {
	var gotMethod, gotLayer string
	var gotErr error
	observer := APIObserverFunc(func(method, layer string, err error, _ float64) {
		gotMethod, gotLayer, gotErr = method, layer, err
	})
	wantErr := errors.New("failed")
	next := endpoint.Endpoint(func(context.Context, interface{}) (interface{}, error) {
		return nil, wantErr
	})

	_, _ = InstrumentingAPI("CreateTask", "endpoint", observer)(next)(context.Background(), nil)

	if gotMethod != "CreateTask" || gotLayer != "endpoint" || !errors.Is(gotErr, wantErr) {
		t.Fatalf("observed (%q, %q, %v), want (CreateTask, endpoint, %v)", gotMethod, gotLayer, gotErr, wantErr)
	}
}
