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
		tmuxWindows []TmuxWindow
	)
	sessions, err := server.GetRunningSessions()
	if err != nil {
		return TmuxSession{}, err
	}
	existing, ok := TmuxSessions(sessions).FindByName(p.Name)
	if ok {
		return existing, nil
	}

	if p.WorkingDirectory == "" {
		session, err = server.StartSessionByName(p.Name)
	} else {
		session, err = server.StartSessionByNameInDir(p.Name, p.WorkingDirectory)
	}
	tmuxWindows, err = server.GetWindowsForSession(session)
	if err != nil {
		return TmuxSession{}, err
	}
	if len(p.Windows) > 0 {
		var previousWindow = tmuxWindows[0]
		err = server.RenameWindow(previousWindow.Id, p.Windows[0].Name)
		for i, window := range p.Windows {
			if i > 0 && err == nil {
				previousWindow, err = server.CreateWindowAfterTarget(session, previousWindow, window.Name)
			}
		}
	}
	return session, err
}
