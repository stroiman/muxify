package muxify_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	. "github.com/stroiman/muxify"
)

var _ = Describe("Project", func() {
	Describe("EnsureStarted", func() {
		It("Should start a new tmux session if not already started", func() {
			proj := Project{
				Name: "muxify-test-project",
			}
			session, err := proj.EnsureStarted()
			defer session.Kill()
			Expect(err).ToNot(HaveOccurred())
			Expect(session).To(BeStarted())
		})

		It("Should return the existing session if it has been started", func() {
			proj := Project{
				Name: "muxify-test-project",
			}
			s1, err1 := proj.EnsureStarted()
			defer s1.Kill()
			Expect(err1).ToNot(HaveOccurred())
			s2, err2 := proj.EnsureStarted()
			Expect(err2).ToNot(HaveOccurred())
			Expect(s1.Id).To(Equal(s2.Id))
		})
	})
})

func BeStarted() types.GomegaMatcher {
	return HaveField("Id", MatchRegexp("^\\$\\d+"))
}
