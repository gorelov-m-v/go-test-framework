package extension

import (
	"sync"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type BaseSuite struct {
	suite.Suite
	tExt     *TExtension
	asyncWg  sync.WaitGroup
	cleanup  func(t provider.T)
	currentT provider.T
}

func (s *BaseSuite) BeforeEach(t provider.T) {
	s.tExt = nil
	s.cleanup = nil
	s.currentT = t
}

func (s *BaseSuite) Cleanup(fn func(t provider.T)) {
	if s.currentT != nil {
		t := s.currentT
		t.Cleanup(func() {
			s.asyncWg.Wait()
			fn(t)
		})
	} else {
		s.cleanup = fn
	}
}

func (s *BaseSuite) T(t provider.T) *TExtension {
	if s.tExt == nil {
		s.tExt = NewTExtension(t)
	}
	return s.tExt
}

func (s *BaseSuite) Step(t provider.T, name string, fn func(sCtx provider.StepCtx), params ...*allure.Parameter) {
	s.asyncWg.Wait()
	s.T(t).WithNewStep(name, fn, params...)
}

func (s *BaseSuite) AsyncStep(t provider.T, name string, fn func(sCtx provider.StepCtx), params ...*allure.Parameter) {
	s.asyncWg.Add(1)
	s.T(t).WithNewAsyncStep(name, func(sCtx provider.StepCtx) {
		defer s.asyncWg.Done()
		fn(sCtx)
	}, params...)
}

func (s *BaseSuite) AfterEach(t provider.T) {
	s.asyncWg.Wait()
	if s.cleanup != nil {
		s.cleanup(t)
	}
}
