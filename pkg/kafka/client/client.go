package client

import (
	"fmt"
	"time"

	"go-test-framework/pkg/config"
	"go-test-framework/pkg/kafka/consumer"
	"go-test-framework/pkg/kafka/types"
)

type Client struct {
	config             *types.Config
	registry           *types.TopicRegistry
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

	buffer := consumer.NewMessageBuffer(cfg.Topics, cfg.BufferSize)

	finder := consumer.NewMessageFinder()

	backgroundConsumer, err := consumer.NewBackgroundConsumer(&cfg, nil, buffer, finder)
	if err != nil {
		return nil, fmt.Errorf("failed to create background consumer: %w", err)
	}

	if err := backgroundConsumer.Start(); err != nil {
		return nil, fmt.Errorf("failed to start background consumer: %w", err)
	}

	client := &Client{
		config:             &cfg,
		registry:           nil,
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

func (c *Client) GetRegistry() *types.TopicRegistry {
	return c.registry
}

func (c *Client) GetBuffer() MessageBufferInterface {
	return c.buffer
}
