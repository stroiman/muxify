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

func (s *GomegaSuite) Expect(actual interface{}, extra ...interface{}) gomega.Assertion {
	return s.gomega.Expect(actual, extra...)
}

func (s *GomegaSuite) SetupTest() {
	s.gomega = gomega.NewWithT(s.T())
}

type TmuxBaseTestSuite struct {
	GomegaSuite
	server TmuxServer
}

func (s *TmuxBaseTestSuite) SetupTest() {
	s.GomegaSuite.SetupTest()
	s.server = MustCreateTestServer()
}
