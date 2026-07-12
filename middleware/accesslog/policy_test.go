package accesslog

import (
	"fmt"
	"testing"
	"time"
)

func TestPolicyAlwaysLogsImportantAndSlowEvents(t *testing.T) {
	t.Parallel()

	p := newPolicy(Config{
		SlowThreshold: time.Second,
		SampleEvery:   0,
		HTTP:          HTTPConfig{SkipPaths: []string{"/healthz"}},
	})

	tests := []struct {
		name  string
		event Event
		want  bool
	}{
		{name: "ordinary request", event: Event{Protocol: ProtocolHTTP, Operation: "/users"}},
		{name: "successful skipped request", event: Event{Protocol: ProtocolHTTP, Operation: "/healthz"}},
		{name: "important skipped request", event: Event{Protocol: ProtocolHTTP, Operation: "/healthz", Important: true}, want: true},
		{name: "slow request", event: Event{Protocol: ProtocolHTTP, Operation: "/users", Duration: time.Second}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.ShouldLog(tt.event); got != tt.want {
				t.Fatalf("ShouldLog() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPolicyUsesStableRequestIDSampling(t *testing.T) {
	t.Parallel()

	p := newPolicy(Config{SampleEvery: 8})
	foundSampled := false
	for i := 0; i < 100; i++ {
		event := Event{RequestID: fmt.Sprintf("request-%d", i)}
		first := p.ShouldLog(event)
		if second := p.ShouldLog(event); second != first {
			t.Fatalf("sampling changed for the same request ID: first=%v second=%v", first, second)
		}
		foundSampled = foundSampled || first
	}
	if !foundSampled {
		t.Fatal("expected at least one request ID to be sampled")
	}
}

func TestPolicyFallsBackToAtomicCounter(t *testing.T) {
	t.Parallel()

	p := newPolicy(Config{SampleEvery: 2})
	if p.ShouldLog(Event{}) {
		t.Fatal("first event should not be sampled")
	}
	if !p.ShouldLog(Event{}) {
		t.Fatal("second event should be sampled")
	}
}
