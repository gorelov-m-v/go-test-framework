package httpdsl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ozontech/allure-go/pkg/framework/provider"
)

// ToDo: ExpectResponseBodyFieldValue("user.id", 123)
// ToDo: ExpectResponseToMatchJSONSchema("path/to/schema.json")
// ToDo: jsonLookup поддерживает только вложенность через точку (a.b.c). Поддержать массивы a[i].b
// ToDo: WithRetries(count: 3, delay: 2*time.Second)
// ToDo: MultipartBody(form *httpclient.MultipartForm)
// ToDo: parseErrorResponse жестко зашита на определенную структуру, сделать гибкой
func (c *Call[TReq, TResp]) ensureResponseSilent(a provider.Asserts) bool {
	if c.resp == nil {
		a.NotNil(c.resp, "Expected HTTP response to be available (got nil)")
		return false
	}
	if c.resp.NetworkError != "" {
		a.Equal("", c.resp.NetworkError, "Expected no network error")
		return false
	}
	return true
}

func (c *Call[TReq, TResp]) ExpectResponseStatus(code int) *Call[TReq, TResp] {
	title := fmt.Sprintf("Expect response status %d %s", code, http.StatusText(code))

	c.addExpectation(func(parent provider.StepCtx) {
		parent.WithNewStep(title, func(stepCtx provider.StepCtx) {
			a := c.pickAsserter(stepCtx)

			if !c.ensureResponseSilent(a) {
				return
			}

			a.Equal(code, c.resp.StatusCode, "Expected response status %d %s", code, http.StatusText(code))
		})
	})

	return c
}

func (c *Call[TReq, TResp]) ExpectResponseBodyNotEmpty() *Call[TReq, TResp] {
	c.addExpectation(func(parent provider.StepCtx) {
		parent.WithNewStep("Expect response body not empty", func(stepCtx provider.StepCtx) {
			a := c.pickAsserter(stepCtx)

			if !c.ensureResponseSilent(a) {
				return
			}

			a.True(len(c.resp.RawBody) > 0, "Expected non-empty response body")
		})
	})

	return c
}

func (c *Call[TReq, TResp]) ExpectResponseBodyFieldNotEmpty(path string) *Call[TReq, TResp] {
	title := fmt.Sprintf("Expect JSON field not empty: %s", path)

	c.addExpectation(func(parent provider.StepCtx) {
		parent.WithNewStep(title, func(stepCtx provider.StepCtx) {
			a := c.pickAsserter(stepCtx)

			if !c.ensureResponseSilent(a) {
				return
			}

			if len(c.resp.RawBody) == 0 {
				a.True(false, "Expected non-empty response body") // Explicit fail
				return
			}

			var root any
			err := json.Unmarshal(c.resp.RawBody, &root)
			if err != nil {
				a.NoError(err, "Expected valid JSON response body") // Explicit fail
				return
			}

			val, ok := jsonLookup(root, path)
			if !ok {
				a.True(false, fmt.Sprintf("Expected JSON field '%s' to be present", path)) // Explicit fail
				return
			}

			a.True(isNonEmptyJSONValue(val), "Expected JSON field '%s' to be non-empty", path)
		})
	})

	return c
}

func jsonLookup(root any, path string) (any, bool) {
	cur := root
	for _, key := range strings.Split(path, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		v, ok := m[key]
		if !ok {
			return nil, false
		}
		cur = v
	}
	return cur, true
}

func isNonEmptyJSONValue(v any) bool {
	if v == nil {
		return false
	}

	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x) != ""
	case []any:
		return len(x) > 0
	case map[string]any:
		return len(x) > 0
	default:
		return true
	}
}
