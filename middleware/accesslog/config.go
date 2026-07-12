package accesslog

import (
	"time"

	"github.com/chaos-io/chaos/config"
	"github.com/chaos-io/chaos/logs"
)

const configPath = "transport.accessLog"

type Config struct {
	SlowThreshold time.Duration `json:"slowThreshold"`
	SampleEvery   uint64        `json:"sampleEvery"`
	HTTP          HTTPConfig    `json:"http"`
	GRPC          GRPCConfig    `json:"grpc"`
}

type HTTPConfig struct {
	SkipPaths []string `json:"skipPaths"`
}

type GRPCConfig struct {
	SkipMethods []string `json:"skipMethods"`
}

func DefaultConfig() Config {
	return Config{
		SlowThreshold: 500 * time.Millisecond,
		SampleEvery:   100,
		HTTP: HTTPConfig{SkipPaths: []string{
			"/healthz",
			"/readyz",
			"/metrics",
		}},
		GRPC: GRPCConfig{SkipMethods: []string{
			"/grpc.health.v1.Health/Check",
			"/grpc.health.v1.Health/Watch",
		}},
	}
}

func NewConfig() Config {
	cfg := DefaultConfig()
	if err := config.ScanFrom(&cfg, configPath); err != nil {
		logs.Warnw("failed to load access log config", "path", configPath, "error", err)
	}
	return cfg
}
