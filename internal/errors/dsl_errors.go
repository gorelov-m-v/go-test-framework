package errors

import "fmt"

func NilClient(dslName, clientName string) string {
	return fmt.Sprintf("%s DSL Error: %s is nil. Check test configuration.", dslName, clientName)
}

func NotSet(dslName, fieldName string) string {
	return fmt.Sprintf("%s DSL Error: %s is not set.", dslName, fieldName)
}

func NotSetWithHint(dslName, fieldName, hint string) string {
	return fmt.Sprintf("%s DSL Error: %s is not set. %s", dslName, fieldName, hint)
}

func MustBeStruct(dslName, fieldName string, val any) string {
	if val == nil {
		return fmt.Sprintf("%s DSL Error: %s must be a struct, got nil.", dslName, fieldName)
	}
	return fmt.Sprintf("%s DSL Error: %s must be a struct, got %T.", dslName, fieldName, val)
}

func Custom(dslName, message string) string {
	return fmt.Sprintf("%s DSL Error: %s", dslName, message)
}

func ExpectationsAfterSend(dslName string) string {
	return dslName + " DSL Error: Expectations must be added before Send()."
}

func InvalidGenericType(dslName string, got any) string {
	return fmt.Sprintf("%s DSL Error: Query type parameter must be a struct, got %T. Check your NewQuery[T] generic type.", dslName, got)
}

func ConflictingExpectations(dslName, method, conflictsWith string) string {
	return fmt.Sprintf("%s DSL Error: %s cannot be used with %s", dslName, method, conflictsWith)
}

func ExpectationOrder(dslName, method, reason string) string {
	return fmt.Sprintf("%s DSL Error: %s cannot be used after %s", dslName, method, reason)
}

func MissingConfig(dslName, configName, hint string) string {
	return fmt.Sprintf("%s DSL Error: %s requested but no %s configured. %s", dslName, configName, configName, hint)
}

func ContractValidationMissingSpec(dslName string) string {
	return fmt.Sprintf("%s DSL Error: Contract validation requested but no contractSpec configured for this client. Add 'contractSpec' to your HTTP client config.", dslName)
}

func MethodAfterSend(dslName, method string) string {
	return fmt.Sprintf("%s DSL Error: %s must be called before Send()", dslName, method)
}
