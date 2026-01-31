package expect

import (
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

type PreCheckConfig[T any] struct {
	IsNil           func(T) bool
	HasError        func(T) error
	GetNetworkError func(T) string
	EmptyBodyCheck  func(T) bool
}

func BuildPreCheck[T any](cfg PreCheckConfig[T]) func(error, T) (polling.CheckResult, bool) {
	return func(err error, result T) (polling.CheckResult, bool) {
		if err != nil {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    "Request failed",
			}, false
		}

		if cfg.IsNil != nil && cfg.IsNil(result) {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    "Response is nil",
			}, false
		}

		if cfg.HasError != nil {
			if resErr := cfg.HasError(result); resErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Response contains error",
				}, false
			}
		}

		if cfg.GetNetworkError != nil {
			if netErr := cfg.GetNetworkError(result); netErr != "" {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Network error occurred",
				}, false
			}
		}

		return polling.CheckResult{}, true
	}
}

func BuildPreCheckWithBody[T any](cfg PreCheckConfig[T]) func(error, T) (polling.CheckResult, bool) {
	basePreCheck := BuildPreCheck(cfg)

	return func(err error, result T) (polling.CheckResult, bool) {
		if res, ok := basePreCheck(err, result); !ok {
			return res, false
		}

		if cfg.EmptyBodyCheck != nil && cfg.EmptyBodyCheck(result) {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    "Response body is empty",
			}, false
		}

		return polling.CheckResult{}, true
	}
}

type SimplePreCheckConfig[T any] struct {
	IsNil    func(T) bool
	HasError func(T) error
}

func BuildSimplePreCheck[T any](cfg SimplePreCheckConfig[T]) func(error, T) (polling.CheckResult, bool) {
	return BuildPreCheck(PreCheckConfig[T]{
		IsNil:    cfg.IsNil,
		HasError: cfg.HasError,
	})
}

type KeyExistsPreCheckConfig[T any] struct {
	BasePreCheck func(error, T) (polling.CheckResult, bool)
	KeyExists    func(T) bool
	GetKey       func(T) string
}

func BuildKeyExistsPreCheck[T any](cfg KeyExistsPreCheckConfig[T]) func(error, T) (polling.CheckResult, bool) {
	return func(err error, result T) (polling.CheckResult, bool) {
		if cfg.BasePreCheck != nil {
			if res, ok := cfg.BasePreCheck(err, result); !ok {
				return res, false
			}
		}

		if cfg.KeyExists != nil && !cfg.KeyExists(result) {
			key := ""
			if cfg.GetKey != nil {
				key = cfg.GetKey(result)
			}
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    formatKeyNotExistsReason(key),
			}, false
		}

		return polling.CheckResult{}, true
	}
}

func formatKeyNotExistsReason(key string) string {
	if key != "" {
		return "Key '" + key + "' does not exist"
	}
	return "Key does not exist"
}
