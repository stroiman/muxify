package main

import (
	"errors"

	"github.com/google/uuid"
)

type Command = string

type Commands = []Command

type Task struct {
	Commands Commands
}

type Project struct {
	Name             string
	WorkingDirectory string `yaml:"working_dir,omitempty"`
	Windows          []Window
	Tasks            map[string]Task
}

type WindowId = uuid.UUID

type TaskId = string //

type Window struct {
	id     WindowId
	Name   string
	Panes  []TaskId
	Layout string
}

func NewWindow(name string, panes ...TaskId) Window {
	return Window{
		id:    uuid.New(),
		Name:  name,
		Panes: panes,
	}
}

func startSessionAndSetFirstWindowName(
	server TmuxServer,
	project Project,
) (session TmuxSession, err error) {
	if project.WorkingDirectory == "" {
		session, err = server.StartSessionByName(project.Name)
	} else {
		session, err = server.StartSessionByNameInDir(project.Name, project.WorkingDirectory)
	}
	if err == nil && len(project.Windows) > 0 {
		err = server.RenameWindow(session.Id, project.Windows[0].Name)
	}
	return
}

func ensureWindowHasPanes(
	window *TmuxWindow,
	project Project,
	configuredWindow Window,
) error {
	if window == nil {
		panic("Window must not be nil")
	}
	tmuxPanes, err := window.GetPanes()
	if err != nil {
		return err
	}
	for i, pane := range configuredWindow.Panes {
		var existingPane = tmuxPanes.FindByTitle(pane)
		if existingPane == nil {
			var (
				tmuxPane TmuxPane
			)
			if i == 0 {
				tmuxPane, err = window.GetFirstPane()
				if err == nil {
					tmuxPane, err = tmuxPane.Rename(pane)
				}
			} else {
				if configuredWindow.Layout == "horizontal" || configuredWindow.Layout == "" {
					tmuxPane, err = window.SplitHorizontal(pane, project.WorkingDirectory)
				} else if configuredWindow.Layout == "vertical" {
					tmuxPane, err = window.SplitVertical(pane, project.WorkingDirectory)
				} else {
					err = errors.New("Invalid window layout")
				}
			}
			if err != nil {
				return err
			}
			task := project.FindTaskById(pane)
			for _, command := range task.Commands {
				if err == nil {
					err = tmuxPane.RunShellCommand(command)
				}
			}
		}
	}
	return nil
}

func (p Project) FindTaskById(taskId string) *Task {
	task, ok := p.Tasks[taskId]
	if ok {
		return &task
	} else {
		return nil
	}
}

func (p Project) Validate() error {
	for _, window := range p.Windows {
		if window.id == uuid.Nil {
			return errors.New("Window is lacking an ID")
		}
	}
	return nil
}

// TmuxWindowMap maps a desired window configuration to an actual TMUX window
// that _may_ or _may not_ be properly configured
type TmuxWindowMap = map[WindowId]*TmuxWindow

func (p Project) ensureSession(server TmuxServer) (session TmuxSession, err error) {
	var ok bool
	sessions, err := server.GetRunningSessions()
	session, ok = TmuxSessions(sessions).FindByName(p.Name)
	if err == nil && !ok {
		// Set first window name - a session always has a window, and if the name
		// doesn't match a configured window, the tool will leave it be, as if it
		// was created by the user.
		session, err = startSessionAndSetFirstWindowName(server, p)
	}
	return
}

func (p Project) EnsureStarted(server TmuxServer) (session TmuxSession, err error) {
	session, err = p.ensureSession(server)
	if err != nil {
		return
	}
	tmuxWindows, err := session.GetWindows()
	windowMap := make(map[WindowId]*TmuxWindow)
	for _, window := range p.Windows {
		if tmuxWindow, ok := tmuxWindows.FindByName(window.Name); ok {
			windowMap[window.id] = &tmuxWindow
		}
	}

	for i, configuredWindow := range p.Windows {
		if err != nil {
			break
		}
		// Iterate through the list of _desired_ windows. For the first configured
		// window, we want it to be placed _before_ the first existing window. Any
		// subsequent window is then targeted to be created/moved to _after_ the
		// previously configured window. This we will assume is already correct, as
		// that was handled in the previous iteration.
		var windowTarget WindowTarget
		if i == 0 {
			windowTarget = BeforeWindow(&tmuxWindows[0])
		} else {
			windowTarget = AfterWindow(windowMap[p.Windows[i-1].id])
		}

		var existingWindow *TmuxWindow
		if existingWindow = windowMap[configuredWindow.id]; existingWindow == nil {
			existingWindow, err = server.CreateWindow(
				windowTarget,
				configuredWindow.Name,
				p.WorkingDirectory,
			)
			windowMap[configuredWindow.id] = existingWindow
		} else {
			err = server.MoveWindow(existingWindow, windowTarget)
		}
		if err == nil {
			err = ensureWindowHasPanes(existingWindow, p, configuredWindow)
		}
	}
	return
}

// func ensureWindow(
// 	server TmuxServer,
// 	configuredWindow Window,
// 	windowTarget WindowTarget,
// 	windowMap map[WindowId]*TmuxWindow,
// ) (existingWindow *TmuxWindow, err error) {
// 	existingWindow = windowMap[configuredWindow.id]
// 	if existingWindow == nil {
// 		existingWindow, err = server.CreateWindow(
// 			windowTarget,
// 			configuredWindow.Name,
// 			"",
// 		)
// 		windowMap[configuredWindow.id] = existingWindow
// 	} else {
// 		err = server.MoveWindow(existingWindow, windowTarget)
// 	}
// 	return
// }
