package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
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
	Stdout io.Writer
}

func (cli CLI) startProject(projectName string) error {
	configuration, err := ReadConfiguration(cli)
	if err != nil {
		return err
	}
	if project, ok := configuration.GetProject(projectName); ok {
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

func (cli CLI) Run(args []string) error {
	var verbose bool
	var err error
	rootCmd := &cobra.Command{
		Short: "Muxify - automate tmux",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				slog.SetLogLoggerLevel(slog.LevelDebug)
			} else {
				slog.SetLogLoggerLevel(slog.LevelWarn)
			}
		},
		Args: cobra.NoArgs,
	}
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging")
	rootCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List the available configurations",
		Run: func(cmd *cobra.Command, args []string) {
			var conf MuxifyConfiguration
			conf, err = ReadConfiguration(cli)
			if err != nil {
				return
			}
			for _, p := range conf.Projects {
				fmt.Fprintf(cli.Stdout, "%s\n", p.Name)
			}
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "launch",
		Short: "Launch a project",
		Long:  "Launch a project, creating a tmux session if none exists, or potentially recreates closed panes and windows if the session exists.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectName := args[0]
			err = cli.startProject(projectName)
		},
	})
	rootCmd.SetArgs(args[1:])
	if exErr := rootCmd.Execute(); exErr != nil {
		return exErr
	}
	return err
}

type RealOS struct{}

func (o RealOS) Dir(name string) fs.FS {
	return os.DirFS(name)
}

func (o RealOS) LookupEnv(name string) (string, bool) {
	return os.LookupEnv(name)
}

func main() {
	err := CLI{DefaultRunner{}, RealOS{}, os.Stdout}.Run(os.Args)
	if err == nil {
		os.Exit(0)
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}
