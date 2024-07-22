# July 16th 2024 - Isolate test from dev env, and sync in control mode

tldr; Fixed optimising testing as well as isolate the tests from the user's
environment.

---

After the last stream, I wanted to get control mode to work in the tests, which
I eventually did. 

I decided to introduce a `TmuxServer` type for helping send commands to tmux.
This was a significant improvement to the design, as it made it significantly to
control server options.

When that, I managed to create a new server on a new socket for testing, using a
custom configuration, also specifying using `/bin/sh` as a non-login shell, thus
not loading `.profile`, resulting in a better controlled environment, rather
than having the behaviour of tmux under test depend on local configuration of
the developer. 
