# July 8th 2024 - New project, and a tmux session

tldr; Creating a new project, and setting up a test runner. The outcome of this
session was a simple function to ensure a session with a specific name was
started.

---

It has been 7 years since I worked on a Go project, so the first was problem was
to get a grip on how to setup the project structure. Some things have improved
in go, e.g. better dependency management and generic support. Also, better
support for creating tools that don't need to be go-gettable. I started
following that paradigm, but realised that this _should_ be go gettable,
allowing go developers to install the tool using go tools.

Next, I needed get my test runner running. I opted for
[Ginkgo](https://github.com/onsi/ginkgo) as I've used it before, but I may
choose differently. Most of it comes down to what provides the best feedback the
fastest.

Finally, I managed to implement the basic setting, I can start a session, and
starting the same session again is ignored.

