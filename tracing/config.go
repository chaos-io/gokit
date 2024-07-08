package tracing

import (
	"strings"

	"github.com/chaos-io/chaos/config"
	"github.com/chaos-io/chaos/logs"
)

type Config struct {
	Enable bool    `json:"enable" yaml:"Enable" default:"false"`
	Url    string  `json:"url" yaml:"url" default:"localhost:6831"`
	Param  float64 `json:"param" json:"param" default:"100000"`
}

func NewConfig(path ...string) *Config {
	cfg := &Config{}

	if err := config.ScanFrom(&cfg, "tracing"); err != nil {
		logs.Errorw("failed to get the tracing config from "+strings.Join(path, "."), "error", err)
		return nil
	}
	return cfg
}
