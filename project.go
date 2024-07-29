package muxify

type Project struct {
	Name             string
	WorkingDirectory string
	Windows          []Window
}

type Window struct {
	Name string
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
	existing, ok := TmuxSessions(sessions).FindByName(p.Name)
	if !ok {
		if p.WorkingDirectory == "" {
			existing, err = server.StartSessionByName(p.Name)
		} else {
			existing, err = server.StartSessionByNameInDir(p.Name, p.WorkingDirectory)
		}
		if err == nil && len(p.Windows) > 0 {
			err = server.RenameWindow(existing.Id, p.Windows[0].Name)
		}
	}
	tmuxWindows, err = server.GetWindowsForSession(session)
	windowMap := make(map[Window]*TmuxWindow)
	// Map desired windows to actual running windows
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
	return existing, err
}
