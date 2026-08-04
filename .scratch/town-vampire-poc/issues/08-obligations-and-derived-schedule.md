# 08 — Obligations and derived Schedule

**What to build:** Townspeople stop merely following a Routine and start owing each other
things. A Townsperson's day becomes their Routine with their Obligations resolved into
whatever hours the Routine leaves free — and because the Schedule is derived rather than
authored, changing someone's Obligations changes what the town visibly does, with no code
written per Role.

Per ADR 0003, an Obligation points either at another Townsperson or at the town, and is
the only thing this simulation propagates. Each Obligation is authored twice: what it
looks like being performed, and what its absence looks like. Both faces are needed here
even though nothing goes Unmet until a later ticket — the second face is the one that gets
skipped under pressure, and skipping it makes the whole simulation invisible.

Schedules recompute at the Day boundary, never mid-street, so a Townsperson does not
abandon what they are doing when their duties change.

**A derived Schedule has to be walkable.** The journey between consecutive slots must fit
the hours allowed, at the configured walking speed. This is a loading-time and
recomputation-time check, not a runtime surprise, and it is the main new cost of deriving
Schedules rather than authoring them.

**Blocked by:** 07 — Basic test town.

**Status:** done

- [x] Obligations are authored in the record format, each pointing at a Townsperson or at
      the town
- [x] Each Obligation is authored with both a performed appearance and an absent
      appearance
- [x] A Townsperson's Schedule is derived from Routine plus Obligations, not authored
- [x] Schedules recompute at the Day boundary and never mid-slot
- [x] A Schedule is validated as physically walkable in the hours allowed, and an
      infeasible one is reported at load rather than failing silently
- [x] The player can watch a Townsperson perform an Obligation and tell it apart from
      their Routine
- [x] A test asserts a Townsperson performs an authored Obligation at the expected hour
