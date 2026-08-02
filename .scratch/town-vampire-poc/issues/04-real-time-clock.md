# 04 — Real-time clock

**What to build:** time advances on its own. The player stands still and the world keeps
moving — actors act, the clock runs, the time of day changes. This is the conversion from
roguelike turns to a running simulation, and everything after it depends on it.

Per ADR 0001, the AUT and energy model is kept exactly as it is and only its driver
changes. The player-turn entry point on the timekeeper becomes a duration-advancing call
made from the fixed-timestep update rather than from the end-turn action, and the player
becomes an ordinary timed being with its own energy budget. Per-actor speeds, action
costs and status ticking survive untouched. Movement stays grid-discrete; smooth motion
remains presentation.

At the default pace one Day is twenty-four minutes of real time, which makes one tick
equal one in-game second.

**The prerequisite hiding inside this ticket:** the broker currently drains its queued
actions only when zero animations are running anywhere in the world. With a town animating
continuously that is never true and the queue deadlocks. The gate must become per-actor,
or go. Nothing else in this project works until it does.

**Blocked by:** 02 — Demolish the RPG systems.

**Status:** ready-for-agent

- [ ] The simulation advances from the fixed-timestep update, not from a player action
- [ ] The player is an ordinary timed being subject to the same energy rules as everyone
      else
- [ ] Several actors can act and animate simultaneously without the action queue stalling
- [ ] Per-actor speed, action cost in AUT and status ticking behave as they did before
- [ ] The time of day advances continuously and is visible to the player
- [ ] The value of an AUT in seconds is configurable rather than hard-coded
