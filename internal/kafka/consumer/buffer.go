package consumer

import (
	"container/ring"
	"sync"

	"go-test-framework/pkg/kafka/types"
)

type MessageBuffer struct {
	mu         sync.RWMutex
	buffers    map[string]*ringBuffer
	bufferSize int
	topics     []string
}

type ringBuffer struct {
	mu   sync.Mutex
	ring *ring.Ring
	size int
	cap  int
}

func NewMessageBuffer(topicNames []string, bufferSize int) *MessageBuffer {
	if bufferSize <= 0 {
		bufferSize = 1000
	}

	mb := &MessageBuffer{
		buffers:    make(map[string]*ringBuffer),
		bufferSize: bufferSize,
		topics:     make([]string, len(topicNames)),
	}

	copy(mb.topics, topicNames)

	for _, topic := range topicNames {
		mb.buffers[topic] = &ringBuffer{
			ring: ring.New(bufferSize),
			size: 0,
			cap:  bufferSize,
		}
	}

	return mb
}

func (mb *MessageBuffer) AddMessage(msg *types.KafkaMessage) {
	if msg == nil {
		return
	}

	mb.mu.RLock()
	buf, ok := mb.buffers[msg.Topic]
	mb.mu.RUnlock()

	if !ok {
		return
	}

	buf.mu.Lock()
	defer buf.mu.Unlock()

	buf.ring.Value = msg

	buf.ring = buf.ring.Next()

	if buf.size < buf.cap {
		buf.size++
	}
}

func (mb *MessageBuffer) GetMessages(topicName string) []*types.KafkaMessage {
	mb.mu.RLock()
	buf, ok := mb.buffers[topicName]
	mb.mu.RUnlock()

	if !ok {
		return nil
	}

	buf.mu.Lock()
	defer buf.mu.Unlock()

	if buf.size == 0 {
		return nil
	}

	messages := make([]*types.KafkaMessage, 0, buf.size)

	start := buf.ring
	for i := 0; i < buf.size; i++ {
		start = start.Prev()
	}

	current := start
	for i := 0; i < buf.size; i++ {
		if msg, ok := current.Value.(*types.KafkaMessage); ok && msg != nil {
			messages = append(messages, msg)
		}
		current = current.Next()
	}

	return messages
}

func (mb *MessageBuffer) IsTopicConfigured(topicName string) bool {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	_, ok := mb.buffers[topicName]
	return ok
}

func (mb *MessageBuffer) GetConfiguredTopics() []string {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	result := make([]string, len(mb.topics))
	copy(result, mb.topics)
	return result
}

func (mb *MessageBuffer) ClearAll() {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	for _, buf := range mb.buffers {
		buf.mu.Lock()
		buf.ring.Do(func(v interface{}) {
			if v != nil {
				v = nil
			}
		})
		buf.size = 0
		buf.mu.Unlock()
	}
}

func (mb *MessageBuffer) ClearTopic(topicName string) {
	mb.mu.RLock()
	buf, ok := mb.buffers[topicName]
	mb.mu.RUnlock()

	if !ok {
		return
	}

	buf.mu.Lock()
	defer buf.mu.Unlock()

	buf.ring.Do(func(v interface{}) {
		if v != nil {
			v = nil
		}
	})
	buf.size = 0
}
