# 05 — Time controls

**What to build:** the player can stop time to think, skip the hours they do not care
about, and choose the pace of a session.

Pausing simply stops advancing the clock. Waiting runs the simulation forward without
rendering until a condition is met: dawn, dusk, or a named Townsperson changing position.
A wait aborts the moment its condition is met or something notable happens, so skipping
time never costs the player the thing they were waiting for.

Waiting until a particular person moves is not just a convenience — it is the player
stating a hypothesis about that person's Schedule, which is the closest the game comes to
a spying verb.

The Time Multiplier is configuration, fixed for a session, not a runtime slider. At the
default a Day is twenty-four minutes; x2 gives twelve and x4 gives six. **It scales the
entire simulation, walking speed included.** A multiplier that scaled only the calendar
would leave Schedules unwalkable at higher settings and nowhere else — the worst class of
defect to chase. Movement animation durations scale with it.

**Blocked by:** 04 — Real-time clock.

**Status:** ready-for-agent

- [ ] The player can pause and resume at any moment
- [ ] The player can wait until dawn, until dusk, or until a named Townsperson moves
- [ ] A wait is interrupted when something notable happens
- [ ] The Time Multiplier is read from configuration and supports at least x1, x2 and x4
- [ ] A test asserts the simulation reaches identical state at every multiplier for the
      same number of in-game hours
