package types

import "time"

type MessageBufferInterface interface {
	AddMessage(msg *KafkaMessage)

	GetMessages(topicName string) []*KafkaMessage

	IsTopicConfigured(topicName string) bool

	GetConfiguredTopics() []string

	ClearAll()

	ClearTopic(topicName string)
}

type BackgroundConsumerInterface interface {
	Start() error

	Stop() error

	// WaitReady blocks until the consumer has joined the group and is ready to consume,
	// or until the timeout expires.
	WaitReady(timeout time.Duration) error
}
