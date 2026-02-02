package client

import (
	"fmt"
	"log"
	"time"

	"github.com/gorelov-m-v/go-test-framework/internal/kafka/consumer"
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

type Client struct {
	topicPrefix        string
	buffer             MessageBufferInterface
	backgroundConsumer BackgroundConsumerInterface
	defaultTimeout     time.Duration
	uniqueWindow       time.Duration
	AsyncConfig        config.AsyncConfig
}

func New(cfg Config) (*Client, error) {
	cfg = cfg.Merge()

	if len(cfg.Topics) == 0 {
		return nil, fmt.Errorf("no topics configured. Please specify 'topics' list in kafka config")
	}

	fullTopics := make([]string, len(cfg.Topics))
	for i, topic := range cfg.Topics {
		fullTopics[i] = cfg.TopicPrefix + topic
	}

	buffer := consumer.NewMessageBuffer(fullTopics, cfg.BufferSize)

	consumerCfg := consumer.ConsumerConfig{
		BootstrapServers: cfg.BootstrapServers,
		GroupID:          cfg.GroupID,
		Topics:           fullTopics,
		Version:          cfg.Version,
		SaramaConfig:     cfg.SaramaConfig,
	}

	backgroundConsumer, err := consumer.NewBackgroundConsumer(consumerCfg, buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to create background consumer: %w", err)
	}

	if err := backgroundConsumer.Start(); err != nil {
		return nil, fmt.Errorf("failed to start background consumer: %w", err)
	}

	client := &Client{
		topicPrefix:        cfg.TopicPrefix,
		buffer:             buffer,
		backgroundConsumer: backgroundConsumer,
		defaultTimeout:     cfg.FindMessageTimeout,
		uniqueWindow:       time.Duration(cfg.UniqueDuplicateWindowMs) * time.Millisecond,
		AsyncConfig:        cfg.AsyncConfig,
	}

	// Warmup: wait for consumer to join group and be ready
	if cfg.WarmupTimeout > 0 {
		log.Printf("[Kafka] Waiting for consumer to be ready (timeout: %v)...", cfg.WarmupTimeout)
		if err := backgroundConsumer.WaitReady(cfg.WarmupTimeout); err != nil {
			log.Printf("[Kafka] Warning: consumer warmup failed: %v", err)
		} else {
			log.Println("[Kafka] Consumer is ready")
		}
	}

	return client, nil
}

func (c *Client) Close() error {
	if c.backgroundConsumer != nil {
		return c.backgroundConsumer.Stop()
	}
	return nil
}

func (c *Client) GetDefaultTimeout() time.Duration {
	return c.defaultTimeout
}

func (c *Client) GetUniqueWindow() time.Duration {
	return c.uniqueWindow
}

func (c *Client) GetBackgroundConsumer() BackgroundConsumerInterface {
	return c.backgroundConsumer
}

func (c *Client) GetBuffer() MessageBufferInterface {
	return c.buffer
}

func (c *Client) GetTopicPrefix() string {
	return c.topicPrefix
}

// WaitReady blocks until the consumer has joined the group and is ready to consume.
// This should be called before running tests to ensure Kafka messages can be received.
func (c *Client) WaitReady(timeout time.Duration) error {
	if c.backgroundConsumer == nil {
		return nil
	}
	return c.backgroundConsumer.WaitReady(timeout)
}
