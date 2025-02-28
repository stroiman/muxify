//go:generate mockgen -source=cli.go -destination=cli_mocks_test.go -package=main_test
package main_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	. "github.com/stroiman/muxify"
	"go.uber.org/mock/gomock"
)

func TestCli(t *testing.T) {
	fakeOs := FakeOS{
		files: fstest.MapFS{"/users/foo/.config/muxify/projects.yaml": &fstest.MapFile{
			Data: []byte(configuration),
		}},
		env: map[string]string{"HOME": "/users/foo"},
	}
	controller := gomock.NewController(t)
	mock := NewMockRunner(controller)
	cli := CLI{mock, fakeOs}
	call := mock.EXPECT().Run(gomock.Any())
	var actualProject Project
	call.Do(func(project Project) {
		actualProject = project
	})
	cli.Run([]string{"muxify", "Project 1"})
	controller.Finish()
	assert.Equal(t, Project{Name: "Project 1"}, actualProject)
}

var configuration = `projects:
  - name: Project 1
  - name: Project 2
`
