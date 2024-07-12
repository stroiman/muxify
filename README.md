# Muxify - tmux session manager

This tool is supposed to help manage tmux sessions for larger projects, e.g.
with multiple 'packages', e.g. front-end, back-end, and shared code, whach are
all editied in individual editors, are started individually in separate
terminals.

It should be possible to maintain multiple tmux sessions for one project. This
is because a tmux session has a specific size matching the _smallest_ terminal
that currently displays it.

Sometimes you might want to have different tasks in different terminals with
different sizes, e.g. in a multi-monitor setup.

You might switch between a laptop display and a multi-monitor setup, and have
different views depending on the setup. Or depending on the current task, e.g.
are tests in focus, or visual feedback for UI development

The tool should support different layouts, and quickly switch between them
without killing the running tasks.

## General idea

You define 3 different concepts

- Project - a larger project you work on.
- Task - a "program" that needs to run while working on the project. E.g.
  - An editor in the terminal
  - Development web server
  - Compiler running in watch mode, e.g. for TypeScript, `tsc -w`
  - A test runner running in watch mode
- A layout
  - A configuration for the different tasks, and how they are organised in
    sessions, windows, and panes.

## Choice of programming language

This is written in Go
- A modern language with a good development experience, e.g. TDD with reasonably
  fast feedback.
- The code can be packaged into a compiled binary that can be distributed using
  native package managers.
