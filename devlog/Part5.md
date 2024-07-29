# July 29th 2024 - Start multiple panes in each window

tldr; Reorganise windows, massive refactoring, and adding support to start
know will need to change. Added the ability to start multiple panes.

---

The goal was to handle multiple panes, but after adding `Panes []Pane` to the
`Window`, it was no longer applicable as a map key. So I added an `Id` instead,
first uninitialised; which should make the tests break; which they didn't

I then discovered that the use of `map` was unnecessarily complex for the
cases that _were_ solved. But necessary for handling reordering windows after a
relaunch. So I added tests for that scenario (after an unproductive detour -
because the test I wrote to provoke a break did break; but that was because the
test was bugged). Eventually, I implemented the window reordering, after which I
could use an `Id` as a map key, and add the `panes` to the `Window`.

That included a lot of refactoring, and the code to control window arrangement
was significantly improved, e.g. by the `WindowTarget` option controlling
multiple arguments, e.g. `-b` vs. `-a`, as well as `-t target`. And by being
applicatble in both `CreateWindow`, and `MoveWindow`, many branches were reduced
to two relevant questions.

- Does the window currently exist?
  - If yes, move it to the target position
  - If no, create it at the target position
- Is it the first window? 
  - Yes: The target position is _before_ the currently first window
  - No: The target position is _after_ the previously configured window

Finally, I could commence the task I had set out to perform; launch multiple
panes in a Window. The initial version was created that _always_ creates panes;
but doesn't check if it already exists.

That will be the goal for next session.
