package extension

import (
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type BaseSuite struct {
	suite.Suite
	tExt *TExtension
}

func (s *BaseSuite) BeforeEach(t provider.T) {
	s.tExt = nil
}

func (s *BaseSuite) T(t provider.T) *TExtension {
	if s.tExt == nil {
		s.tExt = NewTExtension(t)
	}
	return s.tExt
}

func (s *BaseSuite) Step(t provider.T, name string, fn func(sCtx provider.StepCtx), params ...*allure.Parameter) {
	s.T(t).WithNewStep(name, fn, params...)
}

func (s *BaseSuite) AsyncStep(t provider.T, name string, fn func(sCtx provider.StepCtx), params ...*allure.Parameter) {
	s.T(t).WithNewAsyncStep(name, fn, params...)
}
