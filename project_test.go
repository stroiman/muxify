package muxify_test

import (
	"fmt"
	"os"
	"regexp"

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
			// Ignore the error that occurs if server is not running.
			server.Kill()
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
		r := regexp.MustCompile("^\\%output ([^ ]+) (.*)-END")
		go func() {
			for line := range lines {
				m := r.FindStringSubmatch(line)
				if m != nil {
					event := TmuxOutputEvent{
						PaneId: m[1],
						Data:   removeControlCharacters(m[2]),
					}
					c <- event
				}
			}
			close(c)
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
			proj := Project{
				Name: CreateRandomName(),
			}
			session := handleProjectStart(proj.EnsureStarted(server))
			Expect(session).To(BeStarted())
		})

		It("Should create a session with one pane", func() {
			proj := Project{
				Name: CreateRandomProjectName(),
			}
			session := handleProjectStart(proj.EnsureStarted(server))
			panes, err2 := server.GetPanesForSession(session)
			Expect(err2).ToNot(HaveOccurred())

			Expect(panes).To(HaveExactElements(HaveField("Id", MatchRegexp("^\\%\\d+$"))))
		})

		It("Should start in the correct working directory", Focus, func() {
			proj := Project{
				Name:             CreateRandomProjectName(),
				WorkingDirectory: dir,
			}
			session := handleProjectStart(proj.EnsureStarted(server))
			cm := MustStartControlMode(server, session)
			defer cm.MustClose()
			success := false

			defer func() {
				if !success {
					fmt.Println("FAILURE")
					output, err := server.Command("capture-pane", "-p", "-t", session.Id).Output()
					if err != nil {
						fmt.Println("Error getting target pane", err)
					} else {
						fmt.Println("TARGET PANE OUTPUT")
						fmt.Println(string(output))
					}
				}
			}()

			fmt.Println("Executing verificaion")

			Expect(
				server.Command("send-keys", "-t", session.Id, "echo $PWD-END\n").Run(),
			).To(Succeed())
			Eventually(
				getOutputEvents(GetLines(cm.stdout)),
			).Should(Receive(HaveField("Data", Equal(dir))))
			success = true
		})

		It("Should return the existing session if it has been started", func() {
			proj := Project{
				Name: CreateRandomProjectName(),
			}
			s1 := handleProjectStart(proj.EnsureStarted(server))
			s2 := handleProjectStart(proj.EnsureStarted(server))
			Expect(s1.Id).To(Equal(s2.Id))
		})

		It("Should set the window name according to the specification", func() {
			proj := Project{
				Name: CreateRandomProjectName(),
				Windows: []Window{
					{Name: "Window-1"},
				},
			}
			s1 := handleProjectStart(proj.EnsureStarted(server))
			windows, err2 := server.GetWindowsForSession(s1)
			Expect(err2).ToNot(HaveOccurred())
			Expect(windows).To(HaveExactElements(HaveField("Name", "Window-1")))
		})

		It("Should support creating multiple windows", func() {
			proj := Project{
				Name: CreateRandomProjectName(),
				Windows: []Window{
					{Name: "Window-1"},
					{Name: "Window-2"},
					{Name: "Window-3"},
				},
			}
			session := handleProjectStart(proj.EnsureStarted(server))
			windows, err2 := server.GetWindowsForSession(session)
			Expect(err2).ToNot(HaveOccurred())
			Expect(windows).To(HaveExactElements(
				HaveField("Name", "Window-1"),
				HaveField("Name", "Window-2"),
				HaveField("Name", "Window-3"),
			))
		})

		It("Should create missing windows when the session was already running", func() {
			proj := Project{
				Name: CreateRandomProjectName(),
				Windows: []Window{
					{Name: "Window-1"},
					{Name: "Window-2"},
				},
			}
			session := handleProjectStart(proj.EnsureStarted(server))
			proj.Windows = append(proj.Windows, Window{Name: "Window-3"})
			handleProjectStart(proj.EnsureStarted(server))
			windows, err2 := server.GetWindowsForSession(session)
			Expect(err2).ToNot(HaveOccurred())
			Expect(windows).To(HaveExactElements(
				HaveField("Name", "Window-1"),
				HaveField("Name", "Window-2"),
				HaveField("Name", "Window-3"),
			))
		})

		It("Should create missing windows when the session was already running", func() {
			proj := Project{
				Name: CreateRandomProjectName(),
				Windows: []Window{
					{Name: "Window-2"},
				},
			}
			session := handleProjectStart(proj.EnsureStarted(server))

			proj.Windows = []Window{
				{Name: "Window-1"},
				{Name: "Window-2"},
			}
			handleProjectStart(proj.EnsureStarted(server))
			windows, err2 := server.GetWindowsForSession(session)
			Expect(err2).ToNot(HaveOccurred())
			Expect(windows).To(HaveExactElements(
				HaveField("Name", "Window-1"),
				HaveField("Name", "Window-2"),
			))
		})

		It("Should support custom working folder and environment for each window", func() {
			Skip("TODO")
		})
	})
})

func BeStarted() types.GomegaMatcher {
	return HaveField("Id", MatchRegexp("^\\$\\d+"))
}
