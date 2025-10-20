package main

import (
	"errors"
	"io"
	"io/fs"
	"os"
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
		p.WorkingDirectory = os.ExpandEnv(p.WorkingDirectory)
		for wi := range p.Windows {
			p.Windows[wi].EnsureValid()
		}
		config.Projects[pi] = p
	}
	return
}

func DecodeProject(reader io.Reader) (p Project, err error) {
	decoder := yaml.NewDecoder(reader)
	err = decoder.Decode(&p)
	p.WorkingDirectory = os.ExpandEnv(p.WorkingDirectory)
	for wi := range p.Windows {
		p.Windows[wi].EnsureValid()
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
	// path, err := getConfigDirPath(os)
	// if err != nil {
	// 	return
	// }
	entries, err := fs.ReadDir(dir, ".")
	if err != nil {
		return
	}
	// TODO: Logic needs to be rewritten for better error handling!
	// file, err := dir.Open("projects.yaml")
	// if err == nil {
	// 	defer func() {
	// 		closeErr := file.Close()
	// 		if err == nil {
	// 			err = closeErr
	// 		}
	// 	}()
	// 	config, err = Decode(file)
	// }
	for _, entry := range entries {
		if entry.Name() == "projects.yaml" {
			file, err := dir.Open(entry.Name())
			if err != nil {
				return config, err
			}
			defer func() {
				closeErr := file.Close()
				if err == nil {
					err = closeErr
				}
			}()
			childConfig, err := Decode(file)
			if err != nil {
				return config, err
			}
			config.Projects = append(config.Projects, childConfig.Projects...)
			continue
		}
		if entry.IsDir() {
			continue
		}
		file, err := dir.Open(entry.Name())
		if err != nil {
			return config, err
		}
		project, err := loadProjectFile(file)
		if err != nil {
			return config, err
		}
		config.Projects = append(config.Projects, project)
	}
	if len(config.Projects) == 0 {
		return config, errors.New("muxify: project not found in config dir")
	}
	return
}

func loadProjectFile(file fs.File) (Project, error) {
	defer file.Close()
	return DecodeProject(file)
}

func getOSConfigDir(os OS) (string, error) {
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

func getConfigDirPath(os OS) (string, error) {
	configDir, err := getOSConfigDir(os)
	res := path.Join(configDir, getAppName(os))
	return res, err
}

func getConfigDir(os OS) (dir fs.FS, err error) {
	path, err := getConfigDirPath(os)
	if err == nil {
		dir = os.Dir(path)
	}
	return
}
