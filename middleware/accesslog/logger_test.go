package accesslog

import (
	"context"
	"sync"
	"testing"
)

type logEntry struct {
	Level  level
	Fields map[string]any
}

type recordingLogger struct {
	mu      sync.Mutex
	entries []logEntry
}

func (l *recordingLogger) Log(_ context.Context, level level, _ string, fields ...any) {
	entry := logEntry{Level: level, Fields: make(map[string]any, len(fields)/2)}
	for i := 0; i+1 < len(fields); i += 2 {
		key, _ := fields[i].(string)
		entry.Fields[key] = fields[i+1]
	}
	l.mu.Lock()
	l.entries = append(l.entries, entry)
	l.mu.Unlock()
}

func (l *recordingLogger) count() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries)
}

func (l *recordingLogger) single(t *testing.T) logEntry {
	t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.entries) != 1 {
		t.Fatalf("log count = %d, want 1", len(l.entries))
	}
	return l.entries[0]
}

func assertField(t *testing.T, fields map[string]any, key string, want any) {
	t.Helper()
	if got := fields[key]; got != want {
		t.Fatalf("field %q = %#v, want %#v", key, got, want)
	}
}
