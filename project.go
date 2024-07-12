package muxify

type Project struct {
	Name string
}

func (p Project) EnsureStarted() (TmuxSession, error) {
	sessions, err := GetRunningSessions()
	if err != nil {
		return TmuxSession{}, err
	}
	existing, ok := TmuxSessions(sessions).FindByName(p.Name)
	if ok {
		return existing, nil
	}
	return StartSessionByName(p.Name)

	// ...
}
