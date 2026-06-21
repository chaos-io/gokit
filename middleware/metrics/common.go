package metrics

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

var defaultBuckets = prometheus.DefBuckets

func result(ctx context.Context, err error) string {
	if ctx != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			return "timeout"
		case context.Canceled:
			return "canceled"
		}
	}
	if err != nil {
		return "failure"
	}
	return "success"
}

func httpResult(ctx context.Context, response *http.Response, err error) string {
	if value := result(ctx, err); value != "success" {
		return value
	}
	if response == nil {
		return "failure"
	}
	switch {
	case response.StatusCode >= http.StatusInternalServerError:
		return "server_error"
	case response.StatusCode >= http.StatusBadRequest:
		return "client_error"
	default:
		return "success"
	}
}
