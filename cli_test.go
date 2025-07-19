//go:generate mockgen -source=cli.go -destination=cli_mocks_test.go -package=main_test
package main_test

import (
	"bytes"
	"io"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	. "github.com/stroiman/muxify"
	"go.uber.org/mock/gomock"
)

func TestCliLaunchProject(t *testing.T) {
	controller := gomock.NewController(t)
	mock := NewMockRunner(controller)
	cli := CLI{mock, fakeOs, io.Discard}
	call := mock.EXPECT().Run(gomock.Any())
	var actualProject Project
	call.Do(func(project Project) {
		actualProject = project
	})
	cli.Run([]string{"muxify", "launch", "Project 1"})
	controller.Finish()
	assert.Equal(t, Project{Name: "Project 1"}, actualProject)
}

func TestCliLaunchProjectWithVerbose(t *testing.T) {
	controller := gomock.NewController(t)
	mock := NewMockRunner(controller)
	cli := CLI{mock, fakeOs, io.Discard}
	call := mock.EXPECT().Run(gomock.Any())
	var actualProject Project
	call.Do(func(project Project) {
		actualProject = project
	})
	cli.Run([]string{"muxify", "-v", "launch", "Project 1"})
	controller.Finish()
	assert.Equal(t, Project{Name: "Project 1"}, actualProject)
}

func TestCliListProjects(t *testing.T) {
	var b bytes.Buffer
	cli := CLI{nil, fakeOs, &b}
	cli.Run([]string{"muxify", "list"})
	assert.Equal(t, "Project 1\nProject 2\n", b.String())
}

var configuration = `projects:
  - name: Project 1
  - name: Project 2
`

var fakeOs = FakeOS{
	files: fstest.MapFS{"/users/foo/.config/muxify/projects.yaml": &fstest.MapFile{
		Data: []byte(configuration),
	}},
	env: map[string]string{"HOME": "/users/foo"},
}
