# 4. Unclaimed hours are spent somewhere

Date: 2026-10-05

## Status

Accepted

## Context

Per ADR 0003 a Schedule is a Routine with Obligations resolved into the hours the Routine
leaves free, and the hours nothing is authored into are the town's spare capacity — the
only thing that limits Reassignment, and therefore the whole of its resilience. Those
hours have to stay unclaimed.

What a Townsperson *does* with them was never decided. The behaviour that fell out was
standing still: no Schedule entry, no destination, burn the step cost and wait. Watched
for any length of time this is the strongest thing on screen. The town moves in bursts on
the hour and is frozen in between, and a frozen town reads as a mechanism waiting for
input rather than a place that exists — which is precisely the premise the design cannot
afford to lose, because ADR 0002 leaves no journal or objective list to fall back on when
watching stops being worth doing.

## Decision

An unclaimed hour is spent going somewhere and standing about there for a while, and this
is runtime behaviour rather than authored data.

A Townsperson with no Schedule entry picks a named place they are entitled to be in, walks
to it, lingers for a varied interval, and picks again. Three constraints:

- **It must not consume capacity.** Nothing is added to the Routine or the Schedule, so
  `FreeSecondsPerDay` is unchanged and Reassignment sees exactly what it saw before.
- **It must not make anyone late.** A destination is only chosen if the walk out, the
  lingering, and the walk on to wherever they are next due all fit in the hours left.
- **It must not take anyone where they do not belong.** The predicate that governs
  trespass governs this too, which also means Withdrawal extends into these hours: someone
  who has seen the vampire where he did not belong idles at home instead.

## Consequences

The town is continuously in motion, and the places that are busy are busy for reasons —
free hours cluster around meals and evenings, so the inn and the square fill up without
anyone authoring a crowd.

Idle drift is deliberately not legible as information. It is random within a budget, so a
player who learns to read it learns nothing, and the thing worth watching for — a
Townsperson keeping to a Schedule, and the Obligations in it — stays the only signal on
screen. This is the reason it is random rather than, say, visiting friends: a second
layer of meaningful behaviour would compete with the Obligations for the player's
attention and the Obligations would lose, being rarer.

The lateness constraint costs a pathfinding query or two whenever someone picks a
destination, bounded by a cheap straight-line filter first. On this map it is belt and
braces — the walks are short relative to the windows, and removing the check does not
currently cause a missed duty. It is kept because that margin is a property of the current
map and the current Obligation windows, and neither is something the next person to author
a town should have to know about.
