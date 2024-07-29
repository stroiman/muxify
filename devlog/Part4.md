# July 22nd 2024 - Isolate test from dev env, and sync in control mode

tldr; Basic control of window configuration, and restore missing windows when
re-launching a session, but was left with an erratic test.

---

Things went a little better today, I started building the functionality to
control the windows, and create windows by name, as well as recreating missing
windows when re-launching a project. This still needs to deal with correct
ordering combined with user created windows. Although I have a suspicion that it
does work as intended out of the box, it will need to be tested.

I further improved the tests by adding helper functions to control session and
server lifetime, reducing the repeated code of checking for errors, and killing
the session. I also explicitly kill the tmux server after the tests run, just as
a final cleanup.

I ran into some issues with erratic tests, which did annoy my. I saw two
different symptoms. One error from occurred occasionally when starting a
session, and another encountered unexpected output when checking the working
directory. I think the start error may have been fixed by randomising session
name. The other error, I didn't find out what was the issue. I added that would
execute in the case of it occurring again, writing the actual pane output to
console out, to help understand what goes wrong.
