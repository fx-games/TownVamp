# 01 — Headless test seam

**What to build:** a Go test that constructs the game with the null user interface,
advances the clock, and asserts on world state. This is the single seam every later
ticket tests through, and nothing has ever proven the game is drivable headlessly — the
null UI and the console mode strongly suggest it is, but neither is proof.

The test should be trivial in content. Construct the game, ask it something a player
could see, assert. Its value is entirely in establishing that the seam works and setting
the pattern.

Use the standard library testing package. No framework, no fixtures, no mocking library.

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

- [ ] The game can be constructed in a test with the null user interface and no graphical
      subsystem
- [ ] A test advances the simulation by a given in-game duration
- [ ] A test asserts on observable world state — an actor's position, a Townsperson's
      Presence, the vampire's Blood — without reaching into how that state was computed
- [ ] The simulation is seeded, so a test that needs a specific outcome fixes the seed
      rather than asserting on a distribution
- [ ] Tests run from a clean checkout with no manual setup
