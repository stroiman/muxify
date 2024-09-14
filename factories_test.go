package main_test

import (
	. "github.com/stroiman/muxify"
)

type TestProject struct{ Project }

func CreateProjectWithWindows(windows ...Window) *TestProject {
	return &TestProject{
		Project{
			Name:    CreateRandomName(),
			Windows: windows,
		},
	}
}

// Just an alias to hide the fact that creating an empty project is just
// the same as not supplying any variables arguments
var CreateProject = CreateProjectWithWindows

func CreateProjectWithWindowNames(windowNames ...string) *TestProject {
	windows := make([]Window, len(windowNames))
	for i, name := range windowNames {
		windows[i] = NewWindow(name)
	}
	return CreateProjectWithWindows(windows...)
}

type TestWindow struct {
	*TestProject
	*Window
}

func (p *TestProject) AppendNamedWindow(windowName string) *TestWindow {
	p.Project.Windows = append(p.Project.Windows, NewWindow(windowName))
	windows := p.Project.Windows
	return &TestWindow{
		p,
		&windows[len(windows)-1],
	}
}

func (w *TestWindow) AppendPane(pane TaskId) *TestWindow {
	w.Panes = append(w.Panes, pane)
	return w
}

func (p *TestProject) ReplaceWindowNames(windowNames ...string) {
	windows := make([]Window, len(windowNames))
	for i, name := range windowNames {
		windows[i] = NewWindow(name)
	}
	p.Project.Windows = windows
}

func (s *TestProject) CreatePaneWithCommands(paneName string, commands ...string) TaskId {
	task := Task{Commands(commands)}
	if s.Tasks == nil {
		s.Tasks = make(map[string]Task)
	}
	s.Tasks[paneName] = task
	pane := paneName
	return pane
}

func CreateWindowWithPanes(windowName string, panes ...TaskId) Window {
	window := NewWindow(windowName)
	window.Panes = panes
	return window
}
