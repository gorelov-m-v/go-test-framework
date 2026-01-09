package consumer

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/IBM/sarama"

	kafkaErrors "go-test-framework/pkg/kafka/errors"
	"go-test-framework/pkg/kafka/types"
)

type BackgroundConsumer struct {
	config         *types.Config
	registry       *types.TopicRegistry
	buffer         types.MessageBufferInterface
	finder         *MessageFinder
	consumerGroup  sarama.ConsumerGroup
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	started        bool
	mu             sync.Mutex
	topicPrefix    string
	fullTopicNames []string
}

func NewBackgroundConsumer(
	cfg *types.Config,
	registry *types.TopicRegistry,
	buffer types.MessageBufferInterface,
	finder *MessageFinder,
) (*BackgroundConsumer, error) {
	saramaConfig := sarama.NewConfig()

	version, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Kafka version: %w", err)
	}
	saramaConfig.Version = version

	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest // Читаем только новые сообщения
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	consumerGroup, err := sarama.NewConsumerGroup(cfg.BootstrapServers, cfg.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	if len(cfg.Topics) == 0 {
		return nil, fmt.Errorf("no topics configured")
	}

	fullTopicNames := cfg.Topics

	ctx, cancel := context.WithCancel(context.Background())

	bc := &BackgroundConsumer{
		config:         cfg,
		registry:       registry,
		buffer:         buffer,
		finder:         finder,
		consumerGroup:  consumerGroup,
		ctx:            ctx,
		cancel:         cancel,
		topicPrefix:    "", // Не используется больше
		fullTopicNames: fullTopicNames,
	}

	return bc, nil
}

func (bc *BackgroundConsumer) Start() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if bc.started {
		return fmt.Errorf("background consumer already started")
	}

	if len(bc.fullTopicNames) == 0 {
		return fmt.Errorf("no topics configured for consumption")
	}

	bc.wg.Add(1)
	go bc.consumeLoop()

	bc.started = true
	return nil
}

func (bc *BackgroundConsumer) Stop() error {
	bc.mu.Lock()
	if !bc.started {
		bc.mu.Unlock()
		return nil
	}
	bc.mu.Unlock()

	bc.cancel()
	bc.wg.Wait()

	if err := bc.consumerGroup.Close(); err != nil {
		return fmt.Errorf("failed to close consumer group: %w", err)
	}

	bc.mu.Lock()
	bc.started = false
	bc.mu.Unlock()

	return nil
}

func (bc *BackgroundConsumer) consumeLoop() {
	defer bc.wg.Done()

	handler := &consumerGroupHandler{
		buffer: bc.buffer,
	}

	for {
		if err := bc.ctx.Err(); err != nil {
			return
		}

		if err := bc.consumerGroup.Consume(bc.ctx, bc.fullTopicNames, handler); err != nil {
			if err == context.Canceled {
				return
			}
			log.Printf("[Kafka] Error from consumer: %v", err)
			time.Sleep(time.Second)
		}

		select {
		case <-bc.ctx.Done():
			return
		default:
		}
	}
}

type consumerGroupHandler struct {
	buffer types.MessageBufferInterface
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg := <-claim.Messages():
			if msg == nil {
				return nil
			}

			kafkaMsg := &types.KafkaMessage{
				Topic:     msg.Topic,
				Partition: msg.Partition,
				Offset:    msg.Offset,
				Key:       msg.Key,
				Value:     msg.Value,
				Timestamp: msg.Timestamp.UnixMilli(),
				Headers:   make(map[string]string),
			}

			for _, header := range msg.Headers {
				kafkaMsg.Headers[string(header.Key)] = string(header.Value)
			}

			h.buffer.AddMessage(kafkaMsg)
			session.MarkMessage(msg, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func (bc *BackgroundConsumer) FindMessage(
	filters map[string]string,
	messageType reflect.Type,
) (interface{}, error) {
	suffix, ok := bc.registry.GetTopicSuffix(messageType)
	if !ok {
		return nil, &kafkaErrors.KafkaTopicNotMappedError{
			MessageType: messageType.Name(),
		}
	}

	fullTopicName := bc.topicPrefix + suffix

	if !bc.buffer.IsTopicConfigured(fullTopicName) {
		return nil, &kafkaErrors.KafkaTopicNotListenedError{
			TopicName:        fullTopicName,
			MessageType:      messageType.Name(),
			ConfiguredTopics: bc.buffer.GetConfiguredTopics(),
		}
	}

	messages := bc.buffer.GetMessages(fullTopicName)

	return bc.finder.SearchAndDeserialize(messages, filters, messageType)
}

func (bc *BackgroundConsumer) FindAndCount(
	filters map[string]string,
	messageType reflect.Type,
) (*types.FindResult[interface{}], error) {
	suffix, ok := bc.registry.GetTopicSuffix(messageType)
	if !ok {
		return nil, &kafkaErrors.KafkaTopicNotMappedError{
			MessageType: messageType.Name(),
		}
	}

	fullTopicName := bc.topicPrefix + suffix

	if !bc.buffer.IsTopicConfigured(fullTopicName) {
		return nil, &kafkaErrors.KafkaTopicNotListenedError{
			TopicName:        fullTopicName,
			MessageType:      messageType.Name(),
			ConfiguredTopics: bc.buffer.GetConfiguredTopics(),
		}
	}

	messages := bc.buffer.GetMessages(fullTopicName)
	return bc.finder.FindAndCount(messages, filters, messageType)
}

func (bc *BackgroundConsumer) FindAndCountWithinWindow(
	filters map[string]string,
	messageType reflect.Type,
	windowMs int64,
) (*types.FindResult[interface{}], error) {
	suffix, ok := bc.registry.GetTopicSuffix(messageType)
	if !ok {
		return nil, &kafkaErrors.KafkaTopicNotMappedError{
			MessageType: messageType.Name(),
		}
	}

	fullTopicName := bc.topicPrefix + suffix

	if !bc.buffer.IsTopicConfigured(fullTopicName) {
		return nil, &kafkaErrors.KafkaTopicNotListenedError{
			TopicName:        fullTopicName,
			MessageType:      messageType.Name(),
			ConfiguredTopics: bc.buffer.GetConfiguredTopics(),
		}
	}

	messages := bc.buffer.GetMessages(fullTopicName)
	return bc.finder.FindAndCountWithinWindow(messages, filters, messageType, windowMs)
}
