package main_test

import (
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	. "github.com/stroiman/muxify"

	"github.com/google/go-cmp/cmp/cmpopts"
)

type Env = map[string]string

type FakeOS struct {
	files fstest.MapFS
	env   Env
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

type ConfigurationTestSuite struct {
	GomegaSuite
	fakeOs             FakeOS
	projectsConfigFile *fstest.MapFile
}

func fakeConfigFile(data string) *fstest.MapFile {
	return &fstest.MapFile{
		Data: []byte(data),
		Mode: fs.ModePerm,
	}
}

func (s *ConfigurationTestSuite) SetupSuite() {
	s.projectsConfigFile = &fstest.MapFile{
		Data: []byte(example_config),
		Mode: fs.ModePerm,
	}
}

func (s *ConfigurationTestSuite) SetupTest() {
	s.GomegaSuite.SetupTest()
	s.fakeOs = FakeOS{
		fstest.MapFS{},
		map[string]string{"HOME": "/users/USER"},
	}
}

type DefaultConfigSuiteTestSuite struct {
	ConfigurationTestSuite
}

func (s *DefaultConfigSuiteTestSuite) SetupTest() {
	s.ConfigurationTestSuite.SetupTest()
	s.fakeOs.files["/users/USER/.config/muxify/projects.yaml"] = s.projectsConfigFile
}

func TestDefaultConfigSuite(t *testing.T) {
	suite.Run(t, new(DefaultConfigSuiteTestSuite))
}

func (s *DefaultConfigSuiteTestSuite) TestDeserializeFullConfiguration() {
	project, err := ReadConfiguration(s.fakeOs)
	s.Expect(err).ToNot(HaveOccurred())
	expected := MuxifyConfiguration{
		Projects: []Project{{
			Name:             "Project 1",
			WorkingDirectory: "/work",
			Windows: []Window{
				{
					Name:   "window-1",
					Layout: "vertical",
					Panes:  []string{"editor", "test-runner"},
				},
				{Name: "window-2", Layout: "", Panes: []string{"dev"}}},
			Tasks: map[string]Task{
				"editor": {Commands: []string{"nvim"}},
				"test-runner": {
					WorkingDirectory: "sub-dir",
					Commands:         []string{"docker-compose up -d", "pnpm test:watch"},
				},
				"dev": {},
			},
		}}}
	s.Expect(project).To(BeComparableTo(expected, cmpopts.IgnoreUnexported(Window{})))
}

func (s *DefaultConfigSuiteTestSuite) TestExpandEnvVars() {
	os.Setenv("MUX_TEST_VALUE", "/user/USER")
	s.projectsConfigFile.Data = []byte(`projects:
  - name: project-1
    working_dir: $MUX_TEST_VALUE/work`)
	projects, err := ReadConfiguration(s.fakeOs)
	s.Expect(err).ToNot(HaveOccurred())
	expected := MuxifyConfiguration{
		Projects: []Project{{Name: "project-1", WorkingDirectory: "/user/USER/work"}},
	}
	s.Expect(projects).To(BeComparableTo(expected, cmpopts.IgnoreUnexported(Window{})))
}

type XDGOverwrittenTestSuite struct {
	ConfigurationTestSuite
}

func (s *XDGOverwrittenTestSuite) SetupTest() {
	s.ConfigurationTestSuite.SetupTest()
	s.fakeOs.env["XDG_CONFIG_HOME"] = "/var/config"
}

func Test(t *testing.T) {
	suite.Run(t, new(XDGOverwrittenTestSuite))
}

func (s *XDGOverwrittenTestSuite) TestFileIsUnderTheRedirectedLocation() {
	s.fakeOs.files["/var/config/muxify/projects.yaml"] = s.projectsConfigFile
	_, err := ReadConfiguration(s.fakeOs)
	s.Expect(err).ToNot(HaveOccurred())
}

func (s *XDGOverwrittenTestSuite) TestFailWhenFileIsNotInTheNewLoactionButDefault() {
	s.fakeOs.files["/users/USER/.config/muxify/projects.yaml"] = s.projectsConfigFile
	_, err := ReadConfiguration(s.fakeOs)
	s.Expect(err).To(HaveOccurred())
}

type AppNameOverwrittenTestSuite struct {
	ConfigurationTestSuite
}

func TestAppNameOverwritten(t *testing.T) {
	suite.Run(t, new(AppNameOverwrittenTestSuite))
}

func (s *AppNameOverwrittenTestSuite) SetupTest() {
	s.ConfigurationTestSuite.SetupTest()
	s.fakeOs.env["MUXIFY_APPNAME"] = "muxer"
}

func (s *AppNameOverwrittenTestSuite) TestSuccessWhenFileIsInNewFolder() {
	s.fakeOs.files["/users/USER/.config/muxer/projects.yaml"] = s.projectsConfigFile
	_, err := ReadConfiguration(s.fakeOs)
	s.Expect(err).ToNot(HaveOccurred())
}

func (s *AppNameOverwrittenTestSuite) TestFailureWhenFileIsInDefaultFolder() {
	s.fakeOs.files["/users/USER/.config/muxify/projects.yaml"] = s.projectsConfigFile
	_, err := ReadConfiguration(s.fakeOs)
	s.Expect(err).To(HaveOccurred())
}

type ParseConfigTestSuite struct {
	GomegaSuite
}

func (s *ParseConfigTestSuite) TestParseValidConfig() {
	reader := strings.NewReader(example_config)
	config, err := Decode(reader)
	s.Expect(err).ToNot(HaveOccurred())
	project, _ := config.GetProject("Project 1")
	s.Expect(project.Validate()).To(Succeed())
}

func (s *ParseConfigTestSuite) TestHandleMissingWorkingDir() {
	reader := strings.NewReader("projects:\n- name: \"Project 1\"")
	config, err := Decode(reader)
	s.Expect(err).ToNot(HaveOccurred())
	project, _ := config.GetProject("Project 1")
	s.Expect(project.WorkingDirectory).To(Equal(""))
}

func TestParseConfig(t *testing.T) {
	suite.Run(t, new(ParseConfigTestSuite))
}

var example_config = `
projects:
  - name: "Project 1"
    working_dir: "/work"
    windows:
      - name: "window-1"
        layout: vertical
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
        working_dir: sub-dir
        commands:
          - docker-compose up -d
          - pnpm test:watch
      dev:
`

func TestMultipleFiles(t *testing.T) {
	project1config := `
name: "project-1"
working_dir: "/src/project-1"
`
	project2config := `
name: "project-2"
working_dir: "/src/project-2"
`
	fakeOs := FakeOS{
		files: fstest.MapFS{
			"/users/USER/.config/muxify/project1.yaml": fakeConfigFile(project1config),
			"/users/USER/.config/muxify/project2.yaml": fakeConfigFile(project2config),
		},
		env: Env{"HOME": "/users/USER"},
	}
	config, err := ReadConfiguration(fakeOs)
	assert.NoError(t, err)
	expected := MuxifyConfiguration{
		Projects: []Project{
			{Name: "project-1", WorkingDirectory: "/src/project-1"},
			{Name: "project-2", WorkingDirectory: "/src/project-2"},
		},
	}
	gomega.NewWithT(t).
		Expect(config).
		To(BeComparableTo(expected, cmpopts.IgnoreUnexported(Window{})))
}
