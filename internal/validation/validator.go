package validation

import (
	"reflect"
	"strings"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/errors"
)

type StepBreaker interface {
	Break(args ...interface{})
	BrokenNow()
}

type Validator struct {
	stepCtx StepBreaker
	dslName string
	failed  bool
}

func New(stepCtx provider.StepCtx, dslName string) *Validator {
	return &Validator{
		stepCtx: stepCtx,
		dslName: dslName,
	}
}

func NewWithBreaker(stepCtx StepBreaker, dslName string) *Validator {
	return &Validator{
		stepCtx: stepCtx,
		dslName: dslName,
	}
}

func (v *Validator) RequireNotNil(val any, name string) bool {
	if v.failed {
		return false
	}
	if val == nil || (reflect.ValueOf(val).Kind() == reflect.Ptr && reflect.ValueOf(val).IsNil()) {
		v.fail(errors.NilClient(v.dslName, name))
		return false
	}
	return true
}

func (v *Validator) RequireNotEmpty(val string, name string) bool {
	if v.failed {
		return false
	}
	if strings.TrimSpace(val) == "" {
		v.fail(errors.NotSet(v.dslName, name))
		return false
	}
	return true
}

func (v *Validator) RequireNotEmptyWithHint(val string, name, hint string) bool {
	if v.failed {
		return false
	}
	if strings.TrimSpace(val) == "" {
		v.fail(errors.NotSetWithHint(v.dslName, name, hint))
		return false
	}
	return true
}

func (v *Validator) RequireStruct(val any, name string) bool {
	if v.failed {
		return false
	}
	if val == nil {
		v.fail(errors.MustBeStruct(v.dslName, name, nil))
		return false
	}
	t := reflect.TypeOf(val)
	if t == nil || t.Kind() != reflect.Struct {
		v.fail(errors.MustBeStruct(v.dslName, name, val))
		return false
	}
	return true
}

func (v *Validator) Require(condition bool, message string) bool {
	if v.failed {
		return false
	}
	if !condition {
		v.fail(errors.Custom(v.dslName, message))
		return false
	}
	return true
}

func (v *Validator) fail(msg string) {
	v.failed = true
	v.stepCtx.Break(msg)
	v.stepCtx.BrokenNow()
}
