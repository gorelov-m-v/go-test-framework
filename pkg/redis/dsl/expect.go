package dsl

import (
	"fmt"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/redis/client"
)

func (q *Query) ExpectExists() *Query {
	q.addExpectation(makeExistsExpectation())
	return q
}

func (q *Query) ExpectNotExists() *Query {
	q.addExpectation(makeNotExistsExpectation())
	return q
}

func (q *Query) ExpectValue(expected string) *Query {
	q.addExpectation(makeValueExpectation(expected))
	return q
}

func (q *Query) ExpectValueNotEmpty() *Query {
	q.addExpectation(makeValueNotEmptyExpectation())
	return q
}

func (q *Query) ExpectJSONField(path string, expected any) *Query {
	q.addExpectation(makeJSONFieldExpectation(path, expected))
	return q
}

func (q *Query) ExpectJSONFieldNotEmpty(path string) *Query {
	q.addExpectation(makeJSONFieldNotEmptyExpectation(path))
	return q
}

func (q *Query) ExpectTTL(minTTL, maxTTL time.Duration) *Query {
	q.addExpectation(makeTTLExpectation(minTTL, maxTTL))
	return q
}

func (q *Query) ExpectNoTTL() *Query {
	q.addExpectation(makeNoTTLExpectation())
	return q
}

func preCheck(err error, result *client.Result) (polling.CheckResult, bool) {
	if err != nil {
		return polling.CheckResult{
			Ok:        false,
			Retryable: true,
			Reason:    fmt.Sprintf("Query failed: %v", err),
		}, false
	}
	if result == nil {
		return polling.CheckResult{
			Ok:        false,
			Retryable: true,
			Reason:    "Result is nil",
		}, false
	}
	if result.Error != nil {
		return polling.CheckResult{
			Ok:        false,
			Retryable: true,
			Reason:    fmt.Sprintf("Redis error: %v", result.Error),
		}, false
	}
	return polling.CheckResult{}, true
}

func preCheckKeyExists(err error, result *client.Result) (polling.CheckResult, bool) {
	if res, ok := preCheck(err, result); !ok {
		return res, false
	}
	if !result.Exists {
		return polling.CheckResult{
			Ok:        false,
			Retryable: true,
			Reason:    fmt.Sprintf("Key '%s' does not exist", result.Key),
		}, false
	}
	return polling.CheckResult{}, true
}

func makeExistsExpectation() *expect.Expectation[*client.Result] {
	name := "Expect: Key exists"
	return expect.New(
		name,
		func(err error, result *client.Result) polling.CheckResult {
			if res, ok := preCheckKeyExists(err, result); !ok {
				return res
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[*client.Result](name),
	)
}

func makeNotExistsExpectation() *expect.Expectation[*client.Result] {
	name := "Expect: Key not exists"
	return expect.New(
		name,
		func(err error, result *client.Result) polling.CheckResult {
			if res, ok := preCheck(err, result); !ok {
				return res
			}
			if result.Exists {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Key '%s' exists but expected not to", result.Key),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[*client.Result](name),
	)
}

func makeValueExpectation(expected string) *expect.Expectation[*client.Result] {
	name := fmt.Sprintf("Expect: Value = %q", expected)
	return expect.New(
		name,
		func(err error, result *client.Result) polling.CheckResult {
			if res, ok := preCheckKeyExists(err, result); !ok {
				return res
			}
			if result.Value != expected {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected value %q, got %q", expected, result.Value),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[*client.Result](name),
	)
}

func makeValueNotEmptyExpectation() *expect.Expectation[*client.Result] {
	name := "Expect: Value not empty"
	return expect.New(
		name,
		func(err error, result *client.Result) polling.CheckResult {
			if res, ok := preCheckKeyExists(err, result); !ok {
				return res
			}
			if strings.TrimSpace(result.Value) == "" {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Value is empty",
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[*client.Result](name),
	)
}

func makeJSONFieldExpectation(path string, expected any) *expect.Expectation[*client.Result] {
	name := fmt.Sprintf("Expect: JSON field '%s' = %v", path, expected)
	return expect.BuildJSONFieldExpectation(expect.JSONFieldExpectationConfig[*client.Result]{
		Path:       path,
		ExpectName: name,
		GetJSON: func(result *client.Result) ([]byte, error) {
			if !gjson.Valid(result.Value) {
				return nil, fmt.Errorf("value is not valid JSON")
			}
			return []byte(result.Value), nil
		},
		PreCheck: preCheckKeyExists,
		Check:    expect.JSONCheckEquals(expected),
	})
}

func makeJSONFieldNotEmptyExpectation(path string) *expect.Expectation[*client.Result] {
	name := fmt.Sprintf("Expect: JSON field '%s' not empty", path)
	return expect.BuildJSONFieldExpectation(expect.JSONFieldExpectationConfig[*client.Result]{
		Path:       path,
		ExpectName: name,
		GetJSON: func(result *client.Result) ([]byte, error) {
			if !gjson.Valid(result.Value) {
				return nil, fmt.Errorf("value is not valid JSON")
			}
			return []byte(result.Value), nil
		},
		PreCheck: preCheckKeyExists,
		Check:    expect.JSONCheckNotEmpty(),
	})
}

func makeTTLExpectation(minTTL, maxTTL time.Duration) *expect.Expectation[*client.Result] {
	name := fmt.Sprintf("Expect: TTL between %v and %v", minTTL, maxTTL)
	return expect.New(
		name,
		func(err error, result *client.Result) polling.CheckResult {
			if res, ok := preCheckKeyExists(err, result); !ok {
				return res
			}
			if result.TTL < minTTL || result.TTL > maxTTL {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("TTL is %v, expected between %v and %v", result.TTL, minTTL, maxTTL),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[*client.Result](name),
	)
}

func makeNoTTLExpectation() *expect.Expectation[*client.Result] {
	name := "Expect: No TTL (persistent)"
	return expect.New(
		name,
		func(err error, result *client.Result) polling.CheckResult {
			if res, ok := preCheckKeyExists(err, result); !ok {
				return res
			}
			if result.TTL != -1 {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Key has TTL of %v, expected no TTL", result.TTL),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[*client.Result](name),
	)
}
