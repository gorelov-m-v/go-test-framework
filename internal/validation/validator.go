package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ozontech/allure-go/pkg/framework/provider"
)

type Validator struct {
	sCtx    provider.StepCtx
	dslName string
	failed  bool
}

func New(sCtx provider.StepCtx, dslName string) *Validator {
	return &Validator{
		sCtx:    sCtx,
		dslName: dslName,
	}
}

func (v *Validator) RequireNotNil(val any, name string) bool {
	if v.failed {
		return false
	}
	if val == nil || (reflect.ValueOf(val).Kind() == reflect.Ptr && reflect.ValueOf(val).IsNil()) {
		v.fail("%s DSL Error: %s is nil. Check test configuration.", v.dslName, name)
		return false
	}
	return true
}

func (v *Validator) RequireNotEmpty(val string, name string) bool {
	if v.failed {
		return false
	}
	if strings.TrimSpace(val) == "" {
		v.fail("%s DSL Error: %s is not set.", v.dslName, name)
		return false
	}
	return true
}

func (v *Validator) RequireNotEmptyWithHint(val string, name, hint string) bool {
	if v.failed {
		return false
	}
	if strings.TrimSpace(val) == "" {
		v.fail("%s DSL Error: %s is not set. %s", v.dslName, name, hint)
		return false
	}
	return true
}

func (v *Validator) RequireStruct(val any, name string) bool {
	if v.failed {
		return false
	}
	if val == nil {
		v.fail("%s DSL Error: %s must be a struct, got nil.", v.dslName, name)
		return false
	}
	t := reflect.TypeOf(val)
	if t == nil || t.Kind() != reflect.Struct {
		v.fail("%s DSL Error: %s must be a struct, got %T.", v.dslName, name, val)
		return false
	}
	return true
}

func (v *Validator) Require(condition bool, message string) bool {
	if v.failed {
		return false
	}
	if !condition {
		v.fail("%s DSL Error: %s", v.dslName, message)
		return false
	}
	return true
}

func (v *Validator) fail(format string, args ...any) {
	v.failed = true
	v.sCtx.Break(fmt.Sprintf(format, args...))
	v.sCtx.BrokenNow()
}
