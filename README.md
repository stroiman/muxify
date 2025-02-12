# Muxify - tmux session manager

This tool is supposed to help manage tmux sessions for larger projects, e.g.
with multiple 'packages', e.g. front-end, back-end, and shared code, which are
all editied in individual editors, grouped into separate tmux windows.

## How is this different from other tools?

This tool embraces that your TMUX environment may change, but you generally have
a _desired_ configuration that can be expressed in a configuration.

This means this tool is designed to be executed while you have the session
running. Any missing windows and panes will be recreated. Did you accidentally
close a window you didn't mean to? Start the project, and the missing window
will be recreated.

### Multiple layouts

This features is not yet implemented, but an original design goal was to support
multiple layouts for different tasks, or different working environment.

E.e., in my office environment, I have the desktop monitor and 2 external
monitors. Here, I'd like to have the center monitor for my editor, and the other
external monitof for feedback. Feedback could include one or more test runners
and compilers. Thus, I would have test runner and editor on separate windows.

But when just using the laptop, I'd want to have the test runner and editor as
two separate panes in the same TMUX window.

So for a single configured project, you should be able to switch between
different layouts.

### Why this tool

I have tried other tools, e.g. tmuxinator the tmux-resurrect. 

Tmuxinator, apart from being bugged[^1], was a one-time configuration. If you
change the configuration, you need to completely kill the session, and restart. 

Muxify can in that respect be described as a bug-free tmuxinator, that can apply
configuration changes to an existing session.

Tmux-resurrect served as a trusty replacement for tmuxinator for a long time.
But I had a lot of projects each with their own setup. Resurrect restores _all_
tmux sessions. The result was that I would have 10-20 tmux sessions running,
when only a few were relevent.

## Current state!

tldr; It works, but configuration format _will_ change.

- The configuration files doesn't yet support multiple layouts for a project.
- Error messages are very poor.
- There is no validation, e.g. a project and window names must be valie project
  and session names; but that is not validated when loading the configuration.

### Configuration file

The tool is currently working, i.e. you can run it, it will read a configuration
from `$XDG_CONFIG_HOME/muxify/projects.yaml` or fallback to
`$HOME/.config/muxify/projects.yaml`, and initialise a tmux session. The tool
will not start a client, see the [tips](#Tips) below.

Error messages are also notoriously poor, and there is little validation of
configuration file. E.g. a project name _must_ be a valid tmux session name; but
there is no validation of that yet.

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

## Configuration file

This is a simple example configuration file. More windows can easily be created
by adding to the list.

```yaml
projects:
  - name: muxify
    working_dir: $HOME/src/muxify
    tasks:
      editor:
        commands:
          - nvim .
      test:
        commands:
          - gow test
    windows:
      - name: Editor
        layout: horizontal
        panes:
          - editor
          - test
```

## Tips - the `m` command

Currently, the tool only creates and configures sessions, windows, and panes on
a tmux server; but it does not launch a client. I use the following shell
script which automates launching a tmux client; or switching session if run from
_inside_ an existing tmux session.

This script isn't _yet_ part of muxify, but should be.

I call this `m` - and have manually created it in my search path.

```sh
#!/bin/sh

set -xe
muxify $1 

if [ -z "$TMUX" ]; then
  tmux attach -t $1
else
  tmux switch-client -t $1
fi
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

### [July 22nd 2024 - Isolate test from dev env, and sync in control mode](https://github.com/stroiman/muxify/blob/main/devlog/Part4.md)

Basic control of window configuration, and restore missing windows when
re-launching a session, but was left with an erratic test.

### [July 29th 2024 - Start multiple panes in each window](https://github.com/stroiman/muxify/blob/main/devlog/Part5.md)

Reorganise windows, massive refactoring, and adding support to start know will
need to change. Added the ability to start multiple panes.


### [September 12-14th 2024 - Create an executable](https://github.com/stroiman/muxify/blob/main/devlog/Part6.md)

Create an executable binary that reads configuration from a file, and I am now
starting to use the tool.

---

[^1]: Sessions started with tmuxinator someimes ran into weird states where the
    same app would be shown in multiple panes, or keyboard input would be
    directed to multiple shells. I also saw extreme lag with grouped tmux sessions.
