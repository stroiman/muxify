# July 19th 2025 - Refactor CLI parsing

tldr; After having used the tool for a long time, I decided to add command
parsing to the CLI to help build better integration _inside_ tmux.


---

I have been using the tool for a long time now. While not in the state matching
my original vision, it has been useful.

I wanted to bring better integration _inside_ tmux, which dictated a change to
the CLI interface, using specific _commands_, so the interface was into two
commands:

- `list` Show the available projects
- `launch` Launch a project.

I used [Cobra] to help refactor the CLI.

[Cobra]: https://cobra.dev/
