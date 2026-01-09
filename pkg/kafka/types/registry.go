package types

import (
	"reflect"
	"sync"
)

type TopicRegistry struct {
	mu       sync.RWMutex
	mappings map[reflect.Type]string
}

func NewTopicRegistry() *TopicRegistry {
	return &TopicRegistry{
		mappings: make(map[reflect.Type]string),
	}
}

func Register[T any](registry *TopicRegistry, topicSuffix string) {
	var zero T
	typ := reflect.TypeOf(zero)
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.mappings[typ] = topicSuffix
}

func (r *TopicRegistry) GetTopicSuffix(messageType reflect.Type) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	suffix, ok := r.mappings[messageType]
	return suffix, ok
}

func (r *TopicRegistry) GetAllTopicSuffixes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	suffixes := make([]string, 0, len(r.mappings))
	seen := make(map[string]bool)

	for _, suffix := range r.mappings {
		if suffix != "" && !seen[suffix] {
			suffixes = append(suffixes, suffix)
			seen[suffix] = true
		}
	}

	return suffixes
}
