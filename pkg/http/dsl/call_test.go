package dsl

import (
	"net/http"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorelov-m-v/go-test-framework/pkg/http/client"
)

var _ provider.Asserts = (*mockAsserts)(nil)

type mockStepCtx struct {
	brokenCalled    bool
	brokenNowCalled bool
	breakMessage    string
	steps           []string
}

func (m *mockStepCtx) Step(step *allure.Step)                                            {}
func (m *mockStepCtx) NewStep(stepName string, parameters ...*allure.Parameter)          {}
func (m *mockStepCtx) WithNewStep(stepName string, step func(sCtx provider.StepCtx), params ...*allure.Parameter) {
	m.steps = append(m.steps, stepName)
	step(m)
}
func (m *mockStepCtx) WithNewAsyncStep(stepName string, step func(sCtx provider.StepCtx), params ...*allure.Parameter) {
}
func (m *mockStepCtx) WithParameters(parameters ...*allure.Parameter)      {}
func (m *mockStepCtx) WithNewParameters(kv ...interface{})                 {}
func (m *mockStepCtx) WithAttachments(attachment ...*allure.Attachment)    {}
func (m *mockStepCtx) WithNewAttachment(name string, mimeType allure.MimeType, content []byte) {
}
func (m *mockStepCtx) Assert() provider.Asserts  { return &mockAsserts{} }
func (m *mockStepCtx) Require() provider.Asserts { return &mockAsserts{} }
func (m *mockStepCtx) LogStep(args ...interface{})                         {}
func (m *mockStepCtx) LogfStep(format string, args ...interface{})         {}
func (m *mockStepCtx) WithStatusDetails(message, trace string)             {}
func (m *mockStepCtx) CurrentStep() *allure.Step                           { return nil }
func (m *mockStepCtx) Broken()                                             { m.brokenCalled = true }
func (m *mockStepCtx) BrokenNow()                                          { m.brokenNowCalled = true }
func (m *mockStepCtx) Fail()                                               {}
func (m *mockStepCtx) FailNow()                                            {}
func (m *mockStepCtx) Log(args ...interface{})                             {}
func (m *mockStepCtx) Logf(format string, args ...interface{})             {}
func (m *mockStepCtx) Error(args ...interface{})                           {}
func (m *mockStepCtx) Errorf(format string, args ...interface{})           {}
func (m *mockStepCtx) Break(args ...interface{}) {
	m.brokenCalled = true
	if len(args) > 0 {
		m.breakMessage = args[0].(string)
	}
}
func (m *mockStepCtx) Breakf(format string, args ...interface{}) { m.brokenCalled = true }
func (m *mockStepCtx) Name() string                              { return "mock" }

type mockAsserts struct{}

func (m *mockAsserts) Exactly(expected, actual interface{}, msgAndArgs ...interface{})          {}
func (m *mockAsserts) Same(expected, actual interface{}, msgAndArgs ...interface{})             {}
func (m *mockAsserts) NotSame(expected, actual interface{}, msgAndArgs ...interface{})          {}
func (m *mockAsserts) Equal(expected, actual interface{}, msgAndArgs ...interface{})            {}
func (m *mockAsserts) NotEqual(expected, actual interface{}, msgAndArgs ...interface{})         {}
func (m *mockAsserts) EqualValues(expected, actual interface{}, msgAndArgs ...interface{})      {}
func (m *mockAsserts) NotEqualValues(expected, actual interface{}, msgAndArgs ...interface{})   {}
func (m *mockAsserts) Error(err error, msgAndArgs ...interface{})                               {}
func (m *mockAsserts) NoError(err error, msgAndArgs ...interface{})                             {}
func (m *mockAsserts) EqualError(theError error, errString string, msgAndArgs ...interface{})   {}
func (m *mockAsserts) ErrorIs(err, target error, msgAndArgs ...interface{})                     {}
func (m *mockAsserts) ErrorAs(err error, target interface{}, msgAndArgs ...interface{})         {}
func (m *mockAsserts) NotNil(object interface{}, msgAndArgs ...interface{})                     {}
func (m *mockAsserts) Nil(object interface{}, msgAndArgs ...interface{})                        {}
func (m *mockAsserts) Len(object interface{}, length int, msgAndArgs ...interface{})            {}
func (m *mockAsserts) NotContains(s, contains interface{}, msgAndArgs ...interface{})           {}
func (m *mockAsserts) Contains(s, contains interface{}, msgAndArgs ...interface{})              {}
func (m *mockAsserts) Greater(e1, e2 interface{}, msgAndArgs ...interface{})                    {}
func (m *mockAsserts) GreaterOrEqual(e1, e2 interface{}, msgAndArgs ...interface{})             {}
func (m *mockAsserts) Less(e1, e2 interface{}, msgAndArgs ...interface{})                       {}
func (m *mockAsserts) LessOrEqual(e1, e2 interface{}, msgAndArgs ...interface{})                {}
func (m *mockAsserts) Implements(interfaceObject, object interface{}, msgAndArgs ...interface{}) {
}
func (m *mockAsserts) Empty(object interface{}, msgAndArgs ...interface{})    {}
func (m *mockAsserts) NotEmpty(object interface{}, msgAndArgs ...interface{}) {}
func (m *mockAsserts) WithinDuration(expected, actual time.Time, delta time.Duration, msgAndArgs ...interface{}) {
}
func (m *mockAsserts) JSONEq(expected, actual string, msgAndArgs ...interface{})       {}
func (m *mockAsserts) JSONContains(expected, actual string, msgAndArgs ...interface{}) {}
func (m *mockAsserts) Subset(list, subset interface{}, msgAndArgs ...interface{})      {}
func (m *mockAsserts) NotSubset(list, subset interface{}, msgAndArgs ...interface{})   {}
func (m *mockAsserts) IsType(expectedType, object interface{}, msgAndArgs ...interface{}) {
}
func (m *mockAsserts) True(value bool, msgAndArgs ...interface{})                      {}
func (m *mockAsserts) False(value bool, msgAndArgs ...interface{})                     {}
func (m *mockAsserts) Regexp(rx, str interface{}, msgAndArgs ...interface{})           {}
func (m *mockAsserts) ElementsMatch(listA, listB interface{}, msgAndArgs ...interface{}) {
}
func (m *mockAsserts) DirExists(path string, msgAndArgs ...interface{}) {}
func (m *mockAsserts) Condition(condition assert.Comparison, msgAndArgs ...interface{}) {
}
func (m *mockAsserts) Zero(i interface{}, msgAndArgs ...interface{}) {}
func (m *mockAsserts) NotZero(i interface{}, msgAndArgs ...interface{}) {}
func (m *mockAsserts) InDelta(expected, actual interface{}, delta float64, msgAndArgs ...interface{}) {
}
func (m *mockAsserts) Eventually(condition func() bool, waitFor, tick time.Duration, msgAndArgs ...interface{}) {
}

func newTestClient() *client.Client {
	c, _ := client.New(client.Config{BaseURL: "https://api.example.com"})
	return c
}

func TestNewCall(t *testing.T) {
	mockCtx := &mockStepCtx{}
	httpClient := newTestClient()

	call := NewCall[any, any](mockCtx, httpClient)

	require.NotNil(t, call)
	assert.Equal(t, mockCtx, call.sCtx)
	assert.Equal(t, httpClient, call.client)
	assert.NotNil(t, call.ctx)
	assert.NotNil(t, call.req)
	assert.NotNil(t, call.req.Headers)
	assert.NotNil(t, call.req.PathParams)
	assert.NotNil(t, call.req.QueryParams)
	assert.False(t, call.sent)
	assert.Empty(t, call.expectations)
}

func TestCallBuilderMethods(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*Call[any, any])
		wantMethod string
		wantPath   string
	}{
		{
			name:       "GET",
			setup:      func(c *Call[any, any]) { c.GET("/users") },
			wantMethod: http.MethodGet,
			wantPath:   "/users",
		},
		{
			name:       "POST",
			setup:      func(c *Call[any, any]) { c.POST("/users") },
			wantMethod: http.MethodPost,
			wantPath:   "/users",
		},
		{
			name:       "PUT",
			setup:      func(c *Call[any, any]) { c.PUT("/users/1") },
			wantMethod: http.MethodPut,
			wantPath:   "/users/1",
		},
		{
			name:       "PATCH",
			setup:      func(c *Call[any, any]) { c.PATCH("/users/1") },
			wantMethod: http.MethodPatch,
			wantPath:   "/users/1",
		},
		{
			name:       "DELETE",
			setup:      func(c *Call[any, any]) { c.DELETE("/users/1") },
			wantMethod: http.MethodDelete,
			wantPath:   "/users/1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := &mockStepCtx{}
			call := NewCall[any, any](mockCtx, newTestClient())

			tt.setup(call)

			assert.Equal(t, tt.wantMethod, call.req.Method)
			assert.Equal(t, tt.wantPath, call.req.Path)
		})
	}
}

func TestCallHeader(t *testing.T) {
	mockCtx := &mockStepCtx{}
	call := NewCall[any, any](mockCtx, newTestClient())

	result := call.Header("Authorization", "Bearer token123")

	assert.Same(t, call, result)
	assert.Equal(t, "Bearer token123", call.req.Headers["Authorization"])
}

func TestCallHeaderMultiple(t *testing.T) {
	mockCtx := &mockStepCtx{}
	call := NewCall[any, any](mockCtx, newTestClient())

	call.Header("Authorization", "Bearer token").
		Header("Content-Type", "application/json").
		Header("X-Custom", "value")

	assert.Equal(t, "Bearer token", call.req.Headers["Authorization"])
	assert.Equal(t, "application/json", call.req.Headers["Content-Type"])
	assert.Equal(t, "value", call.req.Headers["X-Custom"])
}

func TestCallPathParam(t *testing.T) {
	mockCtx := &mockStepCtx{}
	call := NewCall[any, any](mockCtx, newTestClient())

	result := call.PathParam("id", "123")

	assert.Same(t, call, result)
	assert.Equal(t, "123", call.req.PathParams["id"])
}

func TestCallPathParamMultiple(t *testing.T) {
	mockCtx := &mockStepCtx{}
	call := NewCall[any, any](mockCtx, newTestClient())

	call.PathParam("userId", "1").PathParam("postId", "42")

	assert.Equal(t, "1", call.req.PathParams["userId"])
	assert.Equal(t, "42", call.req.PathParams["postId"])
}

func TestCallQueryParam(t *testing.T) {
	mockCtx := &mockStepCtx{}
	call := NewCall[any, any](mockCtx, newTestClient())

	result := call.QueryParam("page", "1")

	assert.Same(t, call, result)
	assert.Equal(t, "1", call.req.QueryParams["page"])
}

func TestCallQueryParamMultiple(t *testing.T) {
	mockCtx := &mockStepCtx{}
	call := NewCall[any, any](mockCtx, newTestClient())

	call.QueryParam("page", "1").QueryParam("limit", "10").QueryParam("sort", "desc")

	assert.Equal(t, "1", call.req.QueryParams["page"])
	assert.Equal(t, "10", call.req.QueryParams["limit"])
	assert.Equal(t, "desc", call.req.QueryParams["sort"])
}

func TestCallRequestBody(t *testing.T) {
	type CreateUserRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	mockCtx := &mockStepCtx{}
	call := NewCall[CreateUserRequest, any](mockCtx, newTestClient())

	body := CreateUserRequest{Name: "John", Email: "john@example.com"}
	result := call.RequestBody(body)

	assert.Same(t, call, result)
	require.NotNil(t, call.req.Body)
	assert.Equal(t, "John", call.req.Body.Name)
	assert.Equal(t, "john@example.com", call.req.Body.Email)
}

func TestCallRequestBodyMap(t *testing.T) {
	mockCtx := &mockStepCtx{}
	call := NewCall[any, any](mockCtx, newTestClient())

	body := map[string]interface{}{
		"name":  "John",
		"email": "john@example.com",
	}
	result := call.RequestBodyMap(body)

	assert.Same(t, call, result)
	assert.Equal(t, body, call.req.BodyMap)
}

func TestCallChaining(t *testing.T) {
	type Request struct {
		Name string `json:"name"`
	}

	mockCtx := &mockStepCtx{}
	call := NewCall[Request, any](mockCtx, newTestClient())

	call.POST("/users/{id}").
		PathParam("id", "123").
		QueryParam("notify", "true").
		Header("Authorization", "Bearer token").
		RequestBody(Request{Name: "John"})

	assert.Equal(t, http.MethodPost, call.req.Method)
	assert.Equal(t, "/users/{id}", call.req.Path)
	assert.Equal(t, "123", call.req.PathParams["id"])
	assert.Equal(t, "true", call.req.QueryParams["notify"])
	assert.Equal(t, "Bearer token", call.req.Headers["Authorization"])
	require.NotNil(t, call.req.Body)
	assert.Equal(t, "John", call.req.Body.Name)
}

func TestCallValidate(t *testing.T) {
	tests := []struct {
		name         string
		setupCall    func(*Call[any, any])
		wantBroken   bool
		wantContains string
	}{
		{
			name: "valid call",
			setupCall: func(c *Call[any, any]) {
				c.GET("/users")
			},
			wantBroken: false,
		},
		{
			name: "nil client",
			setupCall: func(c *Call[any, any]) {
				c.client = nil
				c.GET("/users")
			},
			wantBroken:   true,
			wantContains: "client is nil",
		},
		{
			name: "nil request",
			setupCall: func(c *Call[any, any]) {
				c.req = nil
			},
			wantBroken:   true,
			wantContains: "request is nil",
		},
		{
			name: "empty method",
			setupCall: func(c *Call[any, any]) {
				c.req.Path = "/users"
			},
			wantBroken:   true,
			wantContains: "method is not set",
		},
		{
			name: "whitespace method",
			setupCall: func(c *Call[any, any]) {
				c.req.Method = "   "
				c.req.Path = "/users"
			},
			wantBroken:   true,
			wantContains: "method is not set",
		},
		{
			name: "empty path",
			setupCall: func(c *Call[any, any]) {
				c.req.Method = "GET"
			},
			wantBroken:   true,
			wantContains: "path is not set",
		},
		{
			name: "whitespace path",
			setupCall: func(c *Call[any, any]) {
				c.req.Method = "GET"
				c.req.Path = "   "
			},
			wantBroken:   true,
			wantContains: "path is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := &mockStepCtx{}
			call := NewCall[any, any](mockCtx, newTestClient())

			tt.setupCall(call)
			call.validate()

			assert.Equal(t, tt.wantBroken, mockCtx.brokenCalled)
			if tt.wantContains != "" {
				assert.Contains(t, mockCtx.breakMessage, tt.wantContains)
			}
		})
	}
}

func TestCallValidateContractConfig(t *testing.T) {
	tests := []struct {
		name             string
		validateContract bool
		contractSchema   string
		wantBroken       bool
	}{
		{
			name:             "no contract validation",
			validateContract: false,
			contractSchema:   "",
			wantBroken:       false,
		},
		{
			name:             "contract validation without validator",
			validateContract: true,
			contractSchema:   "",
			wantBroken:       true,
		},
		{
			name:             "schema validation without validator",
			validateContract: false,
			contractSchema:   "UserResponse",
			wantBroken:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := &mockStepCtx{}
			call := NewCall[any, any](mockCtx, newTestClient())

			call.validateContract = tt.validateContract
			call.contractSchema = tt.contractSchema

			call.validateContractConfig()

			assert.Equal(t, tt.wantBroken, mockCtx.brokenCalled)
		})
	}
}

func TestCallAddExpectation(t *testing.T) {
	mockCtx := &mockStepCtx{}
	call := NewCall[any, any](mockCtx, newTestClient())

	call.ExpectResponseStatus(200)

	assert.Len(t, call.expectations, 1)
	assert.False(t, mockCtx.brokenCalled)
}

func TestCallAddExpectationAfterSend(t *testing.T) {
	mockCtx := &mockStepCtx{}
	call := NewCall[any, any](mockCtx, newTestClient())
	call.sent = true

	call.addExpectation(nil)

	assert.True(t, mockCtx.brokenCalled)
	assert.Contains(t, mockCtx.breakMessage, "before Send()")
}

func TestExpectMethodsAddExpectations(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*Call[any, any])
	}{
		{
			name:  "ExpectResponseStatus",
			setup: func(c *Call[any, any]) { c.ExpectResponseStatus(200) },
		},
		{
			name:  "ExpectBodyNotEmpty",
			setup: func(c *Call[any, any]) { c.ExpectBodyNotEmpty() },
		},
		{
			name:  "ExpectFieldNotEmpty",
			setup: func(c *Call[any, any]) { c.ExpectFieldNotEmpty("id") },
		},
		{
			name:  "ExpectFieldEquals",
			setup: func(c *Call[any, any]) { c.ExpectFieldEquals("status", "active") },
		},
		{
			name:  "ExpectFieldIsNull",
			setup: func(c *Call[any, any]) { c.ExpectFieldIsNull("deleted_at") },
		},
		{
			name:  "ExpectFieldIsNotNull",
			setup: func(c *Call[any, any]) { c.ExpectFieldIsNotNull("id") },
		},
		{
			name:  "ExpectFieldTrue",
			setup: func(c *Call[any, any]) { c.ExpectFieldTrue("is_active") },
		},
		{
			name:  "ExpectFieldFalse",
			setup: func(c *Call[any, any]) { c.ExpectFieldFalse("is_deleted") },
		},
		{
			name:  "ExpectMatchesContract",
			setup: func(c *Call[any, any]) { c.ExpectMatchesContract() },
		},
		{
			name:  "ExpectMatchesSchema",
			setup: func(c *Call[any, any]) { c.ExpectMatchesSchema("UserResponse") },
		},
		{
			name: "ExpectArrayContains",
			setup: func(c *Call[any, any]) {
				c.ExpectArrayContains("items", map[string]interface{}{"id": 1})
			},
		},
		{
			name: "ExpectArrayContainsExact",
			setup: func(c *Call[any, any]) {
				c.ExpectArrayContainsExact("items", map[string]interface{}{"id": 1})
			},
		},
		{
			name: "ExpectBodyEquals",
			setup: func(c *Call[any, any]) {
				c.ExpectBodyEquals(map[string]interface{}{"id": 1})
			},
		},
		{
			name: "ExpectBodyPartial",
			setup: func(c *Call[any, any]) {
				c.ExpectBodyPartial(map[string]interface{}{"id": 1})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := &mockStepCtx{}
			call := NewCall[any, any](mockCtx, newTestClient())

			initialLen := len(call.expectations)
			tt.setup(call)

			if tt.name == "ExpectMatchesContract" {
				assert.True(t, call.validateContract)
			} else if tt.name == "ExpectMatchesSchema" {
				assert.Equal(t, "UserResponse", call.contractSchema)
			} else {
				assert.Greater(t, len(call.expectations), initialLen)
			}
		})
	}
}

func TestCallMethodsReturnSelf(t *testing.T) {
	mockCtx := &mockStepCtx{}
	call := NewCall[any, any](mockCtx, newTestClient())

	assert.Same(t, call, call.GET("/path"))
	assert.Same(t, call, call.POST("/path"))
	assert.Same(t, call, call.PUT("/path"))
	assert.Same(t, call, call.PATCH("/path"))
	assert.Same(t, call, call.DELETE("/path"))
	assert.Same(t, call, call.Header("key", "value"))
	assert.Same(t, call, call.PathParam("key", "value"))
	assert.Same(t, call, call.QueryParam("key", "value"))
	assert.Same(t, call, call.RequestBodyMap(nil))
	assert.Same(t, call, call.ExpectResponseStatus(200))
	assert.Same(t, call, call.ExpectBodyNotEmpty())
	assert.Same(t, call, call.ExpectFieldNotEmpty("id"))
	assert.Same(t, call, call.ExpectFieldEquals("id", 1))
	assert.Same(t, call, call.ExpectFieldIsNull("id"))
	assert.Same(t, call, call.ExpectFieldIsNotNull("id"))
	assert.Same(t, call, call.ExpectFieldTrue("active"))
	assert.Same(t, call, call.ExpectFieldFalse("deleted"))
	assert.Same(t, call, call.ExpectMatchesContract())
	assert.Same(t, call, call.ExpectMatchesSchema("Schema"))
	assert.Same(t, call, call.ExpectArrayContains("items", nil))
	assert.Same(t, call, call.ExpectArrayContainsExact("items", nil))
	assert.Same(t, call, call.ExpectBodyEquals(nil))
	assert.Same(t, call, call.ExpectBodyPartial(nil))
}
