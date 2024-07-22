package muxify_test

import (
	"os"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	. "github.com/stroiman/muxify"
)

var _ = Describe("Project", func() {
	var server TmuxServer

	BeforeEach(func() {
		server = MustCreateTestServer()
	})

	var getOutputEvents = func(lines <-chan string) <-chan TmuxOutputEvent {
		c := make(chan TmuxOutputEvent)
		r := regexp.MustCompile("^\\%output ([^ ]+) (.*)$")
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
				Name: "muxify-test-project",
			}
			session, err := proj.EnsureStarted(server)
			Expect(err).ToNot(HaveOccurred())
			defer server.KillSession(session)
			Expect(session).To(BeStarted())
		})

		It("Should create a session with one pane", func() {
			proj := Project{
				Name: "muxify-test-project",
			}
			session, err := proj.EnsureStarted(server)
			defer server.KillSession(session)
			Expect(err).ToNot(HaveOccurred())
			panes, err2 := server.GetPanesForSession(session)
			Expect(err2).ToNot(HaveOccurred())

			Expect(panes).To(HaveExactElements(HaveField("Id", MatchRegexp("^\\%\\d+$"))))
		})

		It("Should start in the correct working directory2", func() {
			proj := Project{
				Name:             "muxify-test-project",
				WorkingDirectory: dir,
			}
			session, err := proj.EnsureStarted(server)
			Expect(err).ToNot(HaveOccurred())
			defer server.KillSession(session)
			cm := MustStartControlMode(server, session)
			defer cm.MustClose()

			Expect(server.Command("send-keys", "-t", session.Id, "echo $PWD\n").Run()).To(Succeed())
			Eventually(getOutputEvents(GetLines(cm.stdout))).Should(Receive(HaveField("Data", Equal(dir))))
		})

		It("Should return the existing session if it has been started", func() {
			proj := Project{
				Name: "muxify-test-project",
			}
			s1, err1 := proj.EnsureStarted(server)
			defer server.KillSession(s1)
			Expect(err1).ToNot(HaveOccurred())
			s2, err2 := proj.EnsureStarted(server)
			Expect(err2).ToNot(HaveOccurred())
			Expect(s1.Id).To(Equal(s2.Id))
		})

		It("Should set the window name according to the specification", func() {
			proj := Project{
				Name: "muxify-test-project",
				Windows: []Window{
					Window{Name: "Window-1"},
				},
			}
			s1, err1 := proj.EnsureStarted(server)
			defer server.KillSession(s1)
			Expect(err1).ToNot(HaveOccurred())
			windows, err2 := server.GetWindowsForSession(s1)
			Expect(err2).ToNot(HaveOccurred())
			Expect(windows).To(ContainElements(HaveField("Name", "Window-1")))
		})

		It("Should support creating multiple windows", func() {
			Skip("TODO")
		})

		It("Should support window names with a colon", func() {
			Skip("TODO - or whatever character that the implementation detail chose as a separator")
		})

		It("Should create missing windows when the session was already running", func() {
			// E.g. you define two windows. One exists allready, the other doesn't.
			// The missing window should be created - at the correct index.
			Skip("TODO")
		})
	})
})

func BeStarted() types.GomegaMatcher {
	return HaveField("Id", MatchRegexp("^\\$\\d+"))
}
