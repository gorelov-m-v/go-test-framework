package client

import "go-test-framework/pkg/kafka/types"

type MessageBufferInterface = types.MessageBufferInterface
type MessageFinderInterface = types.MessageFinderInterface
type BackgroundConsumerInterface = types.BackgroundConsumerInterface

type KafkaSetter interface {
	SetKafka(client *Client)
}
