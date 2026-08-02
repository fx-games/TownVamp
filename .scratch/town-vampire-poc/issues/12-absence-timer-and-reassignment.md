# 12 — Absence timer and Reassignment

**What to build:** the town notices someone has stopped showing up, tries to repair
itself, and visibly fails where it cannot. This is the payoff of the entire design — the
first ticket where a death has consequences the player can watch.

A Townsperson Absent for one full Day has their Obligations offered for Reassignment.
Candidates are tried in order: whoever the Obligation pointed at first — the widow being
fed can fetch her own bread — then anyone whose Role suits it. Relationship before
competence.

**A candidate needs free hours in their Routine and nothing else.** No displacement of
their existing duties, no priority ordering, no degraded performance. If no candidate has
capacity the Obligation goes Unmet, permanently, and nobody ever picks it up. That single
constraint is what makes the town absorb early deaths and then fail everything at once
later — the difficulty curve emerges rather than being tuned.

**The half of this ticket that must not be cut:** an Unmet Obligation has to be *visible*.
The unlit lamp, the stall nobody opens, the person waiting at a door for someone who will
not come. Ticket 08 authored that second face for every Obligation; this is where it goes
on screen. If it is skipped, the simulation degrades invisibly and the player has nothing
to read — and per ADR 0002 there is no journal or objective list to compensate.

**This is where the project's central risk is discovered.** If Unmet Obligations do not
read on screen, the premise does not work, and it is better to learn that here than after
the town is fully authored.

**Blocked by:** 08 — Obligations and derived Schedule. 11 — The vampire.

**Status:** ready-for-agent

- [ ] Reassignment does not fire before a full Day of Absence, and does fire after
- [ ] Candidates are tried pointed-at first, then Role-compatible
- [ ] A candidate without free hours cannot take an Obligation
- [ ] An Obligation nobody can take goes Unmet permanently and is never revisited
- [ ] A taker's existing Obligations are untouched
- [ ] Every Unmet Obligation is visible in the town, not merely recorded
- [ ] Successive deaths are absorbed less well as spare capacity is consumed
- [ ] Tests cover the ordering, the capacity limit, the permanence of Unmet, and the
      accelerating failure across several deaths
