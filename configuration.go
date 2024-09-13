package muxify

import (
	"io"

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
