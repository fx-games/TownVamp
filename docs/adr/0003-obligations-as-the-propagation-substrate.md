# 3. Obligations as the propagation substrate

Date: 2026-08-02

## Status

Accepted

## Context

The game is a story generator. It has no notion of a correct or incorrect victim; it
only has consequences. The requirement is that every death has an impact on the town.

For that to be true rather than aspirational, a death needs something to propagate
along. Two obvious substrates were available.

A **social graph**: typed edges between people — spouse, rival, creditor, employer — with
a death sending a shock along the edges and changing each neighbour's disposition.

**Town functions**: services the town needs — food, order, faith, trade — each fed by
some Role, with a death degrading a service and shifting behaviour through town-level
state.

Both are common, both work, and running both means tuning two systems that can
contradict one another.

## Decision

Neither. A Townsperson is a bundle of **Obligations**, each pointing at another
Townsperson ("feeds Kessa's children") or at the town ("opens the gate at dawn"). The
simulation propagates exactly one thing: an Obligation nobody is meeting.

A Schedule is derived — a Townsperson's authored Routine, with their Obligations
resolved into the hours the Routine leaves free. When a Townsperson has been Absent for
a day their Obligations are offered for Reassignment: first to whoever the Obligation
pointed at, then to anyone whose Role suits it, and otherwise the Obligation goes Unmet.
A taker needs the Obligation's own hours free — not merely free hours somewhere in the
day, which is a different and much weaker thing; there is no other cost and no
displacement of their existing duties.

## Consequences

The two substrates above are not lost, they are unified — an Obligation aimed at a
person is the social graph, and one aimed at the town is the town-function model. One
mechanism, one place to tune.

Because Schedules are derived rather than authored, a Reassignment changes what the town
visibly does the next day with no bespoke code per Role. This is the property the whole
design rests on. It also means Schedules must be checked for physical feasibility —
whether the walk from one slot to the next fits the hours available — which is new work
that a hand-authored schedule would not have needed.

Spare capacity becomes the town's resilience, without anyone designing it that way. The
town absorbs early deaths through relationships and slack, and once the survivors run
out of free hours it begins failing everything at once. The difficulty curve is emergent
rather than tuned. The corollary is that a Townsperson's Routine — the part of them that
does nothing and reacts to nothing — is the most consequential number attached to them.

Each Obligation must be authored twice: what performing it looks like, and what its
absence looks like. The second is the one that gets skipped under pressure, and skipping
it makes the simulation invisible — the town would degrade with nothing on screen to
show it. That authoring cost is the real constraint on town size, not the engine.
