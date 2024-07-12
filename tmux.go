package muxify

import (
	"fmt"
	"os/exec"
	"strings"
)

type TmuxSession struct {
	Id   string
	Name string
}

type TmuxSessions []TmuxSession

func StartSessionByName(name string) (TmuxSession, error) {

	out, err := exec.Command("tmux", "new-session", "-s", name, "-d", "-F", "#{session_id}", "-P").Output()
	if err != nil {
		return TmuxSession{}, err
	} else {
		return TmuxSession{
			Id:   strings.Trim(string(out), "\n"),
			Name: name,
		}, nil
	}
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
	if s.Id == "" {
		panic("Trying to kill a session with no id")
	}
	return exec.Command("tmux", "kill-session", "-t", s.Id).Run()
}

func (s TmuxSessions) FindByName(name string) (session TmuxSession, ok bool) {
	for _, session := range s {
		if session.Name == name {
			return session, true
		}
	}
	return TmuxSession{}, false
}
