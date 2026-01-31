package dsl

import (
	"fmt"
	"strings"
	"time"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/jsonutil"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/redis/client"
)

var preCheck = client.BuildPreCheck()
var preCheckKeyExists = client.BuildKeyExistsPreCheck(preCheck)

var jsonSource = &expect.JSONExpectationSource[*client.Result]{
	GetJSON: func(result *client.Result) ([]byte, error) {
		if err := jsonutil.ValidateString(result.Value); err != nil {
			return nil, fmt.Errorf("value is not valid JSON")
		}
		return []byte(result.Value), nil
	},
	PreCheck:         preCheck,
	PreCheckWithBody: preCheckKeyExists,
}

func (q *Query) ExpectExists() *Query {
	q.addExpectation(makeExistsExpectation())
	return q
}

func (q *Query) ExpectNotExists() *Query {
	q.addExpectation(makeNotExistsExpectation())
	return q
}

func (q *Query) ExpectValueEquals(expected string) *Query {
	q.addExpectation(makeValueExpectation(expected))
	return q
}

func (q *Query) ExpectValueNotEmpty() *Query {
	q.addExpectation(makeValueNotEmptyExpectation())
	return q
}

func (q *Query) ExpectFieldEquals(path string, expected any) *Query {
	q.addExpectation(jsonSource.FieldEquals(path, expected))
	return q
}

func (q *Query) ExpectFieldNotEmpty(path string) *Query {
	q.addExpectation(jsonSource.FieldNotEmpty(path))
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
