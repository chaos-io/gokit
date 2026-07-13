package accesslog

import (
	"testing"
	"time"

	chaosconfig "github.com/chaos-io/chaos/config"
	sourcememory "github.com/chaos-io/chaos/config/source/memory"
)

func TestLoadConfigReadsTransportAccessLog(t *testing.T) {
	err := chaosconfig.InitDefault(
		chaosconfig.WithWatcherDisabled(),
		chaosconfig.WithSource(sourcememory.NewSource(sourcememory.WithJSON([]byte(`{
			"transport": {
				"accessLog": {
					"slowThreshold": "750ms",
					"sampleEvery": 25,
					"http": {"skipPaths": ["/livez"]},
					"grpc": {"skipMethods": ["/grpc.health.v1.Health/Check"]}
				}
			}
		}`)))),
	)
	if err != nil {
		t.Fatal(err)
	}

	cfg := LoadConfig()
	if cfg.SlowThreshold != 750*time.Millisecond {
		t.Fatalf("slow threshold = %v, want 750ms", cfg.SlowThreshold)
	}
	if cfg.SampleEvery != 25 {
		t.Fatalf("sample every = %d, want 25", cfg.SampleEvery)
	}
	if len(cfg.HTTP.SkipPaths) != 1 || cfg.HTTP.SkipPaths[0] != "/livez" {
		t.Fatalf("HTTP skip paths = %v, want [/livez]", cfg.HTTP.SkipPaths)
	}
	if len(cfg.GRPC.SkipMethods) != 1 || cfg.GRPC.SkipMethods[0] != "/grpc.health.v1.Health/Check" {
		t.Fatalf("gRPC skip methods = %v", cfg.GRPC.SkipMethods)
	}
}
