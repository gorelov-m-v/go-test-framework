package errors

import "fmt"

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

type KafkaTopicNotListenedError struct {
	TopicName        string
	MessageType      string
	ConfiguredTopics []string
}

func (e *KafkaTopicNotListenedError) Error() string {
	return fmt.Sprintf("Topic '%s' (for type %s) is not configured to be listened to. Configured topics: %v",
		e.TopicName, e.MessageType, e.ConfiguredTopics)
}
