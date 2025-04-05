package main_test

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/stretchr/testify/suite"
	. "github.com/stroiman/muxify"
)

type ProjectTestSuite struct {
	suite.Suite
	gomega        gomega.Gomega
	server        TmuxServer
	knownSessions []TmuxSession
}

func (s *ProjectTestSuite) Expect(actual interface{}, extra ...interface{}) Assertion {
	return s.gomega.Expect(actual, extra...)
}

func (s *ProjectTestSuite) Eventually(actual interface{}, extra ...interface{}) AsyncAssertion {
	if s.gomega == nil {
		s.gomega = gomega.NewWithT(s.T())
	}
	return s.gomega.Eventually(actual, extra...)
}

func (s *ProjectTestSuite) SetupTest() {
	s.gomega = gomega.NewWithT(s.T())
	s.knownSessions = nil
}

func (s *ProjectTestSuite) TearDownTest() {
	for _, knownSession := range s.knownSessions {
		s.server.KillSession(knownSession)
	}
}

func (s *ProjectTestSuite) SetupSuite() {
	s.server = MustCreateTestServer()
}

func TestProjectType(t *testing.T) {
	suite.Run(t, new(ProjectTestSuite))
}

func (s *ProjectTestSuite) TearDownSuite() {
	err := s.server.KillServer()
	s.Assert().Error(err, "If killing doesn't fail, a test did not clean up correctly")
}

type ProjectEnsureStartedTestSuite struct {
	ProjectTestSuite
	dir string
}

func (s *ProjectEnsureStartedTestSuite) SetupTest() {
	s.ProjectTestSuite.SetupTest()
	var err error
	s.dir, err = os.MkdirTemp("", "muxify-test-")
	s.Assert().NoError(err)
}

func (s *ProjectEnsureStartedTestSuite) TearDownTest() {
	s.Assert().NoError(os.Remove(s.dir))
	s.ProjectTestSuite.TearDownTest()
}

func TestProjectEnsureStarted(t *testing.T) {
	suite.Run(t, new(ProjectEnsureStartedTestSuite))
}

func (s *ProjectEnsureStartedTestSuite) TestStartWhenNotAlreadyStarted() {
	proj := CreateProject()
	session := s.handleProjectStart(proj.EnsureStarted(s.server))
	s.Expect(session).To(BeStarted())
}

func (s *ProjectEnsureStartedTestSuite) TestStartProjectWithOnePane() {
	proj := CreateProject()
	session := s.handleProjectStart(proj.EnsureStarted(s.server))

	s.Expect(
		session.GetPanes(),
	).To(HaveExactElements(HaveField("Id", MatchRegexp("^\\%\\d+$"))))
}

func (s *ProjectEnsureStartedTestSuite) TestWorkingDirectory() {
	proj := CreateProject()
	proj.WorkingDirectory = s.dir
	session := s.handleProjectStart(proj.EnsureStarted(s.server))
	cm := MustStartControlMode(s.server, session)
	defer cm.MustClose()

	s.Expect(
		session.RunShellCommand("echo $PWD"),
	).To(Succeed())
	s.Eventually(
		s.getOutputLinesFromEvents(s.getOutputEvents(GetLines(cm.stdout))),
	).Should(Receive(Equal(s.dir)))
}

func (s *ProjectEnsureStartedTestSuite) TestWorkingDirectoryMultipleWindows() {
	proj := CreateProjectWithWindowNames(
		"Window-1",
		"Window-2",
		"Window-3",
	)
	proj.WorkingDirectory = s.dir
	session := s.handleProjectStart(proj.EnsureStarted(s.server))
	cm := MustStartControlMode(s.server, session)
	defer cm.MustClose()

	outputStream := s.getOutputLinesFromEvents(s.getOutputEvents(GetLines(cm.stdout)))
	windows := session.MustGetWindows()
	for winNo, win := range windows {
		s.Expect(
			win.RunShellCommand("echo $PWD"),
		).To(Succeed())
		s.Eventually(
			outputStream,
		).Should(Receive(Equal(s.dir)), fmt.Sprintf("Window no: %d", winNo+1))
	}
}

func (s *ProjectEnsureStartedTestSuite) TestWorkingDirectoryMultiplePanes() {
	proj := CreateProject()
	proj.AppendNamedWindow("Window-1").
		AppendPane(proj.CreatePaneWithCommands("pane-1")).
		AppendPane(proj.CreatePaneWithCommands("pane-2"))
	proj.WorkingDirectory = s.dir
	session := s.handleProjectStart(proj.EnsureStarted(s.server))
	cm := MustStartControlMode(s.server, session)
	defer cm.MustClose()

	outputStream := s.getOutputLinesFromEvents(s.getOutputEvents(GetLines(cm.stdout)))
	windows := session.MustGetWindows()
	for winNo, window := range windows {
		panes := window.MustGetPanes()
		for paneNo, pane := range panes {
			s.Expect(
				pane.RunShellCommand("echo $PWD"),
			).To(Succeed())
			s.Eventually(
				outputStream,
			).Should(Receive(Equal(s.dir)), fmt.Sprintf("Win, pane no: %d, %d", winNo+1, paneNo+1))
		}
	}
}
func (s *ProjectEnsureStartedTestSuite) TestWorkingFolderForTask() {
	subdir := path.Join(s.dir, "sub_dir")
	os.Mkdir(subdir, 0700)
	defer func() { os.Remove(subdir) }()
	proj := CreateProject(ProjectWorkingDir(s.dir))
	pane1id := proj.CreatePane("pane-1")
	pane2id := proj.CreatePane("pane-2", TaskWorkingDir("./sub_dir"))
	win := proj.AppendNamedWindow("Window-1")
	win.AppendPane(pane1id)
	win.AppendPane(pane2id)
	proj.WorkingDirectory = s.dir
	session := s.handleProjectStart(proj.EnsureStarted(s.server))
	cm := MustStartControlMode(s.server, session)
	defer cm.MustClose()

	outputStream := s.getOutputLinesFromEvents(s.getOutputEvents(GetLines(cm.stdout)))
	windows := session.MustGetWindows()
	for winNo, window := range windows {
		panes := window.MustGetPanes()
		pane1 := panes.FindByTitle(pane1id)
		pane2 := panes.FindByTitle(pane2id)

		s.Expect(
			pane1.RunShellCommand("echo $PWD"),
		).To(Succeed())
		s.Eventually(
			outputStream,
		).Should(Receive(Equal(s.dir)), fmt.Sprintf("Win, pane no: %d, %d", winNo+1, 1))
		s.Expect(
			pane2.RunShellCommand("echo $PWD"),
		).To(Succeed())
		s.Eventually(
			outputStream,
		).Should(Receive(Equal(subdir)), fmt.Sprintf("Win, pane no: %d, %d", winNo+1, 2))
	}
}

func (s *ProjectEnsureStartedTestSuite) TestTaskSubfolderInNewWindow() {
	subdir := path.Join(s.dir, "sub_dir")
	os.Mkdir(subdir, 0700)
	defer func() { os.Remove(subdir) }()
	proj := CreateProject(ProjectWorkingDir(s.dir))
	pane1id := proj.CreatePane("pane-1")
	pane2id := proj.CreatePane("pane-2", TaskWorkingDir("./sub_dir"))
	proj.AppendNamedWindow("Window-1").AppendPane(pane1id)
	proj.AppendNamedWindow("Window-2").AppendPane(pane2id)
	proj.WorkingDirectory = s.dir
	session := s.handleProjectStart(proj.EnsureStarted(s.server))
	cm := MustStartControlMode(s.server, session)
	defer cm.MustClose()

	outputStream := s.getOutputLinesFromEvents(s.getOutputEvents(GetLines(cm.stdout)))
	panes := make(TmuxPanes, 0)
	windows := session.MustGetWindows()
	for _, window := range windows {
		panes = append(panes, window.MustGetPanes()...)
	}
	pane1 := panes.FindByTitle(pane1id)
	pane2 := panes.FindByTitle(pane2id)

	s.Expect(
		pane1.RunShellCommand("echo $PWD"),
	).To(Succeed())
	s.Eventually(
		outputStream,
	).Should(Receive(Equal(s.dir)), fmt.Sprintf("Pane: %s", pane1id))
	s.Expect(
		pane2.RunShellCommand("echo $PWD"),
	).To(Succeed())
	s.Eventually(
		outputStream,
	).Should(Receive(Equal(subdir)), fmt.Sprintf("Pane: %s", pane2id))
}

func (s *ProjectEnsureStartedTestSuite) TestWorkingFolderForFirstTask() {
	subdir := path.Join(s.dir, "sub_dir")
	os.Mkdir(subdir, 0700)
	defer func() { os.Remove(subdir) }()
	proj := CreateProject(ProjectWorkingDir(s.dir))
	pane1id := proj.CreatePane("pane-1", TaskWorkingDir("./sub_dir"))
	pane2id := proj.CreatePane("pane-2")
	win := proj.AppendNamedWindow("Window-1")
	win.AppendPane(pane1id)
	win.AppendPane(pane2id)
	proj.WorkingDirectory = s.dir
	session := s.handleProjectStart(proj.EnsureStarted(s.server))
	cm := MustStartControlMode(s.server, session)
	defer cm.MustClose()

	outputStream := s.getOutputLinesFromEvents(s.getOutputEvents(GetLines(cm.stdout)))
	windows := session.MustGetWindows()
	for winNo, window := range windows {
		panes := window.MustGetPanes()
		pane1 := panes.FindByTitle(pane1id)
		pane2 := panes.FindByTitle(pane2id)

		s.Expect(
			pane1.RunShellCommand("echo $PWD"),
		).To(Succeed())
		s.Eventually(
			outputStream,
		).Should(Receive(Equal(subdir)), fmt.Sprintf("Win, pane no: %d, %d", winNo+1, 1))
		s.Expect(
			pane2.RunShellCommand("echo $PWD"),
		).To(Succeed())
		s.Eventually(
			outputStream,
		).Should(Receive(Equal(s.dir)), fmt.Sprintf("Win, pane no: %d, %d", winNo+1, 2))
	}
}

func (s *ProjectEnsureStartedTestSuite) TestReturnSameSessionIfStarted() {
	proj := CreateProject()
	s1 := s.handleProjectStart(proj.EnsureStarted(s.server))
	s2 := s.handleProjectStart(proj.EnsureStarted(s.server))
	s.Expect(s1.Id).To(Equal(s2.Id))
}

func (s *ProjectEnsureStartedTestSuite) TestWindowName() {
	proj := CreateProjectWithWindowNames("Window-1")
	s1 := s.handleProjectStart(proj.EnsureStarted(s.server))
	s.Expect(
		s.server.GetWindowsForSession(s1),
	).To(HaveExactElements(HaveField("Name", "Window-1")))
}

func (s *ProjectEnsureStartedTestSuite) TestMultipleWindows() {
	proj := CreateProjectWithWindowNames(
		"Window-1",
		"Window-2",
		"Window-3",
	)
	session := s.handleProjectStart(proj.EnsureStarted(s.server))
	windows, err := s.server.GetWindowsForSession(session)
	s.Expect(err).ToNot(HaveOccurred())
	s.Expect(windows).To(HaveExactElements(
		HaveField("Name", "Window-1"),
		HaveField("Name", "Window-2"),
		HaveField("Name", "Window-3"),
	))

	s.Expect(s.server.GetCurrentWindowIndexForSession(session)).To(Equal(0))
	s.Expect(windows[0].Index()).To(Equal(0))
	s.Expect(windows[1].Index()).To(Equal(1))
	s.Expect(windows[2].Index()).To(Equal(2))
}

func (s *ProjectEnsureStartedTestSuite) TestRecreateMissingWindowsOnRunningSession() {
	proj := CreateProjectWithWindowNames(
		"Window-1",
		"Window-2",
	)
	session := s.handleProjectStart(proj.EnsureStarted(s.server))
	proj.AppendNamedWindow("Window-3")
	s.handleProjectStart(proj.EnsureStarted(s.server))
	s.Expect(s.server.GetWindowsForSession(session)).To(HaveExactElements(
		HaveField("Name", "Window-1"),
		HaveField("Name", "Window-2"),
		HaveField("Name", "Window-3"),
	))
}

func (s *ProjectEnsureStartedTestSuite) TestRecreateWindowsOutOfOrder() {
	proj := CreateProjectWithWindowNames("Window-4", "Window-1", "Window-3")
	session := s.handleProjectStart(proj.EnsureStarted(s.server))
	proj.ReplaceWindowNames("Window-1", "Window-2", "Window-3", "Window-4")
	s.handleProjectStart(proj.EnsureStarted(s.server))
	s.Expect(s.server.GetWindowsForSession(session)).To(HaveExactElements(
		HaveField("Name", "Window-1"),
		HaveField("Name", "Window-2"),
		HaveField("Name", "Window-3"),
		HaveField("Name", "Window-4"),
	))
}

func (s *ProjectEnsureStartedTestSuite) TestRecreateMissingWindowsAgain_IsThisADuplicateTest() {
	proj := CreateProjectWithWindowNames("Window-2")
	session := s.handleProjectStart(proj.EnsureStarted(s.server))
	proj.ReplaceWindowNames("Window-1", "Window-2")
	s.handleProjectStart(proj.EnsureStarted(s.server))
	s.Expect(s.server.GetWindowsForSession(session)).To(HaveExactElements(
		HaveField("Name", "Window-1"),
		HaveField("Name", "Window-2"),
	))
}

func (s *ProjectEnsureStartedTestSuite) TestCreatePanesWithCorrectNames() {
	proj := CreateProject()
	proj.AppendNamedWindow("Window-1").
		AppendPane(proj.CreatePaneWithCommands("Pane-1")).
		AppendPane(proj.CreatePaneWithCommands("Pane-2"))
	proj.AppendNamedWindow("Window-2").
		AppendPane(proj.CreatePaneWithCommands("Pane-3")).
		AppendPane(proj.CreatePaneWithCommands("Pane-4"))
	s.handleProjectStart(proj.EnsureStarted(s.server))
	expected := []T{
		{"Window-1", "Pane-1"},
		{"Window-1", "Pane-2"},
		{"Window-2", "Pane-3"},
		{"Window-2", "Pane-4"},
	}
	s.Expect(s.server.GetWindowAndPaneNames()).To(HaveExactElements(expected))
}

func (s *ProjectEnsureStartedTestSuite) TestPaneLayoutTopBottomLeftRight() {
	proj := CreateProject()
	session := s.handleProjectStart(proj.EnsureStarted(s.server))
	panes := session.MustGetPanes()
	layout := panes[0].Layout
	s.Expect(layout.Top).To(BeNumerically("<", layout.Bottom), "Top < Bottom")
	s.Expect(layout.Left).To(BeNumerically("<", layout.Right), "Left < Right")
}

func (s *ProjectEnsureStartedTestSuite) TestHorizontalLayout() {
	proj := CreateProject()
	proj.AppendNamedWindow("Window-1").
		SetHorizontalLayout().
		AppendPane(proj.CreatePaneWithCommands("Pane-1")).
		AppendPane(proj.CreatePaneWithCommands("Pane-2"))
	session := s.handleProjectStart(proj.EnsureStarted(s.server))
	panes := session.MustGetPanes()
	s.Expect(panes[0].Title).To(Equal("Pane-1"))
	s.Expect(panes[1].Title).To(Equal("Pane-2"))
	s.Expect(panes[0].Layout.Top).To(Equal(0), "First pane top")
	s.Expect(panes[0].Layout.Left).To(Equal(0), "First pane left")
	s.Expect(panes[1].Layout.Top).To(Equal(0), "Second pane top")
	s.Expect(panes[1].Layout.Left).ToNot(Equal(0), "Second pane left")
}
func (s *ProjectEnsureStartedTestSuite) TestVerticalLayout() {
	proj := CreateProject()
	proj.AppendNamedWindow("Window-1").
		SetVerticalLayout().
		AppendPane(proj.CreatePaneWithCommands("Pane-1")).
		AppendPane(proj.CreatePaneWithCommands("Pane-2"))
	session := s.handleProjectStart(proj.EnsureStarted(s.server))
	panes := session.MustGetPanes()
	s.Expect(panes[0].Layout.Top).To(Equal(0), "First pane top")
	s.Expect(panes[0].Layout.Left).To(Equal(0), "First pane left")
	s.Expect(panes[1].Layout.Top).ToNot(Equal(0), "Second pane top")
	s.Expect(panes[1].Layout.Left).To(Equal(0), "Second pane left")
}

func (s *ProjectEnsureStartedTestSuite) TestEnsureStartedDoesntAddMorePanes() {
	proj := CreateProject()
	proj.AppendNamedWindow("Window-1").
		AppendPane(proj.CreatePaneWithCommands("Pane-1")).
		AppendPane(proj.CreatePaneWithCommands("Pane-2"))
	proj.AppendNamedWindow("Window-2").
		AppendPane(proj.CreatePaneWithCommands("Pane-3")).
		AppendPane(proj.CreatePaneWithCommands("Pane-4"))
	s.handleProjectStart(proj.EnsureStarted(s.server))
	s.handleProjectStart(proj.EnsureStarted(s.server))

	expected := []T{
		{"Window-1", "Pane-1"},
		{"Window-1", "Pane-2"},
		{"Window-2", "Pane-3"},
		{"Window-2", "Pane-4"},
	}
	s.Expect(s.server.GetWindowAndPaneNames()).To(HaveExactElements(expected))
}

func (s *ProjectEnsureStartedTestSuite) TestExecuteCommandsInConfiguration() {
	proj := CreateProject()
	proj.AppendNamedWindow("Window-1").
		AppendPane(proj.CreatePaneWithCommands("Pane-1", "echo \"Foo\"")).
		AppendPane(proj.CreatePaneWithCommands("Pane-2", "echo \"Bar\""))
	session := s.handleProjectStart(proj.EnsureStarted(s.server))
	cm := MustStartControlMode(s.server, session)
	defer cm.MustClose()
	outputEvents := s.getOutputLinesFromEvents(s.getOutputEvents(GetLines(cm.stdout)))
	panes, err := session.GetPanes()
	s.Expect(err).ToNot(HaveOccurred())
	panes[0].MustRunShellCommand("echo \"DONE 1\"")
	s.Eventually(outputEvents).Should(Receive(Equal("DONE 1")))
	panes[1].MustRunShellCommand("echo \"DONE 2\"")
	s.Eventually(outputEvents).Should(Receive(Equal("DONE 2")))
	output1 := s.server.Command("capture-pane", "-p", "-t", panes[0].Id).MustOutput()
	output2 := s.server.Command("capture-pane", "-p", "-t", panes[1].Id).MustOutput()
	s.Expect(output1).To(MatchRegexp("(?m:^Foo$)"))
	s.Expect(output2).To(MatchRegexp("(?m:^Bar$)"))
}

func (s *ProjectEnsureStartedTestSuite) TestVerifyFirstPanesDoesntRunTwiceOnRestart() {
	proj := CreateProject()
	proj.AppendNamedWindow("Window-1").
		AppendPane(proj.CreatePaneWithCommands("Pane-1", "echo \"Foo\"")).
		AppendPane(proj.CreatePaneWithCommands("Pane-2", "echo \"Bar\""))
	session := s.handleProjectStart(proj.EnsureStarted(s.server))
	cm := MustStartControlMode(s.server, session)
	defer cm.MustClose()
	outputEvents := s.getOutputLinesFromEvents(s.getOutputEvents(GetLines(cm.stdout)))
	panes, err := session.GetPanes()
	s.Expect(err).ToNot(HaveOccurred())

	// Wait for the commands to have executed
	panes[0].MustRunShellCommand("echo \"DONE 1\"")
	s.Eventually(outputEvents).Should(Receive(Equal("DONE 1")))

	// Start this again
	s.handleProjectStart(proj.EnsureStarted(s.server))

	// Wait for all commands to have executed
	panes[0].MustRunShellCommand("echo \"DONE 1\"")
	s.Eventually(outputEvents).Should(Receive(Equal("DONE 1")))

	panes, err = session.GetPanes()
	s.Expect(err).ToNot(HaveOccurred())

	output1 := s.server.Command("capture-pane", "-p", "-t", panes[0].Id).MustOutput()
	output2 := s.server.Command("capture-pane", "-p", "-t", panes[1].Id).MustOutput()

	var exp *regexp.Regexp = regexp.MustCompile(`(?m:^(?:Foo|Bar))`)
	s.Expect(exp.FindAllString(string(output1), -1)).To(Equal([]string{"Foo"}))
	s.Expect(exp.FindAllString(string(output2), -1)).To(Equal([]string{"Bar"}))
}

func BeStarted() types.GomegaMatcher {
	return HaveField("Id", MatchRegexp("^\\$\\d+"))
}

func (s *ProjectTestSuite) getOutputEvents(lines <-chan string) <-chan TmuxOutputEvent {

	c := make(chan TmuxOutputEvent)
	r := regexp.MustCompile("^\\%output ([^ ]+) (.*)")
	go func() {
		defer close(c)
		for line := range lines {
			m := r.FindStringSubmatch(line)
			if m != nil {
				event := TmuxOutputEvent{
					PaneId: m[1],
					Data:   m[2],
				}
				c <- event
			}
		}
	}()
	return c
}

func (s *ProjectTestSuite) getOutputLinesFromEvents(events <-chan TmuxOutputEvent) <-chan string {
	c := make(chan string)
	go func() {
		defer close(c)
		var buffer string
		for event := range events {
			buffer = buffer + event.Data
			lines := strings.Split(buffer, "\\015\\012")
			for i, line := range lines {
				if i == len(lines)-1 {
					buffer = line
				} else {
					c <- line
				}
			}
		}
		lines := RemoveEmptyLines(strings.Split(buffer, "\\015\\012"))
		for _, line := range lines {
			c <- line
		}
	}()
	return c
}

func (s *ProjectEnsureStartedTestSuite) handleProjectStart(
	session TmuxSession,
	err error,
) TmuxSession {
	s.Expect(err).ToNot(HaveOccurred())
	for _, knownSession := range s.knownSessions {
		if knownSession.Id == session.Id {
			return session
		}
	}
	s.knownSessions = append(s.knownSessions, session)
	return session
}
