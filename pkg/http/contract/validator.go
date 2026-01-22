package contract

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type Validator struct {
	spec *openapi3.T
}

func NewValidator(specPath string) (*Validator, error) {
	spec, err := Load(specPath)
	if err != nil {
		return nil, err
	}
	return &Validator{spec: spec}, nil
}

func NewValidatorFromSpec(spec *openapi3.T) *Validator {
	return &Validator{spec: spec}
}

func (v *Validator) ValidateResponse(method, path string, statusCode int, body []byte) error {
	op, err := v.findOperation(method, path)
	if err != nil {
		return err
	}

	return v.validateResponseBody(op, statusCode, body)
}

func (v *Validator) ValidateResponseBySchema(schemaName string, body []byte) error {
	schema, err := v.findSchema(schemaName)
	if err != nil {
		return err
	}

	return v.validateAgainstSchema(schema, body)
}

func (v *Validator) findOperation(method, path string) (*openapi3.Operation, error) {
	method = strings.ToUpper(method)

	normalizedPath := normalizePath(path)

	for specPath, pathItem := range v.spec.Paths.Map() {
		if !pathMatches(specPath, normalizedPath) {
			continue
		}

		op := pathItem.GetOperation(method)
		if op == nil {
			return nil, &ValidationError{
				Type:    ErrOperationNotFound,
				Message: fmt.Sprintf("method %s not defined for path %s in spec", method, specPath),
			}
		}
		return op, nil
	}

	return nil, &ValidationError{
		Type:    ErrPathNotFound,
		Message: fmt.Sprintf("path %s not found in spec", path),
	}
}

func (v *Validator) findSchema(name string) (*openapi3.Schema, error) {
	if v.spec.Components == nil || v.spec.Components.Schemas == nil {
		return nil, &ValidationError{
			Type:    ErrSchemaNotFound,
			Message: fmt.Sprintf("schema %s not found: no schemas defined in spec", name),
		}
	}

	schemaRef, ok := v.spec.Components.Schemas[name]
	if !ok {
		return nil, &ValidationError{
			Type:    ErrSchemaNotFound,
			Message: fmt.Sprintf("schema %s not found in spec", name),
		}
	}

	if schemaRef.Value == nil {
		return nil, &ValidationError{
			Type:    ErrSchemaNotFound,
			Message: fmt.Sprintf("schema %s is empty", name),
		}
	}

	return schemaRef.Value, nil
}

func (v *Validator) validateResponseBody(op *openapi3.Operation, statusCode int, body []byte) error {
	if op.Responses == nil {
		return nil
	}

	responseRef := op.Responses.Status(statusCode)
	if responseRef == nil {
		responseRef = op.Responses.Default()
	}

	if responseRef == nil || responseRef.Value == nil {
		return &ValidationError{
			Type:    ErrResponseNotDefined,
			Message: fmt.Sprintf("response %d not defined in spec", statusCode),
		}
	}

	response := responseRef.Value

	if response.Content == nil || len(response.Content) == 0 {
		if len(body) > 0 {
			return nil
		}
		return nil
	}

	mediaType := response.Content.Get("application/json")
	if mediaType == nil {
		mediaType = response.Content.Get("*/*")
	}
	if mediaType == nil {
		for _, mt := range response.Content {
			mediaType = mt
			break
		}
	}

	if mediaType == nil || mediaType.Schema == nil || mediaType.Schema.Value == nil {
		return nil
	}

	return v.validateAgainstSchema(mediaType.Schema.Value, body)
}

func (v *Validator) validateAgainstSchema(schema *openapi3.Schema, body []byte) error {
	if len(body) == 0 {
		return nil
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return &ValidationError{
			Type:    ErrInvalidJSON,
			Message: fmt.Sprintf("invalid JSON: %v", err),
		}
	}

	if err := schema.VisitJSON(data); err != nil {
		return &ValidationError{
			Type:    ErrSchemaValidation,
			Message: formatSchemaError(err),
			Cause:   err,
		}
	}

	return nil
}

func normalizePath(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	return path
}

func pathMatches(specPath, actualPath string) bool {
	specParts := strings.Split(specPath, "/")
	actualParts := strings.Split(actualPath, "/")

	if len(specParts) != len(actualParts) {
		return false
	}

	for i, specPart := range specParts {
		if strings.HasPrefix(specPart, "{") && strings.HasSuffix(specPart, "}") {
			continue
		}
		if specPart != actualParts[i] {
			return false
		}
	}

	return true
}

func formatSchemaError(err error) string {
	if err == nil {
		return ""
	}

	msg := err.Error()

	msg = strings.ReplaceAll(msg, "doesn't match schema", "does not match schema")

	return msg
}
