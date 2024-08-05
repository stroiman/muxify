package muxify

import (
	"github.com/google/uuid"
)

type Command = string

type Task struct {
	Name     string // must be unique
	Commands Command
}

type Project struct {
	Name             string // must be unique
	WorkingDirectory string
	Windows          []Window
	Tasks            []Task
}

type WindowId = uuid.UUID

type Window struct {
	Id    WindowId
	Name  string
	Panes []Pane
}

type Pane struct {
	Name     string
	Commands []string
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
	server TmuxServer,
	window *TmuxWindow,
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
		var targetId string
		var existingPane = tmuxPanes.FindByTitle(pane.Name)
		if existingPane == nil {
			if i == 0 {
				err := server.Command("select-pane", "-t", window.Id, "-T", pane.Name).Run()
				if err != nil {
					return err
				}
				targetId = window.Id
			} else {
				output, err := server.Command("split-window", "-t", window.Id, "-P", "-F", "#{pane_id}").Output()
				if err != nil {
					return err
				}
				paneId := sanitizeOutput(output)
				err = server.Command("select-pane", "-t", paneId, "-T", pane.Name).Run()
				if err != nil {
					return err
				}
				targetId = paneId
			}
			// TODO: Don't create this type here
			tmuxPane := TmuxPane{TmuxTarget{server, targetId}, pane.Name}
			var err error = nil
			for _, command := range pane.Commands {
				if err == nil {
					err = tmuxPane.RunShellCommand(command)
				}
			}
		}
	}
	return nil
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
		ensureWindowHasPanes(server, existingWindow, configuredWindow)
	}
	return session, err
}
