package dsl

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/client"
	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/topic"
	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/types"
)

func Expect[TTopic topic.TopicName](sCtx provider.StepCtx, kafkaClient *client.Client) *Expectation {
	var topicName TTopic
	return NewExpectation(sCtx, kafkaClient, topicName.TopicName())
}

func Register[T any](kafkaClient *client.Client, topicSuffix string) {
	registry := kafkaClient.GetRegistry()
	if registry != nil {
		types.Register[T](registry, topicSuffix)
	}
}
