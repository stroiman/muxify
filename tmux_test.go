package main_test

import (
	"fmt"
	"os/exec"
	"testing"

	. "github.com/stroiman/muxify"

	g "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

type TmuxTestSuite struct {
	TmuxBaseTestSuite
}

func (s *TmuxTestSuite) TestRunningSessionsWhenServerIsNotStarted() {
	s.server.SocketName = CreateRandomName()
	sessions, err := s.server.GetRunningSessions()
	s.Expect(sessions).To(g.BeEmpty())
	s.Expect(err).ToNot(g.HaveOccurred())
}

func TestTmux(t *testing.T) {
	suite.Run(t, new(TmuxTestSuite))
}

type TmuxRunningServerTestSuite struct {
	TmuxBaseTestSuite
	sessionName string
}

func (s *TmuxRunningServerTestSuite) SetupTest() {
	s.TmuxBaseTestSuite.SetupTest()
	fmt.Println("Server", s.server)
	s.sessionName = CreateRandomName()
	output, err := s.server.Command("new-session", "-s", s.sessionName, "-d").Output()
	if err != nil {
		fmt.Println("Error!", string(output), err.Error())
		if e, ok := err.(*exec.ExitError); ok {
			fmt.Println("Stderr: ", string(e.Stderr))
		}
	}
	s.Assert().NoError(err)
	// s.Expect(err).ToNot(g.HaveOccurred())
}

func (s *TmuxRunningServerTestSuite) TearDownTest() {
	sessions, err := s.server.GetRunningSessions()
	s.Expect(err).ToNot(g.HaveOccurred())
	session, ok := TmuxSessions(sessions).FindByName(s.sessionName)
	if ok {
		s.Expect(s.server.KillSession(session)).To(g.Succeed())
	}
}

func (s *TmuxRunningServerTestSuite) TestRunningSessionsHasAtLeastOneElement() {
	result, err := s.server.GetRunningSessions()
	s.Expect(err).ToNot(g.HaveOccurred())
	s.Expect(result).To(g.ContainElement(g.HaveField("Name", s.sessionName)))
}

func (s *TmuxRunningServerTestSuite) TestKillServer() {
	result, err := s.server.GetRunningSessions()
	s.Expect(err).ToNot(g.HaveOccurred())
	session, _ := TmuxSessions(result).FindByName(s.sessionName)
	s.Expect(s.server.KillSession(session)).To(g.Succeed())
	result, err = s.server.GetRunningSessions()
	s.Expect(err).ToNot(g.HaveOccurred())
	s.Expect(result).ToNot(g.ContainElement(g.HaveField("Name", s.sessionName)))
}

func TestTmuxRunningServer(t *testing.T) {
	suite.Run(t, new(TmuxRunningServerTestSuite))
}
