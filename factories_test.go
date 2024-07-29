package muxify_test

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
