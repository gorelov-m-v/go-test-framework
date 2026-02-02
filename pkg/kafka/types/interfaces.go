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
	WaitReady(timeout time.Duration) error
}
