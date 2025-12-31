package dsl

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/tidwall/gjson"
)

// ToDo: ExpectResponseToMatchJSONSchema("path/to/schema.json")
// ToDo: WithRetries(count: 3, delay: 2*time.Second)
// ToDo: MultipartBody(form *client.MultipartForm)
// ToDo: parseErrorResponse жестко зашита на определенную структуру, сделать гибкой

func getJSONResult(raw []byte, path string) (gjson.Result, error) {
	if !gjson.ValidBytes(raw) {
		return gjson.Result{}, fmt.Errorf("invalid JSON")
	}
	return gjson.GetBytes(raw, path), nil
}

func gjsonTypeToString(t gjson.Type) string {
	switch t {
	case gjson.Null:
		return "null"
	case gjson.False, gjson.True:
		return "boolean"
	case gjson.Number:
		return "number"
	case gjson.String:
		return "string"
	case gjson.JSON:
		return "object/array"
	default:
		return "unknown"
	}
}

func debugValue(res gjson.Result) string {
	if res.Raw != "" {
		return res.Raw
	}
	return fmt.Sprintf("%v", res.Value())
}

func toInt64(v any) (int64, bool) {
	switch val := v.(type) {
	case int:
		return int64(val), true
	case int8:
		return int64(val), true
	case int16:
		return int64(val), true
	case int32:
		return int64(val), true
	case int64:
		return val, true
	default:
		return 0, false
	}
}

func toUint64(v any) (uint64, bool) {
	switch val := v.(type) {
	case uint:
		return uint64(val), true
	case uint8:
		return uint64(val), true
	case uint16:
		return uint64(val), true
	case uint32:
		return uint64(val), true
	case uint64:
		return val, true
	default:
		return 0, false
	}
}

func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}

func isEmptyJSONResult(res gjson.Result) bool {
	if !res.Exists() {
		return true
	}
	if res.Type == gjson.Null {
		return true
	}

	switch res.Type {
	case gjson.String:
		return strings.TrimSpace(res.String()) == ""
	case gjson.JSON:
		if res.IsArray() {
			return len(res.Array()) == 0
		}
		if res.IsObject() {
			return len(res.Map()) == 0
		}
		return false
	default:
		return false
	}
}

func compareJSONResult(res gjson.Result, expected any) (bool, string) {
	if expected == nil {
		if !res.Exists() {
			return false, "field does not exist (expected null)"
		}
		if res.Type != gjson.Null {
			return false, fmt.Sprintf("expected null, got %s: %s", gjsonTypeToString(res.Type), debugValue(res))
		}
		return true, ""
	}

	switch exp := expected.(type) {
	case string:
		if res.Type != gjson.String {
			return false, fmt.Sprintf("expected string %q, got %s: %s", exp, gjsonTypeToString(res.Type), debugValue(res))
		}
		actual := res.String()
		if actual != exp {
			return false, fmt.Sprintf("expected %q, got %q", exp, actual)
		}
		return true, ""

	case bool:
		if res.Type != gjson.True && res.Type != gjson.False {
			return false, fmt.Sprintf("expected boolean %v, got %s: %s", exp, gjsonTypeToString(res.Type), debugValue(res))
		}
		actual := res.Bool()
		if actual != exp {
			return false, fmt.Sprintf("expected %v, got %v", exp, actual)
		}
		return true, ""

	default:
		if expectedInt, ok := toInt64(expected); ok {
			if res.Type != gjson.Number {
				return false, fmt.Sprintf("expected number %d, got %s: %s", expectedInt, gjsonTypeToString(res.Type), debugValue(res))
			}
			actualFloat := res.Float()
			if math.Trunc(actualFloat) != actualFloat {
				return false, fmt.Sprintf("expected integer %d, got float %v", expectedInt, actualFloat)
			}
			actualInt := res.Int()
			if actualInt != expectedInt {
				return false, fmt.Sprintf("expected %d, got %d", expectedInt, actualInt)
			}
			return true, ""
		}

		if expectedUint, ok := toUint64(expected); ok {
			if res.Type != gjson.Number {
				return false, fmt.Sprintf("expected number %d, got %s: %s", expectedUint, gjsonTypeToString(res.Type), debugValue(res))
			}
			actualFloat := res.Float()
			if math.Trunc(actualFloat) != actualFloat {
				return false, fmt.Sprintf("expected integer %d, got float %v", expectedUint, actualFloat)
			}
			actualUint := res.Uint()
			if actualUint != expectedUint {
				return false, fmt.Sprintf("expected %d, got %d", expectedUint, actualUint)
			}
			return true, ""
		}

		if expectedFloat, ok := toFloat64(expected); ok {
			if res.Type != gjson.Number {
				return false, fmt.Sprintf("expected number %v, got %s: %s", expectedFloat, gjsonTypeToString(res.Type), debugValue(res))
			}
			actualFloat := res.Float()
			if actualFloat != expectedFloat {
				return false, fmt.Sprintf("expected %v, got %v", expectedFloat, actualFloat)
			}
			return true, ""
		}

		return false, fmt.Sprintf("unsupported expected type %T; supported: string/bool/int*/uint*/float*/nil", expected)
	}
}

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
				a.True(false, "Expected non-empty response body")
				return
			}

			res, err := getJSONResult(c.resp.RawBody, path)
			if err != nil {
				a.NoError(err, "Expected valid JSON response body")
				return
			}

			if !res.Exists() {
				a.True(false, fmt.Sprintf("Expected JSON field '%s' to be present", path))
				return
			}

			a.True(!isEmptyJSONResult(res), "Expected JSON field '%s' to be non-empty", path)
		})
	})

	return c
}

func (c *Call[TReq, TResp]) ExpectResponseBodyFieldValue(path string, expected any) *Call[TReq, TResp] {
	title := fmt.Sprintf("Expect JSON field '%s' == %v", path, expected)

	c.addExpectation(func(parent provider.StepCtx) {
		parent.WithNewStep(title, func(stepCtx provider.StepCtx) {
			a := c.pickAsserter(stepCtx)

			if !c.ensureResponseSilent(a) {
				return
			}

			if len(c.resp.RawBody) == 0 {
				a.True(false, "Expected non-empty response body")
				return
			}

			res, err := getJSONResult(c.resp.RawBody, path)
			if err != nil {
				a.NoError(err, "Expected valid JSON response body")
				return
			}

			if !res.Exists() {
				a.True(false, fmt.Sprintf("Path '%s' does not exist in response", path))
				return
			}

			ok, msg := compareJSONResult(res, expected)
			if !ok {
				a.True(false, fmt.Sprintf("Field '%s': %s", path, msg))
				return
			}
		})
	})

	return c
}
