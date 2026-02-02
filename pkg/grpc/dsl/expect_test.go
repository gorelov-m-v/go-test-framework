package dsl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
)

func TestPreCheck_Success(t *testing.T) {
	resp := &client.Response[any]{Body: new(any)}
	result, ok := preCheck(nil, resp)

	assert.True(t, ok)
	assert.True(t, result.Ok || result.Reason == "")
}

func TestPreCheck_Error(t *testing.T) {
	result, ok := preCheck(errors.New("test error"), nil)

	assert.False(t, ok)
	assert.False(t, result.Ok)
	assert.True(t, result.Retryable)
}

func TestPreCheck_NilResponse(t *testing.T) {
	result, ok := preCheck(nil, nil)

	assert.False(t, ok)
	assert.False(t, result.Ok)
	assert.Contains(t, result.Reason, "nil")
}

func TestPreCheckWithBody_Success(t *testing.T) {
	body := "test"
	var bodyAny any = body
	resp := &client.Response[any]{Body: &bodyAny}
	_, ok := preCheckWithBody(nil, resp)

	assert.True(t, ok)
}

func TestPreCheckWithBody_NilBody(t *testing.T) {
	resp := &client.Response[any]{Body: nil}
	result, ok := preCheckWithBody(nil, resp)

	assert.False(t, ok)
	assert.Contains(t, result.Reason, "body")
}

func TestPreCheckWithBody_Error(t *testing.T) {
	result, ok := preCheckWithBody(errors.New("test error"), nil)

	assert.False(t, ok)
	assert.True(t, result.Retryable)
}

func TestMakeNoErrorExpectation_Success(t *testing.T) {
	exp := makeNoErrorExpectation()
	resp := &client.Response[any]{}

	result := exp.Check(nil, resp)

	assert.True(t, result.Ok)
}

func TestMakeNoErrorExpectation_WithError(t *testing.T) {
	exp := makeNoErrorExpectation()
	resp := &client.Response[any]{}

	result := exp.Check(errors.New("test error"), resp)

	assert.False(t, result.Ok)
	assert.True(t, result.Retryable)
}

func TestMakeNoErrorExpectation_WithResponseError(t *testing.T) {
	exp := makeNoErrorExpectation()
	resp := &client.Response[any]{Error: errors.New("response error")}

	result := exp.Check(nil, resp)

	assert.False(t, result.Ok)
}

func TestMakeErrorExpectation_Success(t *testing.T) {
	exp := makeErrorExpectation()

	result := exp.Check(errors.New("expected error"), nil)

	assert.True(t, result.Ok)
}

func TestMakeErrorExpectation_NoError(t *testing.T) {
	exp := makeErrorExpectation()
	resp := &client.Response[any]{}

	result := exp.Check(nil, resp)

	assert.False(t, result.Ok)
	assert.True(t, result.Retryable)
}

func TestMakeErrorExpectation_ResponseError(t *testing.T) {
	exp := makeErrorExpectation()
	resp := &client.Response[any]{Error: errors.New("response error")}

	result := exp.Check(nil, resp)

	assert.True(t, result.Ok)
}

func TestMakeStatusCodeExpectation_OK(t *testing.T) {
	exp := makeStatusCodeExpectation(codes.OK)
	resp := &client.Response[any]{}

	result := exp.Check(nil, resp)

	assert.True(t, result.Ok)
}

func TestMakeStatusCodeExpectation_NotFound(t *testing.T) {
	exp := makeStatusCodeExpectation(codes.NotFound)
	err := status.Error(codes.NotFound, "not found")

	result := exp.Check(err, nil)

	assert.True(t, result.Ok)
}

func TestMakeStatusCodeExpectation_Mismatch(t *testing.T) {
	exp := makeStatusCodeExpectation(codes.OK)
	err := status.Error(codes.NotFound, "not found")

	result := exp.Check(err, nil)

	assert.False(t, result.Ok)
	assert.Contains(t, result.Reason, "NotFound")
}

func TestMakeStatusCodeExpectation_InvalidArgument(t *testing.T) {
	exp := makeStatusCodeExpectation(codes.InvalidArgument)
	err := status.Error(codes.InvalidArgument, "invalid")

	result := exp.Check(err, nil)

	assert.True(t, result.Ok)
}

func TestMakeStatusCodeExpectation_PermissionDenied(t *testing.T) {
	exp := makeStatusCodeExpectation(codes.PermissionDenied)
	err := status.Error(codes.PermissionDenied, "denied")

	result := exp.Check(err, nil)

	assert.True(t, result.Ok)
}

func TestMakeStatusCodeExpectation_ResponseError(t *testing.T) {
	exp := makeStatusCodeExpectation(codes.Internal)
	resp := &client.Response[any]{Error: status.Error(codes.Internal, "internal error")}

	result := exp.Check(nil, resp)

	assert.True(t, result.Ok)
}

func TestMakeStatusCodeExpectation_UnknownError(t *testing.T) {
	exp := makeStatusCodeExpectation(codes.Unknown)
	err := errors.New("non-grpc error")

	result := exp.Check(err, nil)

	assert.True(t, result.Ok)
}

func TestGetResponseJSON_Success(t *testing.T) {
	body := map[string]string{"key": "value"}
	var bodyAny any = body
	resp := &client.Response[any]{Body: &bodyAny}

	jsonBytes, err := getResponseJSON(resp)

	assert.NoError(t, err)
	assert.Contains(t, string(jsonBytes), "key")
}

func TestGetResponseJSON_WithRawBody(t *testing.T) {
	body := "test"
	var bodyAny any = body
	resp := &client.Response[any]{
		Body:    &bodyAny,
		RawBody: []byte(`{"raw": "body"}`),
	}

	jsonBytes, err := getResponseJSON(resp)

	assert.NoError(t, err)
	assert.Contains(t, string(jsonBytes), "raw")
}

func TestGetResponseJSON_NilResponse(t *testing.T) {
	_, err := getResponseJSON(nil)

	assert.Error(t, err)
}

func TestGetResponseJSON_NilBody(t *testing.T) {
	resp := &client.Response[any]{Body: nil}

	_, err := getResponseJSON(resp)

	assert.Error(t, err)
}

func TestFieldEquals_Success(t *testing.T) {
	exp := jsonSource.FieldEquals("name", "John")
	body := map[string]string{"name": "John"}
	var bodyAny any = body
	resp := &client.Response[any]{Body: &bodyAny}

	result := exp.Check(nil, resp)

	assert.True(t, result.Ok, "Reason: %s", result.Reason)
}

func TestFieldEquals_Mismatch(t *testing.T) {
	exp := jsonSource.FieldEquals("name", "John")
	body := map[string]string{"name": "Jane"}
	var bodyAny any = body
	resp := &client.Response[any]{Body: &bodyAny}

	result := exp.Check(nil, resp)

	assert.False(t, result.Ok)
}

func TestFieldEquals_PathNotExists(t *testing.T) {
	exp := jsonSource.FieldEquals("nonexistent", "value")
	body := map[string]string{"name": "John"}
	var bodyAny any = body
	resp := &client.Response[any]{Body: &bodyAny}

	result := exp.Check(nil, resp)

	assert.False(t, result.Ok)
}

func TestFieldEquals_NestedPath(t *testing.T) {
	exp := jsonSource.FieldEquals("user.name", "John")
	body := map[string]interface{}{"user": map[string]string{"name": "John"}}
	var bodyAny any = body
	resp := &client.Response[any]{Body: &bodyAny}

	result := exp.Check(nil, resp)

	assert.True(t, result.Ok, "Reason: %s", result.Reason)
}

func TestFieldEquals_IntValue(t *testing.T) {
	exp := jsonSource.FieldEquals("count", 42)
	body := map[string]int{"count": 42}
	var bodyAny any = body
	resp := &client.Response[any]{Body: &bodyAny}

	result := exp.Check(nil, resp)

	assert.True(t, result.Ok, "Reason: %s", result.Reason)
}

func TestFieldEquals_BoolValue(t *testing.T) {
	exp := jsonSource.FieldEquals("active", true)
	body := map[string]bool{"active": true}
	var bodyAny any = body
	resp := &client.Response[any]{Body: &bodyAny}

	result := exp.Check(nil, resp)

	assert.True(t, result.Ok, "Reason: %s", result.Reason)
}

func TestFieldNotEmpty_Success(t *testing.T) {
	exp := jsonSource.FieldNotEmpty("name")
	body := map[string]string{"name": "John"}
	var bodyAny any = body
	resp := &client.Response[any]{Body: &bodyAny}

	result := exp.Check(nil, resp)

	assert.True(t, result.Ok)
}

func TestFieldNotEmpty_Empty(t *testing.T) {
	exp := jsonSource.FieldNotEmpty("name")
	body := map[string]string{"name": ""}
	var bodyAny any = body
	resp := &client.Response[any]{Body: &bodyAny}

	result := exp.Check(nil, resp)

	assert.False(t, result.Ok)
}

func TestFieldExists_Success(t *testing.T) {
	exp := jsonSource.FieldExists("name")
	body := map[string]string{"name": "John"}
	var bodyAny any = body
	resp := &client.Response[any]{Body: &bodyAny}

	result := exp.Check(nil, resp)

	assert.True(t, result.Ok)
}

func TestFieldExists_NotExists(t *testing.T) {
	exp := jsonSource.FieldExists("nonexistent")
	body := map[string]string{"name": "John"}
	var bodyAny any = body
	resp := &client.Response[any]{Body: &bodyAny}

	result := exp.Check(nil, resp)

	assert.False(t, result.Ok)
}

func TestFieldExists_NestedPath(t *testing.T) {
	exp := jsonSource.FieldExists("user.name")
	body := map[string]interface{}{"user": map[string]string{"name": "John"}}
	var bodyAny any = body
	resp := &client.Response[any]{Body: &bodyAny}

	result := exp.Check(nil, resp)

	assert.True(t, result.Ok)
}

func TestMakeMetadataExpectation_Success(t *testing.T) {
	exp := makeMetadataExpectation("x-request-id", "123")
	md := metadata.MD{}
	md.Append("x-request-id", "123")
	resp := &client.Response[any]{Metadata: md}

	result := exp.Check(nil, resp)

	assert.True(t, result.Ok)
}

func TestMakeMetadataExpectation_NotFound(t *testing.T) {
	exp := makeMetadataExpectation("x-request-id", "123")
	md := metadata.MD{}
	resp := &client.Response[any]{Metadata: md}

	result := exp.Check(nil, resp)

	assert.False(t, result.Ok)
	assert.Contains(t, result.Reason, "not found")
}

func TestMakeMetadataExpectation_WrongValue(t *testing.T) {
	exp := makeMetadataExpectation("x-request-id", "123")
	md := metadata.MD{}
	md.Append("x-request-id", "456")
	resp := &client.Response[any]{Metadata: md}

	result := exp.Check(nil, resp)

	assert.False(t, result.Ok)
}

func TestMakeMetadataExpectation_MultipleValues(t *testing.T) {
	exp := makeMetadataExpectation("x-values", "second")
	md := metadata.MD{}
	md.Append("x-values", "first")
	md.Append("x-values", "second")
	resp := &client.Response[any]{Metadata: md}

	result := exp.Check(nil, resp)

	assert.True(t, result.Ok)
}

func TestMakeMetadataExpectation_Error(t *testing.T) {
	exp := makeMetadataExpectation("x-request-id", "123")

	result := exp.Check(errors.New("test error"), nil)

	assert.False(t, result.Ok)
}

func TestExpectation_Names(t *testing.T) {
	tests := []struct {
		name     string
		exp      *expect.Expectation[*client.Response[any]]
		contains string
	}{
		{"NoError", makeNoErrorExpectation(), "No error"},
		{"Error", makeErrorExpectation(), "Error"},
		{"StatusCode", makeStatusCodeExpectation(codes.OK), "OK"},
		{"FieldEquals", jsonSource.FieldEquals("path", "val"), "path"},
		{"FieldNotEmpty", jsonSource.FieldNotEmpty("field"), "field"},
		{"FieldExists", jsonSource.FieldExists("field"), "field"},
		{"Metadata", makeMetadataExpectation("key", "val"), "key"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, tc.exp.Name, tc.contains)
		})
	}
}

func TestFieldEquals_ArrayAccess(t *testing.T) {
	exp := jsonSource.FieldEquals("items.0", "first")
	body := map[string]interface{}{"items": []string{"first", "second"}}
	var bodyAny any = body
	resp := &client.Response[any]{Body: &bodyAny}

	result := exp.Check(nil, resp)

	assert.True(t, result.Ok, "Reason: %s", result.Reason)
}

func TestFieldEquals_FloatValue(t *testing.T) {
	exp := jsonSource.FieldEquals("price", 19.99)
	body := map[string]float64{"price": 19.99}
	var bodyAny any = body
	resp := &client.Response[any]{Body: &bodyAny}

	result := exp.Check(nil, resp)

	assert.True(t, result.Ok, "Reason: %s", result.Reason)
}
