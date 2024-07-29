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
	// Ensure desired windows are in right position
	for i, window := range p.Windows {
		tmuxWindow := windowMap[window]
		var target *TmuxWindow
		var before bool
		if i == 0 {
			target = &tmuxWindows[0]
			before = true
		} else {
			target = windowMap[p.Windows[i-1]]
			before = false
		}
		if tmuxWindow == nil {
			windowMap[window], err = server.CreateWindowBeforeOrAfterTarget(
				target,
				window.Name,
				before,
			)
		} else {
			err = server.MoveWindowBeforeOrAfterTarget(tmuxWindow, target, before)
		}
	}
	return session, err
}
