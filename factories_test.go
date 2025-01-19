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

type CreateProjectOption func(*TestProject)

func ProjectWorkingDir(dir string) CreateProjectOption {
	return func(p *TestProject) {
		p.WorkingDirectory = dir
	}
}

func CreateProject(options ...CreateProjectOption) *TestProject {
	return &TestProject{Project{
		Name: CreateRandomName(),
	}}
}

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

func (w *TestWindow) SetVerticalLayout() *TestWindow {
	w.Layout = "vertical"
	return w
}

func (w *TestWindow) SetHorizontalLayout() *TestWindow {
	w.Layout = "horizontal"
	return w
}

func (p *TestProject) ReplaceWindowNames(windowNames ...string) {
	windows := make([]Window, len(windowNames))
	for i, name := range windowNames {
		windows[i] = NewWindow(name)
	}
	p.Project.Windows = windows
}

type CreatePaneOption struct {
	UpdateTask func(task *Task)
}

func TaskWorkingDir(dir string) CreatePaneOption {
	return CreatePaneOption{
		UpdateTask: func(task *Task) { task.WorkingDirectory = dir },
	}
}

func TaskCommands(commands ...string) CreatePaneOption {
	return CreatePaneOption{
		UpdateTask: func(task *Task) { task.Commands = Commands(commands) },
	}
}

func (s *TestProject) CreatePane(paneName string, options ...CreatePaneOption) TaskId {
	task := Task{}
	if s.Tasks == nil {
		s.Tasks = make(map[string]Task)
	}
	for _, o := range options {
		if o.UpdateTask != nil {
			o.UpdateTask(&task)
		}
	}
	s.Tasks[paneName] = task
	pane := paneName
	return pane
}

func (s *TestProject) CreatePaneWithCommands(paneName string, commands ...string) TaskId {
	// return s.CreatePane(paneName, TaskCommands(commands...));
	task := Task{Commands: Commands(commands)}
	if s.Tasks == nil {
		s.Tasks = make(map[string]Task)
	}
	s.Tasks[paneName] = task
	pane := paneName
	return pane
}
