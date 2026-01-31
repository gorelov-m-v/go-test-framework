package consumer

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"

	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/types"
)

type BackgroundConsumer struct {
	config         *types.Config
	buffer         types.MessageBufferInterface
	consumerGroup  sarama.ConsumerGroup
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	started        bool
	mu             sync.Mutex
	topicPrefix    string
	fullTopicNames []string
	ready          chan struct{}
	readyOnce      sync.Once
}

func NewBackgroundConsumer(
	cfg *types.Config,
	buffer types.MessageBufferInterface,
) (*BackgroundConsumer, error) {
	saramaConfig := sarama.NewConfig()

	version, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Kafka version: %w", err)
	}
	saramaConfig.Version = version

	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	if err := applySaramaConfig(saramaConfig, cfg.SaramaConfig); err != nil {
		return nil, fmt.Errorf("failed to apply SaramaConfig: %w", err)
	}

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
		buffer:         buffer,
		consumerGroup:  consumerGroup,
		ctx:            ctx,
		cancel:         cancel,
		topicPrefix:    "",
		fullTopicNames: fullTopicNames,
		ready:          make(chan struct{}),
	}

	return bc, nil
}

// WaitReady blocks until the consumer has joined the group and is ready to consume,
// or until the timeout expires. Returns nil if ready, error if timeout.
func (bc *BackgroundConsumer) WaitReady(timeout time.Duration) error {
	select {
	case <-bc.ready:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("kafka consumer not ready after %v", timeout)
	case <-bc.ctx.Done():
		return bc.ctx.Err()
	}
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
		buffer:    bc.buffer,
		ready:     bc.ready,
		readyOnce: &bc.readyOnce,
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

			select {
			case <-time.After(time.Second):
			case <-bc.ctx.Done():
				return
			}
		}

		select {
		case <-bc.ctx.Done():
			return
		default:
		}
	}
}

type consumerGroupHandler struct {
	buffer    types.MessageBufferInterface
	ready     chan struct{}
	readyOnce *sync.Once
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	// Signal that consumer is ready (joined group, partitions assigned)
	if h.readyOnce != nil {
		h.readyOnce.Do(func() {
			close(h.ready)
		})
	}
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

func applySaramaConfig(saramaConfig *sarama.Config, userConfig map[string]interface{}) error {
	if userConfig == nil || len(userConfig) == 0 {
		return nil
	}

	configValue := reflect.ValueOf(saramaConfig).Elem()

	for key, value := range userConfig {
		if err := setNestedField(configValue, key, value); err != nil {
			return fmt.Errorf("failed to set field '%s': %w", key, err)
		}
	}

	return nil
}

func setNestedField(structValue reflect.Value, path string, value interface{}) error {
	parts := splitPath(path)

	current := structValue
	for i := 0; i < len(parts)-1; i++ {
		field := current.FieldByName(parts[i])
		if !field.IsValid() {
			return fmt.Errorf("field '%s' not found in path '%s'", parts[i], path)
		}
		current = field
	}

	lastField := current.FieldByName(parts[len(parts)-1])
	if !lastField.IsValid() {
		return fmt.Errorf("field '%s' not found", parts[len(parts)-1])
	}

	if !lastField.CanSet() {
		return fmt.Errorf("field '%s' cannot be set (unexported?)", parts[len(parts)-1])
	}

	return setFieldValue(lastField, value)
}

func setFieldValue(field reflect.Value, value interface{}) error {
	fieldType := field.Type()
	valueRefl := reflect.ValueOf(value)

	if valueRefl.Type().ConvertibleTo(fieldType) {
		field.Set(valueRefl.Convert(fieldType))
		return nil
	}

	switch fieldType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal := reflect.ValueOf(value).Convert(reflect.TypeOf(int64(0))).Int()
		field.SetInt(intVal)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal := reflect.ValueOf(value).Convert(reflect.TypeOf(uint64(0))).Uint()
		field.SetUint(uintVal)
		return nil

	case reflect.Bool:
		field.SetBool(reflect.ValueOf(value).Bool())
		return nil

	case reflect.String:
		field.SetString(reflect.ValueOf(value).String())
		return nil
	}

	return fmt.Errorf("cannot convert %T to %s", value, fieldType)
}

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	return strings.Split(path, ".")
}
