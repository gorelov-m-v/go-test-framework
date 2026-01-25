// Deprecated: This file contains backward compatibility aliases.
// All types and functions here will be removed in v2.0.
// Use Query[T] with Consume[T]() instead.
package dsl

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/client"
	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/topic"
)

// Expectation is an alias for Query[any] for backward compatibility.
//
// Deprecated: Use Query[T] with Consume[T] instead.
// This type alias will be removed in v2.0.
type Expectation = Query[any]

// NewExpectation creates a new Kafka query for the specified topic.
//
// Deprecated: Use Consume[TopicType](sCtx, client) instead.
// This function will be removed in v2.0.
func NewExpectation(sCtx provider.StepCtx, kafkaClient *client.Client, topicName string) *Expectation {
	return NewQuery[any](sCtx, kafkaClient, topicName)
}

// Expect creates a typed Query for a topic.
//
// Deprecated: Use Consume[TopicType](sCtx, client) instead.
// This function will be removed in v2.0.
//
// Example migration:
//
//	// Old (deprecated):
//	kafkaDSL.Expect[topics.PlayerEvents](sCtx, kafkaClient).Send()
//
//	// New (recommended):
//	kafkaDSL.Consume[topics.PlayerEvents](sCtx, kafkaClient).Send()
func Expect[TTopic topic.TopicName](sCtx provider.StepCtx, kafkaClient *client.Client) *Query[TTopic] {
	return Consume[TTopic](sCtx, kafkaClient)
}
