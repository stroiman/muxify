package muxify_test

import (
	"os"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	. "github.com/stroiman/muxify"
)

var _ = Describe("Project", Ordered, func() {
	var server TmuxServer
	var knownSessions []TmuxSession

	handleProjectStart := func(session TmuxSession, err error) TmuxSession {
		Expect(err).ToNot(HaveOccurred())
		for _, knownSession := range knownSessions {
			if knownSession.Id == session.Id {
				return session
			}
		}
		knownSessions = append(knownSessions, session)
		return session
	}

	BeforeAll(func() {
		server = MustCreateTestServer()
		DeferCleanup(func() {
			// kill-server will return an error if no server is running. So if it
			// does not return an error, some test did not clean up correctly
			Expect(server.KillServer()).ToNot(Succeed())
		})
	})

	BeforeEach(func() {
		knownSessions = nil
	})

	AfterEach(func() {
		for _, knownSession := range knownSessions {
			server.KillSession(knownSession)
		}
	})

	var getOutputEvents = func(lines <-chan string) <-chan TmuxOutputEvent {
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

	var getOutputLinesFromEvents = func(events <-chan TmuxOutputEvent) <-chan string {
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

	Describe("EnsureStarted", func() {
		var dir string

		BeforeEach(func() {
			var err error
			dir, err = os.MkdirTemp("", "muxify-test-")
			if err != nil {
				panic(err)
			}
		})

		AfterEach(func() {
			err := os.Remove(dir)
			if err != nil {
				panic(err)
			}
		})

		It("Should start a new tmux session if not already started", func() {
			proj := CreateProject()
			session := handleProjectStart(proj.EnsureStarted(server))
			Expect(session).To(BeStarted())
		})

		It("Should create a session with one pane", func() {
			proj := CreateProject()
			session := handleProjectStart(proj.EnsureStarted(server))

			Expect(
				session.GetPanes(),
			).To(HaveExactElements(HaveField("Id", MatchRegexp("^\\%\\d+$"))))
		})

		It("Should start in the correct working directory", func() {
			proj := CreateProject()
			proj.WorkingDirectory = dir
			session := handleProjectStart(proj.EnsureStarted(server))
			cm := MustStartControlMode(server, session)
			defer cm.MustClose()

			Expect(
				session.RunShellCommand("echo $PWD"),
			).To(Succeed())
			Eventually(
				getOutputLinesFromEvents(getOutputEvents(GetLines(cm.stdout))),
			).Should(Receive(Equal(dir)))
		})

		It("Should return the existing session if it has been started", func() {
			proj := CreateProject()
			s1 := handleProjectStart(proj.EnsureStarted(server))
			s2 := handleProjectStart(proj.EnsureStarted(server))
			Expect(s1.Id).To(Equal(s2.Id))
		})

		It("Should set the window name according to the specification", func() {
			proj := CreateProjectWithWindowNames("Window-1")
			s1 := handleProjectStart(proj.EnsureStarted(server))
			Expect(
				server.GetWindowsForSession(s1),
			).To(HaveExactElements(HaveField("Name", "Window-1")))
		})

		It("Should support creating multiple windows", func() {
			proj := CreateProjectWithWindowNames(
				"Window-1",
				"Window-2",
				"Window-3",
			)
			session := handleProjectStart(proj.EnsureStarted(server))
			Expect(server.GetWindowsForSession(session)).To(HaveExactElements(
				HaveField("Name", "Window-1"),
				HaveField("Name", "Window-2"),
				HaveField("Name", "Window-3"),
			))
		})

		It("Should create missing windows when the session was already running", func() {
			proj := CreateProjectWithWindowNames(
				"Window-1",
				"Window-2",
			)
			session := handleProjectStart(proj.EnsureStarted(server))
			proj.AppendNamedWindow("Window-3")
			handleProjectStart(proj.EnsureStarted(server))
			Expect(server.GetWindowsForSession(session)).To(HaveExactElements(
				HaveField("Name", "Window-1"),
				HaveField("Name", "Window-2"),
				HaveField("Name", "Window-3"),
			))
		})

		It("Should create missing windows and rearrange out-of-order windows", func() {
			proj := CreateProjectWithWindowNames("Window-4", "Window-1", "Window-3")
			session := handleProjectStart(proj.EnsureStarted(server))
			proj.ReplaceWindowNames("Window-1", "Window-2", "Window-3", "Window-4")
			handleProjectStart(proj.EnsureStarted(server))
			Expect(server.GetWindowsForSession(session)).To(HaveExactElements(
				HaveField("Name", "Window-1"),
				HaveField("Name", "Window-2"),
				HaveField("Name", "Window-3"),
				HaveField("Name", "Window-4"),
			))
		})

		It("Should create missing windows when the session was already running", func() {
			proj := CreateProjectWithWindowNames("Window-2")
			session := handleProjectStart(proj.EnsureStarted(server))
			proj.ReplaceWindowNames("Window-1", "Window-2")
			handleProjectStart(proj.EnsureStarted(server))
			Expect(server.GetWindowsForSession(session)).To(HaveExactElements(
				HaveField("Name", "Window-1"),
				HaveField("Name", "Window-2"),
			))
		})

		It("Should create the panes with the correct names", func() {
			proj := CreateProject()
			proj.AppendNamedWindow("Window-1").
				AppendPane(proj.CreatePaneWithCommands("Pane-1")).
				AppendPane(proj.CreatePaneWithCommands("Pane-2"))
			proj.AppendNamedWindow("Window-2").
				AppendPane(proj.CreatePaneWithCommands("Pane-3")).
				AppendPane(proj.CreatePaneWithCommands("Pane-4"))
			handleProjectStart(proj.EnsureStarted(server))
			expected := []T{
				{"Window-1", "Pane-1"},
				{"Window-1", "Pane-2"},
				{"Window-2", "Pane-3"},
				{"Window-2", "Pane-4"},
			}
			Expect(server.GetWindowAndPaneNames()).To(HaveExactElements(expected))
		})

		It("Should not add more panes when re-launching", func() {
			proj := CreateProject()
			proj.AppendNamedWindow("Window-1").
				AppendPane(proj.CreatePaneWithCommands("Pane-1")).
				AppendPane(proj.CreatePaneWithCommands("Pane-2"))
			proj.AppendNamedWindow("Window-2").
				AppendPane(proj.CreatePaneWithCommands("Pane-3")).
				AppendPane(proj.CreatePaneWithCommands("Pane-4"))
			handleProjectStart(proj.EnsureStarted(server))
			handleProjectStart(proj.EnsureStarted(server))

			expected := []T{
				{"Window-1", "Pane-1"},
				{"Window-1", "Pane-2"},
				{"Window-2", "Pane-3"},
				{"Window-2", "Pane-4"},
			}
			Expect(server.GetWindowAndPaneNames()).To(HaveExactElements(expected))
		})

		It("Should execute the commands defined in the pane configuration", func() {
			proj := CreateProject()
			proj.AppendNamedWindow("Window-1").
				AppendPane(proj.CreatePaneWithCommands("Pane-1", "echo \"Foo\"")).
				AppendPane(proj.CreatePaneWithCommands("Pane-2", "echo \"Bar\""))
			session := handleProjectStart(proj.EnsureStarted(server))
			cm := MustStartControlMode(server, session)
			defer cm.MustClose()
			outputEvents := getOutputLinesFromEvents(getOutputEvents(GetLines(cm.stdout)))
			panes, err := session.GetPanes()
			Expect(err).ToNot(HaveOccurred())
			panes[0].MustRunShellCommand("echo \"DONE 1\"")
			Eventually(outputEvents).Should(Receive(Equal("DONE 1")))
			panes[1].MustRunShellCommand("echo \"DONE 2\"")
			Eventually(outputEvents).Should(Receive(Equal("DONE 2")))
			output1 := server.Command("capture-pane", "-p", "-t", panes[0].Id).MustOutput()
			output2 := server.Command("capture-pane", "-p", "-t", panes[1].Id).MustOutput()
			Expect(output1).To(MatchRegexp("(?m:^Foo$)"))
			Expect(output2).To(MatchRegexp("(?m:^Bar$)"))
		})

		It("Should not run the first pane's command twice on second startup", func() {
			proj := CreateProject()
			proj.AppendNamedWindow("Window-1").
				AppendPane(proj.CreatePaneWithCommands("Pane-1", "echo \"Foo\"")).
				AppendPane(proj.CreatePaneWithCommands("Pane-2", "echo \"Bar\""))
			session := handleProjectStart(proj.EnsureStarted(server))
			cm := MustStartControlMode(server, session)
			defer cm.MustClose()
			outputEvents := getOutputLinesFromEvents(getOutputEvents(GetLines(cm.stdout)))
			panes, err := session.GetPanes()
			Expect(err).ToNot(HaveOccurred())

			// Wait for the commands to have executed
			panes[0].MustRunShellCommand("echo \"DONE 1\"")
			Eventually(outputEvents).Should(Receive(Equal("DONE 1")))

			// Start this again
			handleProjectStart(proj.EnsureStarted(server))

			// Wait for all commands to have executed
			panes[0].MustRunShellCommand("echo \"DONE 1\"")
			Eventually(outputEvents).Should(Receive(Equal("DONE 1")))

			panes, err = session.GetPanes()
			Expect(err).ToNot(HaveOccurred())

			output1 := server.Command("capture-pane", "-p", "-t", panes[0].Id).MustOutput()
			output2 := server.Command("capture-pane", "-p", "-t", panes[1].Id).MustOutput()

			var exp *regexp.Regexp = regexp.MustCompile(`(?m:^(?:Foo|Bar))`)
			Expect(exp.FindAllString(string(output1), -1)).To(Equal([]string{"Foo"}))
			Expect(exp.FindAllString(string(output2), -1)).To(Equal([]string{"Bar"}))
		})
	})
})

func BeStarted() types.GomegaMatcher {
	return HaveField("Id", MatchRegexp("^\\$\\d+"))
}
