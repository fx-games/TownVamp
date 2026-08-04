# 14 — Epilogue and run endings

**What to build:** a run ends, and the player sees the shape of what they did.

A run ends four ways: the player leaves town by choice at any moment, or the vampire dies
— starved of Blood, caught by Sunlight, or caught by a Hunter. Leaving whenever they like
matters as much as the failure endings; it lets the player decide when their story here is
finished, and it makes the Epilogue reachable at any point for testing.

The Epilogue is one slide per Townsperson giving their end state and what followed from
it, plus one slide summarising the town as a whole.

**These are raw data readouts, not prose.** The simulation generates the story; the
Epilogue only reports it. Generated or authored prose — including any later use of
language models to write it — is explicitly out of scope, and the readout should be
structured so that swapping in prose later does not mean rebuilding what feeds it.

**The Epilogue does not score the player.** It describes consequences and says nothing
about whether the choices were good. The game has no notion of a right or wrong victim,
per ADR 0002, and the ending must not smuggle one back in.

**Blocked by:** 12 — Absence timer and Reassignment. 13 — Body discovery and the Hunter.

**Status:** done

- [x] The player can leave town at any moment, ending the run
- [x] Starving, burning at dawn and being caught by a Hunter each end the run
- [x] The Epilogue shows one slide per Townsperson with their end state and what followed
- [x] The Epilogue shows one summary slide for the town
- [x] Slides are data readouts with no evaluation, score or judgement of the player
- [x] The Epilogue is reachable from any point in a run for testing
