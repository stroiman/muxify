package muxify

import (
	"io"

	"github.com/go-yaml/yaml"
)

func Decode(reader io.Reader) (p Project, err error) {
	decoder := yaml.NewDecoder(reader)
	err = decoder.Decode(&p)
	for i, w := range p.Windows {
		p.Windows[i] = NewWindow(w.Name, w.Panes...)
	}
	return
}
