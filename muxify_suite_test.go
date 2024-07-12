package muxify_test

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMuxify(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Muxify Suite")
}

type TmuxSession struct {
	Id   string
	Name string
}

func GetRunningSessions() ([]TmuxSession, error) {
	stdOut, err := exec.Command("tmux", "list-sessions", "-F", "#{session_id}:#{session_name}").Output()
	if err != nil {
		return nil, err
	}
	output := string(stdOut)
	lines := strings.Split(strings.Trim(output, "\n"), "\n")
	result := make([]TmuxSession, len(lines))
	for i, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("Bad result from tmux: %s", line)
		}
		result[i] = TmuxSession{
			Id:   parts[0],
			Name: parts[1],
		}
	}
	return result, nil
}

func (s TmuxSession) Kill() error {
	return exec.Command("tmux", "kill-session", "-t", s.Id).Run()
}

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
