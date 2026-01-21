package client

import "time"

type Result struct {
	Key      string
	Value    string
	Exists   bool
	TTL      time.Duration
	Error    error
	Duration time.Duration
}
