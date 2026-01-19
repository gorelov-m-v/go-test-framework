package client

import (
	"fmt"
	"time"

	"github.com/gorelov-m-v/go-test-framework/internal/kafka/consumer"
	"github.com/gorelov-m-v/go-test-framework/pkg/config"
	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/types"
)

type Client struct {
	config             *types.Config
	buffer             MessageBufferInterface
	backgroundConsumer BackgroundConsumerInterface
	defaultTimeout     time.Duration
	uniqueWindow       time.Duration
	asyncConfig        config.AsyncConfig
}

func New(cfg types.Config, asyncConfig config.AsyncConfig) (*Client, error) {
	cfg = cfg.Merge()

	if len(cfg.Topics) == 0 {
		return nil, fmt.Errorf("no topics configured. Please specify 'topics' list in kafka config")
	}

	// Apply topic prefix to all topics
	fullTopics := make([]string, len(cfg.Topics))
	for i, topic := range cfg.Topics {
		fullTopics[i] = cfg.TopicPrefix + topic
	}
	cfg.Topics = fullTopics

	buffer := consumer.NewMessageBuffer(cfg.Topics, cfg.BufferSize)

	backgroundConsumer, err := consumer.NewBackgroundConsumer(&cfg, buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to create background consumer: %w", err)
	}

	if err := backgroundConsumer.Start(); err != nil {
		return nil, fmt.Errorf("failed to start background consumer: %w", err)
	}

	client := &Client{
		config:             &cfg,
		buffer:             buffer,
		backgroundConsumer: backgroundConsumer,
		defaultTimeout:     cfg.FindMessageTimeout,
		uniqueWindow:       time.Duration(cfg.UniqueDuplicateWindowMs) * time.Millisecond,
		asyncConfig:        asyncConfig,
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

func (c *Client) GetAsyncConfig() config.AsyncConfig {
	return c.asyncConfig
}

func (c *Client) GetBackgroundConsumer() BackgroundConsumerInterface {
	return c.backgroundConsumer
}

func (c *Client) GetBuffer() MessageBufferInterface {
	return c.buffer
}

func (c *Client) GetTopicPrefix() string {
	return c.config.TopicPrefix
}
