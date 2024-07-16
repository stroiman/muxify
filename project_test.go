package muxify_test

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	. "github.com/stroiman/muxify"
)

var _ = Describe("Project", func() {
	Describe("EnsureStarted", func() {
		var dir string
		var server TmuxServer

		BeforeEach(func() {
			var err error
			dir, err = os.MkdirTemp("", "muxify-test-")
			server = TmuxServer{}
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
			defer session.Kill()
			Expect(session).To(BeStarted())
		})

		It("Should create a session with one pane", func() {
			proj := Project{
				Name: "muxify-test-project",
			}
			session, err := proj.EnsureStarted(server)
			defer session.Kill()
			Expect(err).ToNot(HaveOccurred())
			panes, err2 := server.GetPanesForSession(session)
			Expect(err2).ToNot(HaveOccurred())

			Expect(panes).To(HaveExactElements(HaveField("Id", MatchRegexp("^\\%\\d+$"))))
		})

		It("Should start in the correct working directory", func() {
			proj := Project{
				Name:             "muxify-test-project",
				WorkingDirectory: dir,
			}
			session, err := proj.EnsureStarted(server)
			Expect(err).ToNot(HaveOccurred())
			defer session.Kill()

			// Verify that it was started in the right working directory.
			// How? E.g. run `echo $PWD` in the shell and read the output

			time.Sleep(400 * time.Millisecond)
			Expect(exec.Command("tmux", "send-keys", "-t", session.Id, "echo $PWD\n").Run()).To(Succeed())

			time.Sleep(500 * time.Millisecond)
			output, err2 := exec.Command("tmux", "capture-pane", "-t", session.Id, "-p", "-E", "10").Output()
			expected := fmt.Sprintf("(?m:^%s$)", dir)
			Expect(string(output)).To(MatchRegexp(expected))

			Expect(err2).ToNot(HaveOccurred())
		})

		It("Should return the existing session if it has been started", func() {
			proj := Project{
				Name: "muxify-test-project",
			}
			s1, err1 := proj.EnsureStarted(server)
			defer s1.Kill()
			Expect(err1).ToNot(HaveOccurred())
			s2, err2 := proj.EnsureStarted(server)
			Expect(err2).ToNot(HaveOccurred())
			Expect(s1.Id).To(Equal(s2.Id))
		})
	})

	// The project defines one layout with one pane running one command in one
	// specific working directory

})

func BeStarted() types.GomegaMatcher {
	return HaveField("Id", MatchRegexp("^\\$\\d+"))
}
