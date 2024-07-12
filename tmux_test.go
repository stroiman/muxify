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
			exec.Command("tmux", "kill-session", "-t", "test-session").Run()
			// Expect(err).ToNot(HaveOccurred())
		})

		It("Returns a slice containing at least one element", func() {
			result, err := GetRunningSessions()
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(ContainElement(HaveField("Name", "test-session")))
		})

		It("Can be killed", func() {
			result, err := GetRunningSessions()
			Expect(err).ToNot(HaveOccurred())
			var session TmuxSession
			for _, s := range result {
				if s.Name == "test-session" {
					session = s
				}
			}
			Expect(session.Kill()).To(Succeed())
			result, err = GetRunningSessions()
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(ContainElement(HaveField("Name", "test-session")))
		})
	})

})
