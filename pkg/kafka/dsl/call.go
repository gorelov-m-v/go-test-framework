package dsl

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/client"
	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/topic"
)

func Expect[TTopic topic.TopicName](sCtx provider.StepCtx, kafkaClient *client.Client) *Expectation {
	var topicName TTopic
	fullTopicName := kafkaClient.GetTopicPrefix() + topicName.TopicName()
	return NewExpectation(sCtx, kafkaClient, fullTopicName)
}
