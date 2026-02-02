package types

import (
	"time"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
)

type Config struct {
	AsyncConfig              config.AsyncConfig     `mapstructure:"async" yaml:"async" json:"async"`
	BootstrapServers         []string               `mapstructure:"bootstrapServers" yaml:"bootstrapServers" json:"bootstrapServers"`
	GroupID                  string                 `mapstructure:"groupId" yaml:"groupId" json:"groupId"`
	Topics                   []string               `mapstructure:"topics" yaml:"topics" json:"topics"`
	TopicPrefix              string                 `mapstructure:"topicPrefix" yaml:"topicPrefix" json:"topicPrefix"`
	BufferSize               int                    `mapstructure:"bufferSize" yaml:"bufferSize" json:"bufferSize"`
	FindMessageTimeout       time.Duration          `mapstructure:"findMessageTimeout" yaml:"findMessageTimeout" json:"findMessageTimeout"`
	FindMessageSleepInterval time.Duration          `mapstructure:"findMessageSleepInterval" yaml:"findMessageSleepInterval" json:"findMessageSleepInterval"`
	UniqueDuplicateWindowMs  int64                  `mapstructure:"uniqueDuplicateWindowMs" yaml:"uniqueDuplicateWindowMs" json:"uniqueDuplicateWindowMs"`
	WarmupTimeout            time.Duration          `mapstructure:"warmupTimeout" yaml:"warmupTimeout" json:"warmupTimeout"`
	Version                  string                 `mapstructure:"version" yaml:"version" json:"version"`
	SaramaConfig             map[string]interface{} `mapstructure:"saramaConfig" yaml:"saramaConfig" json:"saramaConfig"`
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
