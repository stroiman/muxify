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

type TmuxWindow struct {
	Id   string
	Name string
}

type TmuxSessions []TmuxSession

// sanitizeOutput removes new-line character codes. This is useful for parsing
// the standard out of a command that will normally be terminated with a
// new-line character.
func sanitizeOutput(output []byte) string {
	return strings.Trim(string(output), "\n")
}

func removeEmptyLines(lines []string) []string {
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

func getLines(output []byte) []string {
	return removeEmptyLines(strings.Split(string(output), "\n"))
}

type TmuxServer struct {
	ControlMode bool
	SocketName  string
	ConfigFile  string
}

func (s TmuxServer) Command(arg ...string) *exec.Cmd {

	c := make([]string, 0)
	if s.ControlMode {
		c = append(c, "-C")
	}
	if s.SocketName != "" {
		c = append(c, "-L", s.SocketName)
	}
	if s.ConfigFile != "" {
		c = append(c, "-f", s.ConfigFile)
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

func parseLines(output []byte) ([][2]string, error) {
	lines := getLines(output)
	result := make([][2]string, len(lines))
	for i, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			return [][2]string{}, fmt.Errorf("Bad result from tmux: %s", line)
		}
		result[i] = [2]string{parts[0], parts[1]}
	}
	return result, nil
}

func (s TmuxServer) GetRunningSessions() ([]TmuxSession, error) {
	stdOut, err := s.Command("start", ";", "list-sessions", "-F", "#{session_id}:#{session_name}").Output()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			fmt.Println("Exit error!!\n", string(exitErr.Stderr))
		}
		return nil, err
	}
	lines, err := parseLines(stdOut)
	result := make([]TmuxSession, len(lines))
	for i, line := range lines {
		result[i].Id = line[0]
		result[i].Name = line[1]
	}
	return result, err
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
	lines := getLines(output)
	panes = make([]TmuxPane, len(lines))
	for i, l := range lines {
		panes[i] = TmuxPane{Id: l}
	}
	return
}

func (s TmuxServer) GetWindowsForSession(session TmuxSession) (windows []TmuxWindow, err error) {
	var output []byte
	output, err = s.Command("list-windows", "-t", session.Id, "-F", "#{window_id}:#{window_name}").Output()
	if err != nil {
		return
	}
	lines, err := parseLines(output)
	windows = make([]TmuxWindow, len(lines))

	for i, line := range lines {
		windows[i].Id = line[0]
		windows[i].Name = line[1]
	}
	return
}

func (s TmuxServer) RenameWindow(windowId string, name string) error {
	return s.Command("rename-window", "-t", windowId, name).Run()
}
