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

type TmuxPane struct {
	Id string
}

type TmuxSessions []TmuxSession

func sanitizeOutput(output []byte) string {
	return strings.Trim(string(output), "\n")
}

func StartSessionByNameInDir(name string, dir string) (TmuxSession, error) {

	out, err := exec.Command("tmux", "new-session", "-s", name, "-d", "-F", "#{session_id}", "-P", "-c", dir).Output()
	if err != nil {
		return TmuxSession{}, err
	} else {
		return TmuxSession{
			Id:   sanitizeOutput(out),
			Name: name,
		}, nil
	}
}
func StartSessionByName(name string) (TmuxSession, error) {

	out, err := exec.Command("tmux", "new-session", "-s", name, "-d", "-F", "#{session_id}", "-P").Output()
	if err != nil {
		return TmuxSession{}, err
	} else {
		return TmuxSession{
			Id:   sanitizeOutput(out),
			Name: name,
		}, nil
	}
}

func GetRunningSessions() ([]TmuxSession, error) {
	stdOut, err := exec.Command("tmux", "list-sessions", "-F", "#{session_id}:#{session_name}").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(sanitizeOutput(stdOut), "\n")
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

func (s TmuxSession) GetPanes() (panes []TmuxPane, err error) {
	var output []byte
	output, err = exec.Command("tmux", "list-panes", "-t", s.Id, "-F", "#{pane_id}").Output()
	if err != nil {
		return
	}
	lines := strings.Split(sanitizeOutput(output), "\n")
	panes = make([]TmuxPane, len(lines))
	for i, l := range lines {
		panes[i] = TmuxPane{Id: l}
	}
	return
}
