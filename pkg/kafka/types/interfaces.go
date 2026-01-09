package types

import "reflect"

type MessageBufferInterface interface {
	AddMessage(msg *KafkaMessage)

	GetMessages(topicName string) []*KafkaMessage

	IsTopicConfigured(topicName string) bool

	GetConfiguredTopics() []string

	ClearAll()

	ClearTopic(topicName string)
}

type MessageFinderInterface interface {
	SearchAndDeserialize(messages []*KafkaMessage, filters map[string]string, targetType reflect.Type) (interface{}, error)

	FindAndCount(messages []*KafkaMessage, filters map[string]string, targetType reflect.Type) (*FindResult[interface{}], error)

	FindAndCountWithinWindow(messages []*KafkaMessage, filters map[string]string, targetType reflect.Type, windowMs int64) (*FindResult[interface{}], error)

	CountMatching(messages []*KafkaMessage, filters map[string]string) int
}

type BackgroundConsumerInterface interface {
	Start() error

	Stop() error

	FindMessage(filters map[string]string, messageType reflect.Type) (interface{}, error)

	FindAndCount(filters map[string]string, messageType reflect.Type) (*FindResult[interface{}], error)

	FindAndCountWithinWindow(filters map[string]string, messageType reflect.Type, windowMs int64) (*FindResult[interface{}], error)
}
