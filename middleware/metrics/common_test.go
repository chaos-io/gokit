package metrics

import (
	"context"
	"net/http"
	"testing"
)

func TestHTTPResultClassifiesStatus(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{status: http.StatusOK, want: "success"},
		{status: http.StatusBadRequest, want: "client_error"},
		{status: http.StatusServiceUnavailable, want: "server_error"},
	}
	for _, tt := range tests {
		response := &http.Response{StatusCode: tt.status}
		if got := httpResult(context.Background(), response, nil); got != tt.want {
			t.Fatalf("status %d result = %q, want %q", tt.status, got, tt.want)
		}
	}
}
