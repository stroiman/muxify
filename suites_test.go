package main_test

import (
	. "github.com/stroiman/muxify"

	"github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

type GomegaSuite struct {
	suite.Suite
	gomega gomega.Gomega
}

type TmuxBaseTestSuite struct {
	GomegaSuite
	server TmuxServer
}

func (s *TmuxBaseTestSuite) SetupTest() {
	s.server = MustCreateTestServer()
}

func (s *GomegaSuite) Expect(actual interface{}, extra ...interface{}) gomega.Assertion {
	if s.gomega == nil {
		s.gomega = gomega.NewWithT(s.T())
	}
	return s.gomega.Expect(actual, extra...)
}
