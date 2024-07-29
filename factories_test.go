package muxify_test

import (
	. "github.com/stroiman/muxify"
)

type TestProject struct{ Project }

func CreateProject() TestProject {
	return TestProject{
		Project{
			Name: CreateRandomName(),
		},
	}
}

func CreateProjectWithWindowNames(windowNames ...string) *TestProject {
	windows := make([]Window, len(windowNames))
	for i, name := range windowNames {
		windows[i] = NewWindow(name)
	}
	return &TestProject{Project{
		Name:    CreateRandomProjectName(),
		Windows: windows,
	}}
}

func (p *TestProject) AppendNamedWindow(windowName string) {
	p.Project.Windows = append(p.Project.Windows, Window{Name: "Window-3"})
}

func (p *TestProject) ReplaceWindowNames(windowNames ...string) {
	windows := make([]Window, len(windowNames))
	for i, name := range windowNames {
		windows[i] = NewWindow(name)
	}
	p.Project.Windows = windows
}
