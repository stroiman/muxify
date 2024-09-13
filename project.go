package muxify

import (
	"github.com/google/uuid"
)

type Command = string

type Commands = []Command

type Task struct {
	Commands Commands
}

type Project struct {
	Name             string // must be unique
	WorkingDirectory string
	Windows          []Window
	Tasks            map[string]Task
}

type WindowId = uuid.UUID

type TaskId = string //

type Window struct {
	Id    WindowId
	Name  string
	Panes []TaskId
}

func NewWindow(name string) Window {
	return Window{
		Id:   uuid.New(),
		Name: name,
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
				tmuxPane, err = window.Split(pane)
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

func (p Project) EnsureStarted(server TmuxServer) (TmuxSession, error) {
	var (
		session     TmuxSession
		tmuxWindows TmuxWindows
	)
	sessions, err := server.GetRunningSessions()
	if err != nil {
		return TmuxSession{}, err
	}
	session, ok := TmuxSessions(sessions).FindByName(p.Name)
	if !ok {
		session, err = startSessionAndSetFirstWindowName(server, p)
	}
	tmuxWindows, err = server.GetWindowsForSession(session)
	windowMap := make(map[WindowId]*TmuxWindow)
	for _, window := range p.Windows {
		if tmuxWindow, ok := tmuxWindows.FindByName(window.Name); ok {
			windowMap[window.Id] = &tmuxWindow
		}
	}
	for i, configuredWindow := range p.Windows {
		// Target position for window operations. Before or after an
		// existing window
		var windowTarget WindowTarget
		if i == 0 {
			// The first window we place _before_ the currently shown window
			windowTarget = BeforeWindow(&tmuxWindows[0])
		} else {
			// Other windows are placed _after_ the previously configured window
			// which we assume is already in the right place because it was
			// processed in the previous iteration.
			windowTarget = AfterWindow(windowMap[p.Windows[i-1].Id])
		}

		existingWindow := windowMap[configuredWindow.Id]
		if existingWindow == nil {
			existingWindow, err = server.CreateWindow(
				windowTarget,
				configuredWindow.Name,
			)
			windowMap[configuredWindow.Id] = existingWindow
		} else {
			err = server.MoveWindow(existingWindow, windowTarget)
		}
		ensureWindowHasPanes(existingWindow, p, configuredWindow)
	}
	return session, err
}
