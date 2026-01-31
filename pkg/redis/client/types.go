package client

import (
	"time"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

type Result struct {
	Key      string
	Value    string
	Exists   bool
	TTL      time.Duration
	Error    error
	Duration time.Duration
}

func (r *Result) GetError() error {
	if r == nil {
		return nil
	}
	return r.Error
}

func ResultPreCheckConfig() expect.PreCheckConfig[*Result] {
	return expect.PreCheckConfig[*Result]{
		IsNil:    func(r *Result) bool { return r == nil },
		HasError: func(r *Result) error { return r.Error },
	}
}

func BuildPreCheck() func(error, *Result) (polling.CheckResult, bool) {
	return expect.BuildPreCheck(ResultPreCheckConfig())
}

func BuildKeyExistsPreCheck(basePreCheck func(error, *Result) (polling.CheckResult, bool)) func(error, *Result) (polling.CheckResult, bool) {
	return expect.BuildKeyExistsPreCheck(expect.KeyExistsPreCheckConfig[*Result]{
		BasePreCheck: basePreCheck,
		KeyExists:    func(r *Result) bool { return r.Exists },
		GetKey:       func(r *Result) string { return r.Key },
	})
}
