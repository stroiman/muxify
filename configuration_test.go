package muxify_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/stroiman/muxify"

	"github.com/google/go-cmp/cmp/cmpopts"
)

var _ = Describe("Configuration", func() {
	It("Should deserialise a full configuration", func() {
		reader := strings.NewReader(example_config)
		project, err := Decode(reader)
		Expect(err).ToNot(HaveOccurred())
		Expect(project.Name).To(Equal("Project 1"))
		Expect(project.WorkingDirectory).To(Equal("/work"))
		expected := Project{
			Name:             "Project 1",
			WorkingDirectory: "/work",
			Windows: []Window{
				{Name: "window-1", Panes: []string{"editor", "test-runner"}},
				{Name: "window-2", Panes: []string{"dev"}}},
			Tasks: map[string]Task{
				"editor":      {Commands: []string{"nvim"}},
				"test-runner": {Commands: []string{"docker-compose up -d", "pnpm test:watch"}},
				"dev":         {},
			},
		}
		Expect(project).To(BeComparableTo(expected, cmpopts.IgnoreUnexported(Window{})))
	})

	It("Should generate an valid configuration", func() {
		reader := strings.NewReader(example_config)
		project, err := Decode(reader)
		Expect(err).ToNot(HaveOccurred())
		Expect(project.Validate()).To(Succeed())
	})

	It("Should allow a missing working_dir", func() {
		reader := strings.NewReader(`name: "Project"`)
		project, _ := Decode(reader)
		Expect(project.WorkingDirectory).To(Equal(""))
	})
})

var example_config = `
name: "Project 1"
working_dir: "/work"
windows:
  - name: "window-1"
    panes:
      - editor
      - test-runner
  - name: "window-2"
    panes:
      - dev
tasks:
  editor:
    commands:
      - nvim
  test-runner:
    commands:
      - docker-compose up -d
      - pnpm test:watch
  dev:
`

// tasks:
//   editor:
//     commands:
//       - nvim .
//   test-watcher:
//     commands:
//       - pnpm test:watch
//   compiler:
//     commands:
//       - pnpm dev
// windows:
//   - name: window-1
//     panes:
//       - editor
//       - test-watcher
//   - name: window-2
//     panes:
//       - compiler
// `
