package tracing

import (
	"strings"

	"github.com/chaos-io/chaos/config"
	"github.com/chaos-io/chaos/logs"
)

type Config struct {
	Enable   bool   `json:"enable" yaml:"enable" default:"false"`
	Endpoint string `json:"endpoint" yaml:"endpoint" default:"localhost:4318"`
}

const DefaultOTLPEndpoint = "localhost:4318"

func NewConfig(path ...string) *Config {
	cfg := &Config{}
	if err := config.ScanFrom(&cfg, "tracing"); err != nil {
		logs.Errorw("failed to get the tracing config from "+strings.Join(path, "."), "error", err)
		return nil
	}
	return cfg
}
