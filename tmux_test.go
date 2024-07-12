package muxify_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/stroiman/muxify"
)

var _ = Describe("Tmux", func() {
	var _ = Describe("Muxify", func() {
		BeforeEach(func() {
			err := exec.Command("tmux", "new-session", "-s", "test-session", "-d").Run()
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			sessions, err := GetRunningSessions()
			Expect(err).ToNot(HaveOccurred())
			session, ok := TmuxSessions(sessions).FindByName("test-session")
			if ok {
				Expect(session.Kill()).To(Succeed())
			}
		})

		It("Returns a slice containing at least one element", func() {
			result, err := GetRunningSessions()
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(ContainElement(HaveField("Name", "test-session")))
		})

		It("Can be killed", func() {
			result, err := GetRunningSessions()
			Expect(err).ToNot(HaveOccurred())
			session, _ := TmuxSessions(result).FindByName("test-session")
			Expect(session.Kill()).To(Succeed())
			result, err = GetRunningSessions()
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(ContainElement(HaveField("Name", "test-session")))
		})
	})

})
