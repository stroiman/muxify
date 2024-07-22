# July 15th  2024 - Session working dir and test sync

tldr; Starting the session with a specific working directory, and verifying that
the directory is set correctly. But coming to many dead ends trying to optimise
how this is tested.

---

In this session, I added the option to control the working directory for a
project. The first verification was accomplished by executing `ECHO $PWD` inside
a session, and reading the content of a pane using the `capture-output` tmux
command. But that is a completely different process than my test, so I don't
know _when_ the system is ready for verification, and the solution was to sleep
long enough for the `echo` command to have completed. 

This did work, but I wanted to improve the test setup. Things like `sleep`
should be avoided at all costs in tests. They add unnecessary delay, and when
you have 10 or 100 tests with a sleep, the test suite is delayed significantly.

That lead to a rather long, and many dead ends due to a crucial
misunderstanding. I looked at what options tmux has for emitting events, and I
discovered that "Control Mode" will emit events to stdout, which includes text
written by programs in tmux. That would allow my to create an event the test
could react to know when it was ready for verification.

But when control mode did emit all the events when I tried it out manually, it
didn't emit any relevant events in the tests. I finally figured out that control
mode needs to be attached to a specific session. And although I had tried to do
that before, that generated an error, as I didn't realise that the `-C` flag is
a server option that must be placed _before_ the command.

Finally I managed to get events in the test, but I didn't manage to finish the
task due to lack of time.
