package types

type KafkaMessage struct {
	Topic     string
	Partition int32
	Offset    int64
	Key       []byte
	Value     []byte
	Timestamp int64
	Headers   map[string]string
}

type FindResult[T any] struct {
	FirstMatch *T
	AllMatches []T
	Count      int
}
