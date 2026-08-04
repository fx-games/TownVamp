# 07 — Basic test town

**What to build:** a living town. Six Townspeople in the existing town map, each with a
Role and a Routine, all moving on their own while the player watches. This is the first
point at which the project looks like the game it is meant to be, and it is the substrate
every later ticket is demonstrated against.

Reuse what exists. The town map already ships with the game and already has named
locations and spawn points. The four named inhabitants already defined — kirioth, shekel,
abel and mary — are reused as Townspeople rather than replaced, with their vendor and
quest identities dropped. Two more are added to reach six.

Roles are drawn from the closed set: Farmer, Smith, Builder, Trader, Authority. Six people
across five Roles means one Role is held by two of them; which one is a content choice,
but at least four of the five must be represented so that Reassignment's Role fallback has
something to fall back to later.

Routines should differ in density on purpose. At least one Townsperson keeps a light
Routine with plenty of free hours and at least one keeps a heavy one with almost none —
that difference becomes the town's resilience once Reassignment exists, and it is the
deepest thing observation can eventually teach the player.

No Obligations yet. This ticket is Routines only.

**Blocked by:** 06 — Townsperson, Role and Routine.

**Status:** mostly-done

- [x] Six Townspeople live in the existing town map, each with a Role and a Routine
- [x] At least four of the five Roles are represented
      *all five are, with farmer held twice*
- [x] All six move on their Routines unattended while the player stands still
- [x] Routines vary in density, including at least one notably light and one notably heavy
- [x] Every Townsperson is authored in the record format, with no behaviour hard-coded
- [ ] Existing town inhabitants are reused rather than replaced, with vendor and quest
      identities removed
      *kirioth, shekel, abel and mary are gone from `maps.ldtk`; the six entities are named elias, martin, greta, otto, marta and aldren*
- [ ] A test advances several Days and asserts the town returns to a stable daily rhythm
      *no multi-Day rhythm test exists*
