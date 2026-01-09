package types

import (
	"time"
)

type Config struct {
	BootstrapServers []string `mapstructure:"bootstrapServers"`

	GroupID string `mapstructure:"groupId"`

	// Topics - список топиков для подписки (полные имена)
	Topics []string `mapstructure:"topics"`

	TopicPrefix string `mapstructure:"topicPrefix"` // Deprecated: используйте Topics

	BufferSize int `mapstructure:"bufferSize"`

	FindMessageTimeout time.Duration `mapstructure:"findMessageTimeout"`

	FindMessageSleepInterval time.Duration `mapstructure:"findMessageSleepInterval"`

	UniqueDuplicateWindowMs int64 `mapstructure:"uniqueDuplicateWindowMs"`

	Version string `mapstructure:"version"`

	SaramaConfig map[string]interface{} `mapstructure:"saramaConfig"`
}

func DefaultConfig() Config {
	return Config{
		BufferSize:               1000,
		FindMessageTimeout:       30 * time.Second,
		FindMessageSleepInterval: 200 * time.Millisecond,
		UniqueDuplicateWindowMs:  5000,
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
	if c.Version == "" {
		c.Version = def.Version
	}

	return c
}
