package muxify_test

import (
	"io/fs"
	"strings"
	"testing/fstest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/stroiman/muxify"

	"github.com/google/go-cmp/cmp/cmpopts"
)

type FakeOS struct {
	files fstest.MapFS
	env   map[string]string
}

func (os FakeOS) LookupEnv(key string) (string, bool) {
	value, ok := os.env[key]
	return value, ok
}

func (os FakeOS) Dir(base string) fs.FS {
	result := make(fstest.MapFS)
	prefix := base + "/"
	for path, file := range os.files {
		if newPath, ok := strings.CutPrefix(path, prefix); ok {
			result[newPath] = file
		}
	}
	return result
}

var _ = Describe("Configuration", Ordered, func() {
	var fakeOs FakeOS
	var projectsConfigFile *fstest.MapFile

	BeforeAll(func() {
		projectsConfigFile = &fstest.MapFile{
			Data: []byte(example_config),
			Mode: fs.ModePerm,
		}
		DeferCleanup(func() {
			projectsConfigFile = nil // Allow GC
		})
	})

	BeforeEach(func() {
		fakeOs = FakeOS{
			fstest.MapFS{},
			map[string]string{"HOME": "/users/foo"},
		}
	})

	Describe("Default config location", func() {
		BeforeEach(func() {
			fakeOs.files["/users/foo/.config/muxify/projects"] = projectsConfigFile
		})

		It("Should deserialise a full configuration", func() {
			project, err := ReadConfiguration(fakeOs)
			Expect(err).ToNot(HaveOccurred())
			expected := MuxifyConfiguration{
				Projects: []Project{{
					Name:             "Project 1",
					WorkingDirectory: "/work",
					Windows: []Window{
						{Name: "window-1", Panes: []string{"editor", "test-runner"}},
						{Name: "window-2", Panes: []string{"dev"}}},
					Tasks: map[string]Task{
						"editor": {Commands: []string{"nvim"}},
						"test-runner": {
							Commands: []string{"docker-compose up -d", "pnpm test:watch"},
						},
						"dev": {},
					},
				}}}
			Expect(project).To(BeComparableTo(expected, cmpopts.IgnoreUnexported(Window{})))
		})
	})

	Describe("User has overridden XDG_CONFIG_HOME", func() {
		BeforeEach(func() {
			fakeOs.env["XDG_CONFIG_HOME"] = "/var/config"
		})

		It("Should succeed when the file is under the new location", func() {
			fakeOs.files["/var/config/muxify/projects"] = projectsConfigFile
			_, err := ReadConfiguration(fakeOs)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should return an error when the file is only in the default location", func() {
			fakeOs.files["/users/foo/.config/muxify/projects"] = projectsConfigFile
			_, err := ReadConfiguration(fakeOs)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("User has specified a MUXIFY_APPNAME", func() {
		BeforeEach(func() {
			fakeOs.env["MUXIFY_APPNAME"] = "muxer"
		})

		It("Should succeed when the file is under the specified folder", func() {
			fakeOs.files["/users/foo/.config/muxer/projects"] = projectsConfigFile
			_, err := ReadConfiguration(fakeOs)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should return an error when the file is only in the default location", func() {
			fakeOs.files["/users/foo/.config/muxify/projects"] = projectsConfigFile
			_, err := ReadConfiguration(fakeOs)
			Expect(err).To(HaveOccurred())
		})
	})

	It("Should generate an valid configuration", func() {
		reader := strings.NewReader(example_config)
		config, err := Decode(reader)
		Expect(err).ToNot(HaveOccurred())
		project, _ := config.GetProject("Project 1")
		Expect(project.Validate()).To(Succeed())
	})

	It("Should allow a missing working_dir", func() {
		reader := strings.NewReader("projects:\n- name: \"Project 1\"")
		config, err := Decode(reader)
		Expect(err).ToNot(HaveOccurred())
		project, _ := config.GetProject("Project 1")
		Expect(project.WorkingDirectory).To(Equal(""))
	})
})

var example_config = `
projects:
  - name: "Project 1"
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
