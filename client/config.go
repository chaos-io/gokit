package client

import (
	"strings"

	"github.com/chaos-io/chaos/config"
	"github.com/chaos-io/chaos/logs"

	"github.com/chaos-io/gokit/sd"
)

type Config struct {
	sd.Config
}

func NewConfig(path ...string) *Config {
	cfg := &Config{}
	if err := config.ScanFrom(&cfg, "client"); err != nil {
		logs.Warnw("failed to get the client config from ", "path", strings.Join(path, "."), "error", err)
		return nil
	}
	return cfg
}
