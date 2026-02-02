package consumer

type ConsumerConfig struct {
	BootstrapServers []string
	GroupID          string
	Topics           []string
	Version          string
	StartFromNewest  bool
	SkipExisting     bool
	SaramaConfig     map[string]interface{}
}
