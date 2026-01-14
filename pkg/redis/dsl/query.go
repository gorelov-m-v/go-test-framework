package dsl

import (
	"context"
	"fmt"
	"strings"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/pkg/extension"
	"github.com/gorelov-m-v/go-test-framework/pkg/redis/client"
)

// Query represents a Redis query DSL builder
type Query struct {
	sCtx   provider.StepCtx
	client *client.Client
	ctx    context.Context

	stepName string
	key      string

	result *client.Result
	sent   bool

	expectations []*expect.Expectation[*client.Result]
}

// NewQuery creates a new Redis query builder
func NewQuery(sCtx provider.StepCtx, redisClient *client.Client) *Query {
	return &Query{
		sCtx:   sCtx,
		client: redisClient,
		ctx:    context.Background(),
	}
}

// StepName sets a custom step name for Allure reporting
func (q *Query) StepName(name string) *Query {
	q.stepName = strings.TrimSpace(name)
	return q
}

// Context sets the context for the Redis operation
func (q *Query) Context(ctx context.Context) *Query {
	if ctx != nil {
		q.ctx = ctx
	}
	return q
}

// Key sets the Redis key to query
func (q *Query) Key(key string) *Query {
	q.key = key
	return q
}

func (q *Query) addExpectation(exp *expect.Expectation[*client.Result]) {
	if q.sent {
		q.sCtx.Break("Redis DSL Error: Expectations must be added before Send(). Call ExpectExists(), ExpectValue(), etc. before Send().")
		q.sCtx.BrokenNow()
		return
	}
	q.expectations = append(q.expectations, exp)
}

// Send executes the Redis query and returns the result
func (q *Query) Send() *client.Result {
	q.validate()

	name := q.stepName
	if name == "" {
		name = fmt.Sprintf("Redis GET %s", q.key)
	}

	q.sCtx.WithNewStep(name, func(stepCtx provider.StepCtx) {
		attachRequest(stepCtx, q)

		mode := extension.GetStepMode(stepCtx)
		useRetry := mode == extension.AsyncMode && len(q.expectations) > 0

		var (
			result  *client.Result
			err     error
			summary extension.PollingSummary
		)

		if useRetry {
			result, err, summary = q.executeWithRetry(stepCtx, q.expectations)
		} else {
			result, err, summary = q.executeSingle()
		}

		if result == nil {
			result = &client.Result{Key: q.key, Error: fmt.Errorf("nil result")}
			if err == nil {
				err = fmt.Errorf("unexpected nil result")
			}
		}

		q.result = result
		q.sent = true

		if mode == extension.AsyncMode {
			extension.AttachPollingSummary(stepCtx, summary)
		}

		attachResult(stepCtx, result)

		assertionMode := extension.GetAssertionModeFromStepMode(mode)

		if len(q.expectations) == 0 {
			if err != nil {
				extension.NoError(stepCtx, assertionMode, err, "Redis query failed: %v", err)
				return
			}
			return
		}

		expect.ReportAll(stepCtx, assertionMode, q.expectations, err, result)
	})

	return q.result
}

func (q *Query) validate() {
	if q.client == nil {
		q.sCtx.Break("Redis DSL Error: Redis client is nil. Check test configuration.")
		q.sCtx.BrokenNow()
		return
	}
	if strings.TrimSpace(q.key) == "" {
		q.sCtx.Break("Redis DSL Error: Redis key is not set. Use .Key(\"key_name\").")
		q.sCtx.BrokenNow()
		return
	}
}
