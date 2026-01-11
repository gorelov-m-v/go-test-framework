package errors

import "fmt"

type KafkaMessageNotFoundError struct {
	MessageType string
	Filters     map[string]string
	Timeout     string
}

func (e *KafkaMessageNotFoundError) Error() string {
	return fmt.Sprintf("Kafka message %s not found within %s. Filters: %v", e.MessageType, e.Timeout, e.Filters)
}

type KafkaMessageNotUniqueError struct {
	MessageType string
	Filters     map[string]string
	Count       int
	WindowMs    int64
}

func (e *KafkaMessageNotUniqueError) Error() string {
	return fmt.Sprintf("Kafka message %s expected once but found %d within %dms window. Filters: %v",
		e.MessageType, e.Count, e.WindowMs, e.Filters)
}

type KafkaDeserializationError struct {
	Topic       string
	Offset      int64
	MessageType string
	Err         error
}

func (e *KafkaDeserializationError) Error() string {
	return fmt.Sprintf("Failed to deserialize Kafka message (Topic: %s, Offset: %d) into %s: %v",
		e.Topic, e.Offset, e.MessageType, e.Err)
}

func (e *KafkaDeserializationError) Unwrap() error {
	return e.Err
}

type KafkaTopicNotMappedError struct {
	MessageType string
}

func (e *KafkaTopicNotMappedError) Error() string {
	return fmt.Sprintf("No topic suffix configured for message type %s", e.MessageType)
}

type KafkaTopicNotListenedError struct {
	TopicName        string
	MessageType      string
	ConfiguredTopics []string
}

func (e *KafkaTopicNotListenedError) Error() string {
	return fmt.Sprintf("Topic '%s' (for type %s) is not configured to be listened to. Configured topics: %v",
		e.TopicName, e.MessageType, e.ConfiguredTopics)
}
