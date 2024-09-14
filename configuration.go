package main

import (
	"errors"
	"io"
	"io/fs"
	"path"

	"gopkg.in/yaml.v3"
)

type MuxifyConfiguration struct {
	Projects []Project
}

func (c MuxifyConfiguration) GetProject(name string) (Project, bool) {
	for _, p := range c.Projects {
		if p.Name == name {
			return p, true
		}
	}
	return Project{}, false
}

func Decode(reader io.Reader) (config MuxifyConfiguration, err error) {
	decoder := yaml.NewDecoder(reader)
	err = decoder.Decode(&config)
	for pi, p := range config.Projects {
		for wi, w := range p.Windows {
			p.Windows[wi] = NewWindow(w.Name, w.Panes...)
		}
		config.Projects[pi] = p
	}
	return
}

type OS interface {
	Dir(path string) fs.FS
	LookupEnv(key string) (string, bool)
}

func ReadConfiguration(os OS) (config MuxifyConfiguration, err error) {
	dir, err := getConfigDir(os)
	if err != nil {
		return
	}
	file, err := dir.Open("projects")
	if err == nil {
		defer func() {
			closeErr := file.Close()
			if err == nil {
				err = closeErr
			}
		}()
		config, err = Decode(file)
	}
	return
}
func getConfigDirPath(os OS) (string, error) {
	if configDir, configDirFound := os.LookupEnv("XDG_CONFIG_HOME"); configDirFound {
		return configDir, nil
	}
	if homeDir, found := os.LookupEnv("HOME"); found {
		return path.Join(homeDir, ".config"), nil
	}
	return "", errors.New("Home dir not configured")
}

func getAppName(os OS) string {
	if appName, ok := os.LookupEnv("MUXIFY_APPNAME"); ok {
		return appName
	} else {
		return "muxify"
	}
}

func getConfigDir(os OS) (dir fs.FS, err error) {
	configDir, err := getConfigDirPath(os)
	if err == nil {
		appName := getAppName(os)
		dir = os.Dir(path.Join(configDir, appName))
	}
	return
}
