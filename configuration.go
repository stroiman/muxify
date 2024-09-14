package muxify

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
	// configDir, configDirFound := os.LookupEnv("XDG_CONFIG_HOME")
	// if !configDirFound {
	// 	homeDir, found := os.LookupEnv("HOME")
	// 	if !found {
	// 		err = errors.New("Home dir not configured")
	// 		return
	// 	}
	// 	configDir = path.Join(homeDir, ".config")
	// }
	// dir := os.Dir(configDir)
	file, err := dir.Open("muxify/projects")
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

func getConfigDir(os OS) (fs.FS, error) {
	if configDir, configDirFound := os.LookupEnv("XDG_CONFIG_HOME"); configDirFound {
		return os.Dir(configDir), nil
	}
	if homeDir, found := os.LookupEnv("HOME"); found {
		configDir := path.Join(homeDir, ".config")
		return os.Dir(configDir), nil
	}
	return nil, errors.New("Home dir not configured")
}
