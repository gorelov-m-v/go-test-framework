package client

import "time"

// Result represents a Redis operation result
type Result struct {
	Key      string
	Value    string
	Exists   bool
	TTL      time.Duration
	Error    error
	Duration time.Duration
}
