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

type TmuxServer struct {
	ControlMode bool
}

func (s TmuxServer) Command(arg ...string) *exec.Cmd {

	c := make([]string, 0)
	if s.ControlMode {
		c = append(c, "-C")
	}
	c = append(c, arg...)

	return exec.Command("tmux", c...)
}

func (server TmuxServer) StartSession(name string, arg ...string) (TmuxSession, error) {
	c := append([]string{"new-session", "-F", "#{session_id}", "-P", "-d"}, arg...)
	out, err := server.Command(c...).Output()
	if err != nil {
		return TmuxSession{}, err
	} else {
		return TmuxSession{
			Id:   sanitizeOutput(out),
			Name: name,
		}, nil
	}
}

func (s TmuxServer) StartSessionByNameInDir(name string, dir string) (TmuxSession, error) {
	return s.StartSession(name, "-s", name, "-c", dir)
}

func (s TmuxServer) StartSessionByName(name string) (TmuxSession, error) {
	return s.StartSession(name, "-s", name)
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

func (s TmuxServer) KillSession(session TmuxSession) error {
	if session.Id == "" {
		panic("Trying to kill a session with no id")
	}
	return s.Command("kill-session", "-t", session.Id).Run()
}

func (s TmuxSessions) FindByName(name string) (session TmuxSession, ok bool) {
	for _, session := range s {
		if session.Name == name {
			return session, true
		}
	}
	return TmuxSession{}, false
}

func (s TmuxServer) GetPanesForSession(session TmuxSession) (panes []TmuxPane, err error) {
	var output []byte
	output, err = s.Command("list-panes", "-t", session.Id, "-F", "#{pane_id}").Output()
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
