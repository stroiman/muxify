package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

type DefaultRunner struct {
}

func (r DefaultRunner) Run(p Project) error {
	_, err := p.EnsureStarted(TmuxServer{})
	return err
}

type Runner interface {
	Run(p Project) error
}

type CLI struct {
	Runner
	OS
}

func (cli CLI) Run(args []string) error {
	if len(args) < 2 {
		return errors.New("No argument")
	}
	configuration, err := ReadConfiguration(cli)
	if err != nil {
		return err
	}
	if project, ok := configuration.GetProject(args[1]); ok {
		return cli.Runner.Run(project)
	} else {
		var b strings.Builder
		b.WriteString("The project was not found. Valid project names are:\n")
		for _, p := range configuration.Projects {
			b.WriteString(fmt.Sprintf(" - %s\n", p.Name))
		}
		return errors.New(b.String())
	}
}

type RealOS struct{}

func (o RealOS) Dir(name string) fs.FS {
	return os.DirFS(name)
}

func (o RealOS) LookupEnv(name string) (string, bool) {
	return os.LookupEnv(name)
}

func main() {
	err := CLI{DefaultRunner{}, RealOS{}}.Run(os.Args)
	if err == nil {
		os.Exit(0)
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}
