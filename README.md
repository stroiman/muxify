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

### Why this tool?

The need for this tool arose during my last larger project, where the code base
consisted of individual modules for on 3 different front-ends, catering for 3
different user roles. One common backend, and two modules with shared code; All
needed to be started individually; and all with their own test suites.

I had first used tmuxinator, but it just didn't work - and went on to use
tmux-resurrect. But now I kept all those processes alive when working on
something else; just to avoid having to recreate the setup every time I would
launch the project.

And neither of the tools helped me arrange stuff on multiple monitors; or adapt
the layout depending on the task I was working on.

## General idea

You define 3 different concepts

- Project - a larger project you work on.
- Task - a process that needs to run while working on the project. E.g.
  - An editor in the terminal
  - Development web server
  - Compiler running in watch mode, e.g. for TypeScript, `tsc -w`
  - A test runner in watch mode
- A layout
  - A configuration for the different tasks, and how they are organised in
    sessions, windows, and panes.

### Possible future idea

While learning about tmux, I realised that you can attach to a "command mode";
that can receive events in stdin. This could allow you to make tools that react
to, e.g. test-output. This could allow you to have the test runner hidden, but
have a desktop notification on 

This is an idea that arose when I realised the technical possibility exists. But
it is very low priority. But the use case itself is so useful that I want to
investigate this when I get further.

## Choice of programming language

This is written in Go
- A modern language with a good development experience, e.g. TDD with reasonably
  fast feedback.
- The code can be packaged into a compiled binary that can be distributed using
  native package managers.

## Possible configuration properties

This is just an example of where this might end up, but try to illustate the
scenario of swithing between a laptop on the road, or at the desk with multiple
monitors.

When using multiple monitors, you might want to break out the tests to the other
monitor, but on the road, you'd want the two to appear simultaneously.

```yaml
projects:
  - name: My project
  - working_folder: $HOME/src/my-project
  - tasks:
    - tests: pnpm test
    - edit: nvim .
  - layouts:
    - single_monitor:
      - sessions:
          my-project:
          - windows:
              main:
                panes:
                  - tasks:edit
                  - tasks:tests
    - multi_monitor:
      - sessions:
          my-project:
            - windows:
                editor:
                  panes:
                    - tasks:edit
          my-project-tests:
            - windows:
                editor:
                  panes:
                    - tasks:tests
```

## Note about the tests

The system is tested by actually starting a tmux server. The tests starts a new
tmux server, as to not interfere with a normal tmux server (assuming you didn't
launch a server on a socket named `muxify-test-socket`).

The server is loaded with the `tmux.conf` file in this project which sets
`default-command "/bin/sh"` - which _should_ use `/bin/sh` as the default shell
as a non-login shell, as to not load your own `$HOME/.profile`.

This _should_ test the system with a minimal configuration, which provides the
following benefits

 * Starts the tests in a controllable environment, not affected by the user's
   personal setup.
 * A minimal shell profile results in faster startup and execution of the tests.

## Developer log

### [July 8th 2024 - New project, and a tmux session](https://github.com/stroiman/muxify/blob/main/devlog/Part1.md)

Creating a new project, and setting up a test runner. The outcome of this
session was a simple function to ensure a session with a specific name was
started.

### [July 15th  2024 - Session working dir and test sync](https://github.com/stroiman/muxify/blob/main/devlog/Part2.md)

Starting the session with a specific working directory, and verifying that the
directory is set correctly. But coming to many dead ends trying to optimise how
this is tested.

### [July 16th 2024 - Isolate test from dev env, and sync in control mode](https://github.com/stroiman/muxify/blob/main/devlog/Part3.md)

Fixed optimising testing as well as isolate the tests from the user's
environment.

### [July 16th 2024 - Isolate test from dev env, and sync in control mode](https://github.com/stroiman/muxify/blob/main/devlog/Part4.md)

Fixed optimising testing as well as isolate the tests from the user's
environment.

