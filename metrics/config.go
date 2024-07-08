package metrics

import (
	"strings"

	"github.com/chaos-io/chaos/config"
	"github.com/chaos-io/chaos/logs"
)

type Config struct {
	Enable     bool   `json:"enable" default:"true"`
	Department string `json:"department"`
	Project    string `json:"project"`
}

func (c *Config) Enabled() bool {
	if c != nil {
		return c.Enable
	}
	return false
}

func NewConfig(path ...string) *Config {
	cfg := &Config{}

	if err := config.ScanFrom(&cfg, "metrics"); err != nil {
		logs.Warnw("failed to get the metrics config from ", "path", strings.Join(path, "."), "error", err)
		return nil
	}

	return cfg
}
