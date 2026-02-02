package consumer

type ConsumerConfig struct {
	BootstrapServers []string
	GroupID          string
	Topics           []string
	Version          string
	SaramaConfig     map[string]interface{}
}
