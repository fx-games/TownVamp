# 15 — Author the town's Obligations

**What to build:** the content pass that turns six people with Routines into a society
worth watching and worth damaging. Every Townsperson gets one to three Obligations, each
authored with both faces, plus the Private Actions that make them worth spying on.

Every mechanism this needs already exists by now. This ticket is writing, placing and
tuning — and it is where the proof of concept either becomes legible or does not.

Author for consequence, not for flavour. Obligations should point at each other in ways
that make Reassignment interesting: someone whose duty is owed to a specific person, so
that person can take it up themselves; someone whose duty only a matching Role can
inherit; at least one duty that nobody can absorb once its holder is gone, so the player
can watch something fail permanently.

Spread the load unevenly across the six. The Townsperson with the lightest Routine is the
town's shock absorber, and killing them should quietly make everything else fall apart two
days later. That is the deepest thing observation can teach the player and it has to be
authored in deliberately — it does not arise by accident.

Private Actions want the same care. Give at least one secret shared between two
Townspeople so its Witness Whitelist has more than one name, and keep every Private Action
abstract enough to be interpreted rather than read: a door, an object carried, a thing
used, a conversation within earshot.

**Do this incrementally and play between passes.** Author one Townsperson fully, play
several Days, and confirm both that their Obligations read when performed and that their
absence reads when they are gone, before writing the other five. Per ADR 0002 there is no
journal or hint system to fall back on if behaviour does not read.

**Blocked by:** 08 — Obligations and derived Schedule. 09 — Private Actions and seeking
privacy. 10 — Zones, trespass and Withdrawal. 12 — Absence timer and Reassignment.

**Status:** ready-for-agent

- [ ] Each of the six Townspeople has one to three Obligations authored in the record
      format
- [ ] Every Obligation has both a performed appearance and an absent appearance
- [ ] Some Obligations point at a specific Townsperson and some at the town
- [ ] At least one Obligation can be reassigned to the person it pointed at, and at least
      one only to a matching Role
- [ ] At least one Obligation cannot be absorbed by anyone once its holder is gone
- [ ] At least one Private Action is shared between two Townspeople
- [ ] Routine density varies enough that one Townsperson is clearly the town's shock
      absorber
- [ ] Private Zones are authored across the map so trespass is possible and meaningful
- [ ] A full run has been played end to end and the consequences of a killing were legible
      without recourse to code or logs
