package config

import "time"

type AsyncConfig struct {
	Enabled  bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Timeout  time.Duration `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
	Interval time.Duration `mapstructure:"interval" yaml:"interval" json:"interval"`
	Backoff  BackoffConfig `mapstructure:"backoff" yaml:"backoff" json:"backoff"`
	Jitter   float64       `mapstructure:"jitter" yaml:"jitter" json:"jitter"`
}

type BackoffConfig struct {
	Enabled     bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Factor      float64       `mapstructure:"factor" yaml:"factor" json:"factor"`
	MaxInterval time.Duration `mapstructure:"max_interval" yaml:"max_interval" json:"max_interval"`
}

func DefaultAsyncConfig() AsyncConfig {
	return AsyncConfig{
		Enabled:  true,
		Timeout:  10 * time.Second,
		Interval: 200 * time.Millisecond,
		Backoff: BackoffConfig{
			Enabled:     true,
			Factor:      1.5,
			MaxInterval: 1 * time.Second,
		},
		Jitter: 0.2,
	}
}

// WithDefaults returns the config with default values applied if not set.
// Use this when initializing clients to ensure valid async configuration.
func (ac AsyncConfig) WithDefaults() AsyncConfig {
	if ac.Timeout == 0 {
		return DefaultAsyncConfig()
	}
	return ac
}
