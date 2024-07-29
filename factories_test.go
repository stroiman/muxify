package muxify_test

import (
	. "github.com/stroiman/muxify"
)

func CreateProject() Project {
	return Project{
		Name: CreateRandomName(),
	}
}

func CreateProjectWithWindowNames(windowNames ...string) Project {
	windows := make([]Window, len(windowNames))
	for i, name := range windowNames {
		windows[i] = NewWindow(name)
	}
	return Project{
		Name:    CreateRandomProjectName(),
		Windows: windows,
	}
}

func AppendNamedWindowToProject(proj *Project, windowName string) {
	proj.Windows = append(proj.Windows, Window{Name: "Window-3"})
}

func ReplaceWindowNames(proj *Project, windowNames ...string) {
	windows := make([]Window, len(windowNames))
	for i, name := range windowNames {
		windows[i] = NewWindow(name)
	}
	proj.Windows = windows
}
