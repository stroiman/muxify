package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type TmuxServer struct {
	ControlMode bool
	SocketName  string
	ConfigFile  string
}

type TmuxTarget struct {
	TmuxServer
	Id string
}

type TmuxSession struct {
	TmuxTarget
	Name string
}
type PaneLayout struct {
	Top    int
	Bottom int
	Left   int
	Right  int
}

type TmuxPane struct {
	TmuxTarget
	Title  string
	Layout PaneLayout
}

type TmuxWindow struct {
	TmuxTarget
	Name string
}

type TmuxWindows []TmuxWindow

func (ws TmuxWindows) FindByName(name string) (TmuxWindow, bool) {
	for _, window := range ws {
		if window.Name == name {
			return window, true
		}
	}
	return TmuxWindow{}, false
}

type TmuxSessions []TmuxSession

// sanitizeOutput removes new-line character codes. This is useful for parsing
// the standard out of a command that will normally be terminated with a
// new-line character.
func sanitizeOutput(output []byte) string {
	return strings.Trim(string(output), "\n")
}

func RemoveEmptyLines(lines []string) []string {
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

func getLines(output []byte) []string {
	return RemoveEmptyLines(strings.Split(string(output), "\n"))
}

func (s TmuxServer) KillServer() error {
	return s.Command("kill-server").Run()
}

type CmdExt struct {
	*exec.Cmd
}

func (c CmdExt) MustOutput() []byte {
	o, err := c.Output()
	if err != nil {
		panic(err)
	}
	return o
}

func (s TmuxServer) Command(arg ...string) CmdExt {
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
	cmd := exec.Command("tmux", c...)
	return CmdExt{cmd}
}

func (server TmuxServer) StartSession(name string, arg ...string) (TmuxSession, error) {
	c := append([]string{"new-session", "-F", "#{session_id}", "-P", "-d"}, arg...)
	out, err := server.Command(c...).Output()
	if err != nil {
		return TmuxSession{}, err
	} else {
		return TmuxSession{
			TmuxTarget{
				server,
				sanitizeOutput(out),
			},
			name,
		}, nil
	}
}

func (s TmuxServer) StartSessionByNameInDir(name string, dir string) (TmuxSession, error) {
	return s.StartSession(name, "-s", name, "-c", dir)
}

func (s TmuxServer) StartSessionByName(name string) (TmuxSession, error) {
	return s.StartSession(name, "-s", name)
}

var lineParser *regexp.Regexp = regexp.MustCompile(`^"([^"]+)":"([^"]+)"$`)

func parseLinesQuoted(output []byte) ([][2]string, error) {
	lines := getLines(output)
	result := make([][2]string, len(lines))
	for i, line := range lines {
		submatch := lineParser.FindStringSubmatch(line)
		if submatch == nil {
			return [][2]string{}, fmt.Errorf("Bad result from tmux: %s", line)
		}

		result[i] = [2]string{submatch[1], submatch[2]}
	}
	return result, nil
}

var paneLineParser *regexp.Regexp = regexp.MustCompile(
	`^"([^"]+)":"([^"]+)":(\d+),(\d+),(\d+),(\d+)$`,
)

type PaneData struct {
	ids    [2]string
	layout PaneLayout
}

func parsePaneOutput(output []byte) ([]PaneData, error) {
	lines := getLines(output)
	result := make([]PaneData, len(lines))
	for i, line := range lines {
		submatch := paneLineParser.FindStringSubmatch(line)
		if submatch == nil {
			return []PaneData{}, fmt.Errorf("Bad result from tmux: %s", line)
		}

		top, _ := strconv.Atoi(submatch[3])
		bottom, _ := strconv.Atoi(submatch[4])
		left, _ := strconv.Atoi(submatch[5])
		right, _ := strconv.Atoi(submatch[5])
		layout := PaneLayout{
			top, bottom, left, right,
		}
		ids := [2]string{submatch[1], submatch[2]}
		result[i] = PaneData{ids, layout}
	}
	return result, nil
}

func (s TmuxServer) GetRunningSessions() ([]TmuxSession, error) {
	stdOut, err := s.Command("start", ";", "list-sessions", "-F", `"#{session_id}":"#{session_name}"`).
		Output()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			fmt.Println("Exit error!!\n", string(exitErr.Stderr))
		}
		return nil, err
	}
	lines, err := parseLinesQuoted(stdOut)
	result := make([]TmuxSession, len(lines))
	for i, line := range lines {
		result[i] = TmuxSession{
			TmuxTarget{
				s,
				line[0],
			},
			line[1],
		}
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

type TmuxPanes []TmuxPane

func (p TmuxPanes) FindByTitle(title string) *TmuxPane {
	for _, pane := range p {
		if pane.Title == title {
			return &pane
		}
	}
	return nil
}

func (s TmuxTarget) GetPanes() (panes TmuxPanes, err error) {
	var output []byte
	output, err = s.Command("list-panes", "-t", s.Id, "-F", `"#{pane_id}":"#{pane_title}":#{pane_top},#{pane_bottom},#{pane_left},#{pane_right}`).
		Output()
	if err != nil {
		return
	}
	lines, err := parsePaneOutput(output)
	panes = make([]TmuxPane, len(lines))
	for i, data := range lines {
		l := data.ids
		panes[i] = TmuxPane{
			TmuxTarget{
				s.TmuxServer,
				l[0],
			},
			l[1],
			data.layout,
		}
	}
	return
}

func (s TmuxTarget) MustGetPanes() TmuxPanes {
	panes, err := s.GetPanes()
	must(err)
	return panes
}

func (s TmuxServer) GetWindowsForSession(session TmuxSession) (windows TmuxWindows, err error) {
	var output []byte
	output, err = s.Command("list-windows", "-t", session.Id, "-F", `"#{window_id}":"#{window_name}"`).
		Output()
	if err != nil {
		return
	}
	lines, err := parseLinesQuoted(output)
	windows = make([]TmuxWindow, len(lines))
	for i, line := range lines {
		windows[i] = TmuxWindow{TmuxTarget{s, line[0]}, line[1]}
	}
	return
}

func (s TmuxSession) GetWindows() (TmuxWindows, error) {
	return s.GetWindowsForSession(s)
}

func (s TmuxSession) MustGetWindows() TmuxWindows {
	windows, err := s.GetWindows()
	must(err)
	return windows
}

func (s TmuxServer) RenameWindow(windowId string, name string) error {
	return s.Command("rename-window", "-t", windowId, name).Run()
}

// WindowTarget represents how the position of a new TMUX window can be passed
// to TMUX itself, which can be either _before_ or _after_ an existing window.
type WindowTarget struct {
	target *TmuxWindow
	before bool
}

func BeforeWindow(target *TmuxWindow) WindowTarget {
	return WindowTarget{
		target: target,
		before: true,
	}
}

func AfterWindow(target *TmuxWindow) WindowTarget {
	return WindowTarget{
		target: target,
		before: false,
	}
}

func (t WindowTarget) createArgs() []string {
	if t.before {
		return []string{"-b", "-t", t.target.Id}
	} else {
		return []string{"-a", "-t", t.target.Id}
	}
}

func (s TmuxServer) CreateWindow(
	target WindowTarget,
	name string,
	workingDir string,
) (*TmuxWindow, error) {
	args := []string{"new-window", "-n", name, "-F", "#{window_id}", "-P"}
	args = append(args, target.createArgs()...)
	if workingDir != "" {
		args = append(args, "-c", workingDir)
	}
	output, err := s.Command(args...).Output()
	window := TmuxWindow{
		TmuxTarget{
			s,
			sanitizeOutput(output),
		},
		name,
	}
	return &window, err
}

func (s TmuxServer) MoveWindow(
	window *TmuxWindow,
	target WindowTarget) error {
	args := []string{"move-window", "-s", window.Id}
	args = append(args, target.createArgs()...)
	return s.Command(args...).Run()
}

type T struct {
	WindowName string
	PaneTitle  string
}

func (s TmuxServer) GetWindowAndPaneNames() ([]T, error) {
	output, err := s.Command(
		"list-panes",
		"-a",
		"-F",
		`#{window_name}:#{pane_title}`,
	).Output()
	if err != nil {
		return nil, err
	}
	lines := getLines(output)
	result := make([]T, len(lines))
	for i, line := range lines {
		parts := strings.Split(line, ":")
		result[i] = T{parts[0], parts[1]}
	}
	return result, nil
}

func (s TmuxTarget) RunShellCommand(shellCommand string) error {
	return s.Command("send-keys", "-t", s.Id, shellCommand+"\n").Run()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func (s TmuxTarget) MustRunShellCommand(shellCommand string) {
	must(s.RunShellCommand(shellCommand))
}

func (s TmuxTarget) GetFirstPane() (pane TmuxPane, err error) {
	var panes []TmuxPane
	panes, err = s.GetPanes()
	if len(panes) > 0 {
		pane = panes[0]
	}
	return
}

func (p TmuxPane) Rename(name string) (TmuxPane, error) {
	err := p.Command("select-pane", "-t", p.Id, "-T", name).Run()
	if err == nil {
		p.Title = name
	}
	return p, err
}

func (w TmuxWindow) SplitHorizontal(name string, workingDir string) (TmuxPane, error) {
	args := []string{"split-window", "-h", "-t", w.Id, "-P", "-F", "#{pane_id}"}
	if workingDir != "" {
		args = append(args, "-c", workingDir)
	}
	output, err := w.Command(args...).
		Output()
	paneId := sanitizeOutput(output)
	pane := TmuxPane{TmuxTarget{w.TmuxServer, paneId}, "", PaneLayout{}}
	if err == nil {
		pane, err = pane.Rename(name)
	}
	return pane, err
}

func (w TmuxWindow) SplitVertical(name string, workingDir string) (TmuxPane, error) {
	args := []string{"split-window", "-v", "-t", w.Id, "-P", "-F", "#{pane_id}"}
	if workingDir != "" {
		args = append(args, "-c", workingDir)
	}
	output, err := w.Command(args...).
		Output()
	paneId := sanitizeOutput(output)
	pane := TmuxPane{TmuxTarget{w.TmuxServer, paneId}, "", PaneLayout{}}
	if err == nil {
		pane, err = pane.Rename(name)
	}
	return pane, err
}
