package contract

import "fmt"

type ErrorType int

const (
	ErrPathNotFound ErrorType = iota + 1
	ErrOperationNotFound
	ErrResponseNotDefined
	ErrSchemaNotFound
	ErrInvalidJSON
	ErrSchemaValidation
)

func (e ErrorType) String() string {
	switch e {
	case ErrPathNotFound:
		return "PATH_NOT_FOUND"
	case ErrOperationNotFound:
		return "OPERATION_NOT_FOUND"
	case ErrResponseNotDefined:
		return "RESPONSE_NOT_DEFINED"
	case ErrSchemaNotFound:
		return "SCHEMA_NOT_FOUND"
	case ErrInvalidJSON:
		return "INVALID_JSON"
	case ErrSchemaValidation:
		return "SCHEMA_VALIDATION_FAILED"
	default:
		return "UNKNOWN"
	}
}

type ValidationError struct {
	Type    ErrorType
	Message string
	Cause   error
}

func (e *ValidationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s", e.Type, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Cause
}

func IsPathNotFound(err error) bool {
	if ve, ok := err.(*ValidationError); ok {
		return ve.Type == ErrPathNotFound
	}
	return false
}

func IsSchemaValidationError(err error) bool {
	if ve, ok := err.(*ValidationError); ok {
		return ve.Type == ErrSchemaValidation
	}
	return false
}
