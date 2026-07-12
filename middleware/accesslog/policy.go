package accesslog

import (
	"sync/atomic"
	"time"
)

const (
	ProtocolHTTP = "http"
	ProtocolGRPC = "grpc"
)

type Event struct {
	Protocol  string
	Operation string
	RequestID string
	Duration  time.Duration
	Important bool
}

type policy struct {
	slowThreshold time.Duration
	sampleEvery   uint64
	httpSkip      map[string]struct{}
	grpcSkip      map[string]struct{}
	counter       atomic.Uint64
}

func newPolicy(cfg Config) *policy {
	return &policy{
		slowThreshold: cfg.SlowThreshold,
		sampleEvery:   cfg.SampleEvery,
		httpSkip:      toSet(cfg.HTTP.SkipPaths),
		grpcSkip:      toSet(cfg.GRPC.SkipMethods),
	}
}

func (p *policy) ShouldLog(event Event) bool {
	if event.Important {
		return true
	}
	if p.skipped(event.Protocol, event.Operation) {
		return false
	}
	if p.slowThreshold > 0 && event.Duration >= p.slowThreshold {
		return true
	}
	if p.sampleEvery == 0 {
		return false
	}
	if p.sampleEvery == 1 {
		return true
	}
	if event.RequestID != "" {
		return hash(event.RequestID)%p.sampleEvery == 0
	}
	return p.counter.Add(1)%p.sampleEvery == 0
}

func (p *policy) skipped(protocol, operation string) bool {
	var paths map[string]struct{}
	switch protocol {
	case ProtocolHTTP:
		paths = p.httpSkip
	case ProtocolGRPC:
		paths = p.grpcSkip
	}
	_, ok := paths[operation]
	return ok
}

func toSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value != "" {
			set[value] = struct{}{}
		}
	}
	return set
}

func hash(value string) uint64 {
	const (
		offset = 14695981039346656037
		prime  = 1099511628211
	)
	h := uint64(offset)
	for i := 0; i < len(value); i++ {
		h ^= uint64(value[i])
		h *= prime
	}
	return h
}
