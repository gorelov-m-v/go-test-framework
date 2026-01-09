package dsl

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"go-test-framework/pkg/kafka/client"
	"go-test-framework/pkg/kafka/topic"
)

// Expect создает ожидание сообщения из топика с возможностью проверок полей
// Использование:
//
//	Expect[PlayerEventsTopic](sCtx, client).
//	    With("playerId", "123").
//	    ExpectField("playerName", "John").
//	    ExpectFieldNotEmpty("eventType").
//	    Send()
func Expect[TTopic topic.TopicName](sCtx provider.StepCtx, kafkaClient *client.Client) *Expectation {
	var topicName TTopic
	return NewExpectation(sCtx, kafkaClient, string(topicName))
}
