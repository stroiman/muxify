package muxify

import (
	"io"

	"github.com/go-yaml/yaml"
)

type project struct {
	Name             string
	WorkingDirectory string
	Windows          []Window
	Tasks            map[string]Task
}

func Decode(reader io.Reader) (p Project, err error) {
	decoder := yaml.NewDecoder(reader)
	err = decoder.Decode(&p)
	return
}
