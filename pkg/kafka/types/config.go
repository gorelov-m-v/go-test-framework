package types

import (
	"time"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
)

type Config struct {
	AsyncConfig config.AsyncConfig `mapstructure:"async"`
	BootstrapServers []string `mapstructure:"bootstrapServers"`

	GroupID string `mapstructure:"groupId"`

	Topics []string `mapstructure:"topics"`

	TopicPrefix string `mapstructure:"topicPrefix"`

	BufferSize int `mapstructure:"bufferSize"`

	FindMessageTimeout time.Duration `mapstructure:"findMessageTimeout"`

	FindMessageSleepInterval time.Duration `mapstructure:"findMessageSleepInterval"`

	UniqueDuplicateWindowMs int64 `mapstructure:"uniqueDuplicateWindowMs"`

	// WarmupTimeout is the maximum time to wait for consumer to join group and be ready.
	// Set to 0 to disable warmup. Default: 60s
	WarmupTimeout time.Duration `mapstructure:"warmupTimeout"`

	Version string `mapstructure:"version"`

	SaramaConfig map[string]interface{} `mapstructure:"saramaConfig"`
}

func DefaultConfig() Config {
	return Config{
		BufferSize:               1000,
		FindMessageTimeout:       30 * time.Second,
		FindMessageSleepInterval: 200 * time.Millisecond,
		UniqueDuplicateWindowMs:  5000,
		WarmupTimeout:            60 * time.Second,
		Version:                  "2.6.0",
	}
}

func (c Config) Merge() Config {
	def := DefaultConfig()

	if c.BufferSize == 0 {
		c.BufferSize = def.BufferSize
	}
	if c.FindMessageTimeout == 0 {
		c.FindMessageTimeout = def.FindMessageTimeout
	}
	if c.FindMessageSleepInterval == 0 {
		c.FindMessageSleepInterval = def.FindMessageSleepInterval
	}
	if c.UniqueDuplicateWindowMs == 0 {
		c.UniqueDuplicateWindowMs = def.UniqueDuplicateWindowMs
	}
	if c.WarmupTimeout == 0 {
		c.WarmupTimeout = def.WarmupTimeout
	}
	if c.Version == "" {
		c.Version = def.Version
	}

	return c
}
