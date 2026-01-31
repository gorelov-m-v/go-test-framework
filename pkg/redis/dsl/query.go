package dsl

import (
	"context"
	"fmt"
	"strings"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/validation"
	"github.com/gorelov-m-v/go-test-framework/pkg/redis/client"
)

// Query represents a Redis key query builder with fluent interface.
// It supports key existence checks, value expectations, JSON field access,
// TTL validation, and automatic retry in async mode.
//
// Example:
//
//	dsl.NewQuery(sCtx, redisClient).
//	    Key("user:123").
//	    ExpectExists().
//	    ExpectFieldEquals("status", "active").
//	    Send()
type Query struct {
	sCtx   provider.StepCtx
	client *client.Client
	ctx    context.Context

	customStepName string
	key            string

	result *client.Result
	sent   bool

	expectations []*expect.Expectation[*client.Result]
}

// NewQuery creates a new Redis query builder.
//
// Parameters:
//   - sCtx: Allure step context for test reporting
//   - redisClient: Redis client configured with connection settings
//
// Returns a Query builder that can be configured with key and expectations.
func NewQuery(sCtx provider.StepCtx, redisClient *client.Client) *Query {
	return &Query{
		sCtx:   sCtx,
		client: redisClient,
		ctx:    context.Background(),
	}
}

// StepName overrides the default step name in Allure report.
func (q *Query) StepName(name string) *Query {
	q.customStepName = strings.TrimSpace(name)
	return q
}

// Context sets a custom context for the query operation.
func (q *Query) Context(ctx context.Context) *Query {
	if ctx != nil {
		q.ctx = ctx
	}
	return q
}

// Key sets the Redis key to query.
func (q *Query) Key(key string) *Query {
	q.key = key
	return q
}

func (q *Query) addExpectation(exp *expect.Expectation[*client.Result]) {
	expect.AddExpectation(q.sCtx, q.sent, &q.expectations, exp, "Redis")
}

// Send executes the Redis query and validates all expectations.
// In async mode (AsyncStep), automatically retries with backoff until expectations pass.
// Returns the Result containing key existence, value, and TTL information.
func (q *Query) Send() *client.Result {
	q.validate()

	q.sCtx.WithNewStep(q.stepName(), func(stepCtx provider.StepCtx) {
		result, err, summary := q.execute(stepCtx, q.expectations)
		q.result = result
		q.sent = true

		attachRedisReport(stepCtx, q, q.result, summary)
		q.assertResults(stepCtx, err)
	})

	return q.result
}

func (q *Query) stepName() string {
	if q.customStepName != "" {
		return q.customStepName
	}
	return fmt.Sprintf("Redis GET %s", q.key)
}

func (q *Query) assertResults(stepCtx provider.StepCtx, err error) {
	expect.AssertExpectations(stepCtx, q.expectations, err, q.result, q.assertNoExpectations)
}

func (q *Query) assertNoExpectations(stepCtx provider.StepCtx, mode polling.AssertionMode, err error) {
	if err != nil {
		polling.NoError(stepCtx, mode, err, "Redis query failed: %v", err)
	}
}

func (q *Query) validate() {
	v := validation.New(q.sCtx, "Redis")
	v.RequireNotNil(q.client, "Redis client")
	v.RequireNotEmptyWithHint(q.key, "Redis key", "Use .Key(\"key_name\").")
}
