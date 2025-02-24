package main_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/stroiman/muxify"
)

var _ = Describe("Tmux", func() {
	var server TmuxServer

	BeforeEach(func() {
		server = MustCreateTestServer()
	})

	Describe("GetRunningSessions()", func() {
		BeforeEach(func() {
			server.SocketName = CreateRandomName()
		})

		It("Should return an empty slice when no server has been started", func() {
			sessions, err := server.GetRunningSessions()
			Expect(sessions).To(BeEmpty())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	var _ = Describe("Muxify", func() {
		BeforeEach(func() {
			output, err := server.Command("new-session", "-s", "test-session", "-d").Output()
			if err != nil {
				fmt.Println("Error!", string(output))
			}
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			sessions, err := server.GetRunningSessions()
			Expect(err).ToNot(HaveOccurred())
			session, ok := TmuxSessions(sessions).FindByName("test-session")
			if ok {
				Expect(server.KillSession(session)).To(Succeed())
			}
		})

		It("Returns a slice containing at least one element", func() {
			result, err := server.GetRunningSessions()
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(ContainElement(HaveField("Name", "test-session")))
		})

		It("Can be killed", func() {
			result, err := server.GetRunningSessions()
			Expect(err).ToNot(HaveOccurred())
			session, _ := TmuxSessions(result).FindByName("test-session")
			Expect(server.KillSession(session)).To(Succeed())
			result, err = server.GetRunningSessions()
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(ContainElement(HaveField("Name", "test-session")))
		})
	})
})
