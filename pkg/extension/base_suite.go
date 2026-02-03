package extension

import (
	"sync"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type cleanupItem struct {
	name string
	fn   func(sCtx provider.StepCtx)
}

type BaseSuite struct {
	suite.Suite
	tExt     *TExtension
	asyncWg  sync.WaitGroup
	currentT provider.T
	cleanups []cleanupItem
}

func (s *BaseSuite) BeforeEach(t provider.T) {
	s.tExt = nil
	s.currentT = t
	s.cleanups = nil
}

func (s *BaseSuite) DeferCleanup(name string, fn func(sCtx provider.StepCtx)) {
	s.cleanups = append(s.cleanups, cleanupItem{name: name, fn: fn})
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
	for i := len(s.cleanups) - 1; i >= 0; i-- {
		c := s.cleanups[i]
		s.Step(t, c.name, c.fn)
	}
}
