package muxify

type Project struct {
	Name             string
	WorkingDirectory string
	Windows          []Window
}

type Window struct {
	Name string
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
	windowMap := make(map[Window]*TmuxWindow)
	for _, window := range p.Windows {
		if tmuxWindow, ok := tmuxWindows.FindByName(window.Name); ok {
			windowMap[window] = &tmuxWindow
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
			windowTarget = AfterWindow(windowMap[p.Windows[i-1]])
		}

		existingWindow := windowMap[configuredWindow]
		if existingWindow == nil {
			windowMap[configuredWindow], err = server.CreateWindow(
				windowTarget,
				configuredWindow.Name,
			)
		} else {
			err = server.MoveWindow(existingWindow, windowTarget)
		}
	}
	return session, err
}
