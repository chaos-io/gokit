package metrics

import "github.com/chaos-io/chaos/config"

type Config struct {
	Enable     bool   `json:"enable" yaml:"enable" default:"true"`
	Department string `json:"department" yaml:"department"`
	Project    string `json:"project" yaml:"project"`
}

func loadConfig() (*Config, error) {
	cfg := &Config{}
	if err := config.ScanFrom(cfg, "metrics"); err != nil {
		return nil, err
	}
	return cfg, nil
}
