//go:generate mockgen -source=cli.go -destination=cli_mocks_test.go -package=main_test
package main_test

import (
	"testing/fstest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/stroiman/muxify"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Cli", func() {
	It("Should start the configured project", func() {
		fakeOs := FakeOS{
			files: fstest.MapFS{"/users/foo/.config/muxify/projects": &fstest.MapFile{
				Data: []byte(configuration),
			}},
			env: map[string]string{"HOME": "/users/foo"},
		}
		controller := gomock.NewController(GinkgoT())
		mock := NewMockRunner(controller)
		cli := CLI{mock, fakeOs}
		call := mock.EXPECT().Run(gomock.Any())
		var actualProject Project
		call.Do(func(project Project) {
			actualProject = project
		})
		cli.Run([]string{"muxify", "Project 1"})
		controller.Finish()
		Expect(actualProject).To(Equal(Project{Name: "Project 1"}))
	})
})

var configuration = `projects:
  - name: Project 1
  - name: Project 2
`
