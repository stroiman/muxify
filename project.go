package muxify

type Project struct {
	Name             string
	WorkingDirectory string
}

func (p Project) EnsureStarted(server TmuxServer) (TmuxSession, error) {
	sessions, err := server.GetRunningSessions()
	if err != nil {
		return TmuxSession{}, err
	}
	existing, ok := TmuxSessions(sessions).FindByName(p.Name)
	if ok {
		return existing, nil
	}
	if p.WorkingDirectory == "" {
		return server.StartSessionByName(p.Name)
	} else {
		return server.StartSessionByNameInDir(p.Name, p.WorkingDirectory)
	}

	// ...
}
