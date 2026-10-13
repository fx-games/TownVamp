# Spec: Town Vampire — Proof of Concept

Status: ready-for-agent

Vocabulary in this document is defined in `CONTEXT.md`. Capitalised terms — Townsperson,
Obligation, Routine, Schedule, Presence, Blood, Zone, Hunter, Withdrawal, Epilogue — are
glossary terms and are used precisely. Decisions recorded in `docs/adr/0001`, `0002` and
`0003` are binding on this spec.

## Problem Statement

The repository holds a working turn-based roguelike RPG: dungeon crawling, loot,
classes, spells, quests. It is a game that has been made many times, and the parts of it
that are genuinely hard and genuinely good — field of view with light occlusion,
pathfinding, zone-based map semantics, tile rendering, LDtk map loading, a plaintext
authoring format — are buried under content and combat systems that are not interesting.

The world only moves when the player moves. Every NPC is either idle and random-walking
or attacking the player. Nothing in the town has a life of its own, so there is nothing
to observe and nothing that could be disturbed. A player cannot form a hypothesis about
a person, because people do not have anything to be wrong about.

What is wanted instead is a game where a town is a running clockwork of people with
duties to each other, and the player is a vampire who must feed on it while remaining
unseen — and where every death visibly damages the clockwork, in ways that were the
player's to anticipate and are the player's to live with.

## Solution

A real-time town simulation built on the existing engine, with the RPG systems removed.

Nine Townspeople live in a town. Each has a Routine — the fixed hours of their day, sleep
and meals and home — and one to three Obligations, each a duty owed either to another
Townsperson or to the town. Their Schedule is derived: the Routine, with the Obligations
resolved into whatever hours the Routine leaves free. Time advances on its own whether
or not the player acts.

Some of what a Townsperson does is a Private Action, carrying a Witness Whitelist of the
people permitted to see it. Anyone else in view causes them to break off and seek
privacy. The game keeps no record of what the player has learned — there is no dossier
and no journal, and the engine has no concept of spying. What the player understands is
what they were present to see and chose to make of it.

The player must Kill to refill Blood, which drains continuously. A killed Townsperson
becomes Absent; after one Day their Obligations are offered for Reassignment — first to
whoever the Obligation pointed at, then to anyone whose Role suits it, and otherwise
they go Unmet and stay that way. A taker needs free hours and nothing more. So the town
absorbs early deaths through relationships and slack, and once the survivors run out of
hours it begins failing everything at once.

Being seen where he does not belong costs the vampire: that Townsperson Withdraws,
spending more of their day indoors and out of reach. A body that is found makes the
death Confirmed, and a Hunter begins walking the streets at night. Dawn kills. When the
run ends — the player leaves, or starves, or is caught by sunlight or a Hunter — an
Epilogue shows one slide per Townsperson and one for the town.

## User Stories

### Time and presence

1. As a player, I want time to advance whether or not I act, so that the town feels like
   a place that exists rather than a puzzle waiting for my input.
2. As a player, I want Townspeople to move on their own schedules, so that I can learn
   where someone will be by watching where they have been.
3. As a player, I want to pause the game at any moment, so that I can think about what I
   have just seen without losing the town.
4. As a player, I want to wait until a named moment — dawn, dusk, or until a particular
   Townsperson moves — so that I do not have to sit through hours where nothing I care
   about happens.
5. As a player, I want a wait to be interrupted when something notable happens, so that
   skipping time never costs me the thing I was waiting for.
6. As a player, I want to express a wait in terms of a Townsperson's movement, so that
   the act of waiting is itself a statement of what I believe about their Schedule.
7. As a player, I want a configurable Time Multiplier, so that I can run a whole day in
   six minutes while learning the game and slow it down when I want to watch closely.
8. As a player, I want the whole simulation to scale with the Time Multiplier — people
   walking as well as the calendar — so that the town behaves identically at every
   setting and only the pace differs.
9. As a player, I want to see the current time of day, so that I can judge how long I
   have before dawn.

### Observing the town

10. As a player, I want Townspeople to perform visible acts rather than emit messages,
    so that what I learn comes from watching rather than from being told.
11. As a player, I want those acts to be ambiguous — a door opened, an object carried, a
    thing used — so that interpreting them is my work and not the game's.
12. As a player, I want to overhear conversations between Townspeople when I am close
    enough, so that proximity is worth risking.
13. As a player, I want no journal, dossier or fact list, so that the game never tells me
    what I have figured out and never implies there is a checklist to complete.
14. As a player, I want a Townsperson's Routine to be easy to read from a glance at the
    street, so that the ordinary rhythm of the town is free and my attention can go to
    what is unusual.
15. As a player, I want a Townsperson's Obligations to be visible only by being present
    when they perform them, so that understanding someone costs me time I could have
    spent elsewhere.
16. As a player, I want several interesting things to be happening in different places at
    once, so that choosing where to stand is a real decision.

### Secrecy

17. As a player, I want some acts to be performed only when the wrong people cannot see,
    so that there is something worth catching people at.
18. As a player, I want a Townsperson to break off a Private Action and move away when
    they notice me watching, so that observing a secret requires me to be hidden and not
    merely nearby.
19. As a player, I want a Townsperson to also break off if someone walks in partway
    through, so that the rule is about being observed and not about how the act started.
20. As a player, I want some secrets to be shared between several Townspeople, so that
    watching who is permitted to be present teaches me about their relationships.
21. As a Townsperson, I want to seek somewhere unobserved when interrupted, so that a
    secret postponed is not a secret abandoned.
22. As a player, I want to be able to see out of darkness while a Townsperson holding a
    lamp cannot see into it, so that staying unlit is how I watch without being watched.

### Places and trespass

23. As a player, I want the town divided into places I am free to be and places I am not,
    so that moving through it at night is a decision rather than a formality.
24. As a player, I want a Townsperson who sees me somewhere I do not belong to flee, so
    that being careless has a consequence I can observe immediately.
25. As a player, I want that Townsperson to afterwards spend more of their day indoors,
    so that carelessness costs me access to them for the rest of the run.
26. As a player, I want no alarm to be raised by ordinary trespass, so that the town's
    response is a slow tightening rather than a sudden loud event.
27. As a player, I want doors, shutters and windows to matter — some blocking movement
    but not sight — so that there are ways to observe a room without entering it.

### Feeding and consequence

28. As a player, I want Blood to drain continuously, so that watching indefinitely is not
    an option and I am eventually forced to choose someone.
29. As a player, I want Killing to be the only way to refill Blood, so that every meal is
    a decision with permanent consequences.
30. As a player, I want to see how much Blood I have left, so that I can judge how many
    nights of observation I can still afford.
31. As a player, I want a killed Townsperson to become Absent rather than immediately
    known dead, so that there is a window in which nobody understands what has happened.
32. As a player, I want a body to be discoverable by any Townsperson whose line of sight
    crosses it, so that where I leave a body is a real choice.
33. As a player, I want to be able to hide a body, so that I can convert a killing into a
    disappearance.
34. As a player, I want Absence to be indistinguishable from having merely gone away, so
    that the town's uncertainty works in my favour.
35. As a player, I want a Townsperson's duties to be picked up by others a day after they
    vanish, so that the town visibly tries to repair itself.
36. As a player, I want duties to fall first to the people they were owed to, so that the
    repair reflects relationships rather than an org chart.
37. As a player, I want a duty nobody has time for to simply stop being done, so that the
    town's failure is quiet and permanent rather than announced.
38. As a player, I want to see a duty going undone — the unlit lamp, the stall nobody
    opens, the person waiting at a door — so that the damage I have caused is on screen
    rather than in a ledger.
39. As a player, I want killing a Townsperson with a light Routine to hurt the town more
    than killing a busy one, so that there is something genuinely non-obvious for
    observation to teach me.
40. As a player, I want each successive death to be absorbed less well than the last, so
    that the town's decline accelerates without the game announcing a difficulty change.

### Threats

41. As a player, I want to die if I am caught outdoors at dawn, so that every night has a
    deadline and range is a resource.
42. As a player, I want to be able to judge how far from shelter I can safely go, so that
    the deadline is something I can plan around rather than be ambushed by.
43. As a player, I want a Hunter to appear once a death is Confirmed, so that leaving a
    body where it will be found has a cost beyond bookkeeping.
44. As a player, I want a Hunter to pursue me rather than flee, so that there is one kind
    of person in the town I must genuinely avoid.
45. As a player, I want being caught by a Hunter to kill me outright with no fight, so
    that the danger is about evasion and not about combat statistics.
46. As a player, I want Hunters to carry light, so that the same rule that lets them see
    also blinds them to where I am standing.
47. As a player, I want enough Withdrawal across the town to make feeding genuinely hard,
    so that sloppiness can starve me without anyone ever raising a hand against me.

### Ending a run

48. As a player, I want to leave town whenever I choose, so that I can decide when my
    story here is finished.
49. As a player, I want the run to end if I starve, burn, or am caught, so that failure
    is an ending rather than a reload.
50. As a player, I want an Epilogue showing what became of each Townsperson, so that I
    can see the shape of what I did.
51. As a player, I want a summary of the town as a whole, so that I can see the aggregate
    of nine individual outcomes.
52. As a player, I want the Epilogue to describe consequences rather than score me, so
    that the game never tells me whether I chose correctly.

### Authoring

53. As a designer, I want to define a Townsperson's Role, Identity, Routine and
    Obligations in the existing plaintext record format, so that I can add people without
    writing code.
54. As a designer, I want to author what an Obligation looks like when performed, so that
    the player has something to witness.
55. As a designer, I want to author what an Obligation looks like when it is not
    performed, so that its absence is visible on screen.
56. As a designer, I want to mark an action as Private and give it a Witness Whitelist,
    so that secrets are data rather than code.
57. As a designer, I want to define Zones and who belongs in them on the existing map
    format, so that trespass rules come from the map.
58. As a designer, I want to tune Blood drain, the value of an AUT in seconds, and
    Routine density without rebuilding, so that I can find these numbers by playing.

### Removed and hidden systems

59. As a developer, I want the dungeon generator, combat, magic, status effects,
    targeting and loot systems deleted, so that the codebase reflects the game being
    made.
60. As a developer, I want inventory, item and vendor code kept and compiling but
    unreachable from the UI, so that it can be brought back without being rewritten.
61. As a player, I want no inventory, equipment or shop interface, so that the game
    presents only what it is currently about.

## Implementation Decisions

### Clock

The AUT/energy model is kept exactly as it is and only its driver changes, per ADR 0001.
`TimeKeeper`'s player-turn entry point becomes a duration-advancing call made from the
fixed-timestep update rather than from the end-turn action, and the player becomes an
ordinary timed being with its own energy budget. Per-actor speeds, action costs and
status ticking survive untouched. Movement stays grid-discrete; smooth motion remains
presentation, handled by the existing movement animation.

Both existing loops already provide a fixed timestep, so no new loop is introduced.

At the default Time Multiplier one Day is twenty-four minutes of real time, which makes
one render tick equal one in-game second. Multipliers of x2 and x4 shorten the Day to
twelve and six minutes. **The multiplier scales the entire simulation, including walking
speed, and movement animation durations scale with it.** A multiplier that scaled only
the calendar would make Schedules physically infeasible at higher settings, producing
defects that reproduce at no other setting.

The relationship between an AUT and a wall-clock second is a tuning value, exposed as
configuration alongside the Time Multiplier.

The broker's queued-action gate currently drains only when zero animations are running
anywhere in the world. With a town animating continuously this never becomes true and
the queue deadlocks. The gate becomes per-actor, or is removed. This is a prerequisite
for anything else in this spec working at all.

### Time control

The player can pause, which simply stops advancing the clock. The player can wait until a
condition — dawn, dusk, or a named Townsperson changing position — which advances the
simulation without rendering and aborts the moment the condition is met or something
notable occurs. There are no continuous speed controls at runtime; the Time Multiplier is
configuration, fixed for a session.

### Townspeople

A Townsperson carries a Role, an Identity, a Routine, a set of Obligations, and a
Presence of Alive, Absent or Confirmed Dead.

Role is a closed set of five: Farmer, who produces food; Smith, who makes and repairs
metal tools; Builder, covering carpenter, mason and wheelwright; Trader, covering
merchant, innkeeper and transport; and Authority, covering lord, priest, guard and judge.
Because Role is what Reassignment falls back to once relationship has been tried, the
closed set doubles as the compatibility rule — an Obligation can pass to a Townsperson of
a fitting Role and to no one else. At least four of the five must be represented so the
fallback has somewhere to go, and the Roles the town cannot do without want more than one
holder — a Role with a single holder loses its duties permanently the moment that person
does.

The town is built in the existing town map, reusing its named inhabitants as Townspeople
with their vendor and quest identities dropped. It currently holds nine: four Farmers who
between them bring in vegetables, grain, milk, eggs, flour and fish; two Traders, one
cooking at the inn and one selling at a market stall; and one each of Smith, Builder and
Authority. Food has four producers and two destinations on purpose — it is the one thing
the town cannot survive losing, so it is the one thing with redundancy in it.

The Routine is authored and fixed: sleep, meals, home. It never reacts to anything, with
the single exception of Withdrawal. The Schedule is derived — the Routine, with
Obligations resolved into the hours the Routine leaves free — and is recomputed from the
Obligation set rather than authored. Recomputation happens at the Day boundary, not
immediately on change, so that a Townsperson never abandons what they are doing
mid-street.

Because Schedules are derived, they must be checked for physical feasibility: the walk
between consecutive slots has to fit the hours allowed. Infeasible Schedules are a
loading-time and recomputation-time concern, not a runtime surprise.

The two-state actor model of the current AI — idle and attacking — is deleted outright.
Behaviour comes from the Schedule.

### Obligations and Reassignment

Per ADR 0003, an Obligation points either at another Townsperson or at the town, and is
the only thing the simulation propagates. Each Obligation is authored twice: what it
looks like performed, and what its absence looks like on screen.

A Townsperson who has been Absent for one full Day has their Obligations offered for
Reassignment. Candidates are tried in order: whoever the Obligation pointed at, then
anyone whose Role suits it. A candidate must be free during the Obligation's own hours;
that is the only requirement and the only cost. There is no displacement of a taker's
existing duties and no quality degradation. If no candidate has capacity, the Obligation
goes Unmet permanently.

Free *during those hours*, not merely free somewhere in the day: a candidate with six
idle hours still cannot be in two places at ten in the morning. Counting capacity in
aggregate makes a town look resilient and leaves it unable to cover anything, because a
duty handed to someone whose Routine already claims its window is dropped from their day
and then owed by one person and done by nobody.

The town's resilience therefore emerges from Routine density rather than from a tuned
difficulty curve, and the collapse accelerates on its own as spare hours are consumed.

### Actions and secrecy

An action is Public or Private. A Private Action carries a Witness Whitelist naming the
Townspeople permitted to observe it. An unwhitelisted observer causes the actor to
abandon the action and move somewhere unobserved if such a place is reachable.

The perception query is deliberately narrow: **a single "who can currently see me" test,
evaluated only for Townspeople attempting or performing a Private Action, and for
Hunters.** No general NPC vision is built. Cost is bounded by the number of Private
Actions in flight rather than by population, and the check runs per tick while an action
is in progress so that walking in partway aborts it.

The existing field-of-view already blocks a viewer standing in a lit tile from seeing into
unlit tiles. This is the asymmetry the design depends on and is used as-is. Tile
walkability and transparency are already independent, so windows and shutters need no
new tile concept.

### Player knowledge

Per ADR 0002, nothing is modelled. No dossier, no journal, no fact database, no
discovered-state. Nothing may be gated on the player knowing something. NPC actions are
authored to be abstract and open to interpretation.

### Zones and trespass

Zone permission reuses the existing map zone types, which are already loaded from the map
format and currently unused. A Zone's permitted occupants are the same predicate as a
Private Action's Witness Whitelist — *may this person be here, seeing this?* — applied to
a place rather than an act, and are authored the same way.

A Townsperson who sees the vampire in a Zone where he does not belong flees and postpones
any Private Action. They do not challenge and do not raise an alarm. They then Withdraw:
their Routine biases toward private Zones and indoor hours for the remainder of the run.
Withdrawal is a Routine bias, not a new system, and is the only thing that modifies a
Routine.

### The vampire

Blood drains continuously and is refilled only by Killing. Killing sets the victim's
Presence to Absent and leaves a body. A body may be hidden. A body whose tile falls in any
Townsperson's line of sight promotes the victim to Confirmed Dead. Body discovery is
tested per corpse rather than per observer, since corpses are few.

Sunlight kills a vampire outdoors at dawn. The existing daylight and lighting state is the
basis for this.

### Hunters

A Hunter appears once any death is Confirmed and walks the town at night. A Hunter is the
sole exception to the flee-on-sight rule: it pursues. Catching the vampire ends the run
immediately. **There is no combat: no health, no damage, no armour, no attack rolls.** A
Hunter carries light, and is therefore subject to the same occlusion rule that blinds a
lit viewer to unlit tiles. Patrol behaviour uses the existing named-path support in the
map format.

### Ending and Epilogue

A run ends when the player leaves town, or when the vampire dies — starved, burned, or
caught. The Epilogue is one slide per Townsperson giving their end state and what followed
from it, plus one slide summarising the town. For the proof of concept these are raw data
readouts, not prose; the simulation generates the story and the Epilogue only reports it.
No score, no evaluation, no indication of whether choices were good.

### Authoring format

Townspeople, Routines, Obligations, and Witness Whitelists are authored in the existing
plaintext record format already used for game data, with its existing conversion tooling.
No new authoring format is introduced. Zones and named paths are authored on the map in
the existing map format.

### Scope of demolition

Deleted: the dungeon generator, the standalone combat simulator, melee, ranged, spell and
knockback actions, magic definitions, status effects, hit zones, targeting and targeted
and untargeted action bases, weapon definitions, combat hooks, the existing AI action
selector, unique items, quest dialogue, and the RPG-specific UI panels — targeting,
status effects, enchantment transfer. Roughly six and a half thousand lines.

Retained but unreachable: inventory, item, equipment, pickup, drop, item-move and item
hook code, the shop interface, vendor definitions, and the vendor branches of the talk
action. Approximately two thousand lines, kept compiling, with keybindings and menu
entries removed. This is deliberate mothballing, not deletion.

Retained and untouched: the map module including zones, field of view and pathfinding;
the geometry module; the tile renderer; audio; map loading; utilities; the record format;
animation and the broker's hook mechanism; doors; and the majority of the UI shell.

## Testing Decisions

### What makes a good test here

A good test drives the simulation the way the game does and asserts on what a player could
observe. It advances time and then asks the questions a player would ask by looking: where
is this Townsperson, is this Obligation being performed, who is Absent, how much Blood is
left. It never reaches into how a Schedule was computed, how Reassignment enumerated its
candidates, or what intermediate state a Townsperson passed through.

Tests are deterministic. The simulation is seeded, and any test that needs a specific
outcome fixes the seed rather than asserting on a distribution.

### The seam

**There is one seam: the game instance, constructed with the null UI, advanced by a
duration, asserted against world state.**

This is the highest available point and it already exists in shape — the game is already
constructed with a null UI implementation, and a console mode already runs it without the
graphical layer. Headless operation is a supported configuration rather than something
introduced for testing.

A test loads a town definition, advances the clock by some number of in-game hours, and
asserts on observable state. Everything worth testing crosses this seam:

- A Schedule derived from a Routine and a set of Obligations puts a Townsperson in the
  right place at the right hour.
- A derived Schedule is physically walkable in the time allowed.
- A Killed Townsperson is Absent, and remains Absent while their body stays hidden.
- Reassignment does not fire before a full Day of Absence, and does fire after.
- Reassignment prefers the Townsperson the Obligation pointed at, falls back to Role, and
  fails when no candidate has free hours.
- A failed Reassignment leaves the Obligation Unmet permanently.
- Successive deaths are absorbed less well as spare capacity is consumed.
- Blood drains at the configured rate and is restored by Killing.
- A body in line of sight promotes the victim to Confirmed Dead; a hidden one does not.
- A Confirmed death produces a Hunter.
- A Townsperson who sees the vampire trespassing flees and thereafter spends more hours
  indoors.
- A Private Action is abandoned when an unwhitelisted Townsperson can see it, including
  when they arrive partway through.
- The simulation produces identical state at every Time Multiplier for the same number of
  in-game hours.

That last one is worth calling out: it is the cheapest possible guard against the
calendar-versus-simulation scaling defect, and it is a single test.

Not tested at this seam, and not tested at all: rendering, audio, animation timing, input
handling, and the epilogue's presentation.

### Prior art

**There is none. The repository contains zero test files across roughly thirty-five
thousand lines.** The first test written for this feature establishes the pattern, and it
should be written before the town is authored rather than after — a single Townsperson
with a Routine and one Obligation, advanced through a Day, is the cheapest possible proof
that the seam is usable.

The standard Go testing package is sufficient. No test framework, no fixtures beyond a
small town definition in the existing record format, no mocking library.

## Out of Scope

**Feed.** Biting without killing, and the confusion that makes a victim forget an
Obligation or Routine slot, is deliberately deferred. Killing is the only way to refill
Blood in the proof of concept. The reasoning is preserved in the glossary because the
mechanic is interesting and non-obvious — a forgotten Obligation produces no Absence, so
nothing triggers Reassignment and nobody registers the gap.

**Prose in the Epilogue.** Slides are raw data readouts. Generated or authored prose,
including any use of language models to write it, comes later.

**A knowledge model of any kind**, per ADR 0002. Not a journal, not a hint system, not
objectives, not dialogue gated on what the player has seen.

**Combat.** Not for Hunters, not for anyone. No health, damage, armour or attack rolls
anywhere in the game.

**Any alarm or suspicion mechanism.** Trespass causes flight and Withdrawal. Nothing
escalates except through Confirmed death producing a Hunter.

**Displacement or degradation in Reassignment.** A taker with free hours takes the duty in
full; a taker without free hours does not take it at all.

**Continuous runtime speed controls.** The Time Multiplier is configuration, set for a
session.

**Growing the town much further.** Nine Townspeople. The town went from six to nine to
give the economy enough producers that food has redundancy in it and the market has
somebody to run it; the three who were added carry Obligations but no authored secret
yet, which is the one thing about them still owed. What bounds the town from here is the
authoring cost of the second face of every Obligation, not engine capacity, and the
per-Townsperson Epilogue slide approach still stops scaling somewhere well short of
thirty.

**Inventory, items, equipment and vending as player-facing features.** The code stays and
compiles; the interface does not exist.

**Saving and loading a run.**

## Further Notes

The three numbers this design cannot supply are Blood drain rate against how much one
Townsperson restores, the value of an AUT in seconds, and how much time a Routine leaves
free. The first sets run length, the second sets walking speed and therefore Schedule
feasibility, and the third sets how many deaths the town absorbs before it starts failing
everything at once. All three want playtesting rather than reasoning, and all three should
remain adjustable without a rebuild.

Two risks are worth naming plainly.

The first is Schedule feasibility interacting with the Time Multiplier. If the multiplier
were ever to scale the calendar without scaling walking speed, Schedules would become
unwalkable at higher settings only, producing defects invisible at the default. The
decision above forecloses this, and the identical-state-across-multipliers test guards it.

The second has no mitigation and should be confronted early. ADR 0002 removes every
fallback for illegible NPC behaviour: if what Townspeople do does not read, there is no
journal, hint or objective list to compensate, and the player has nothing to reason about.
The cheapest test of the entire premise is one Townsperson with three Obligations and both
authored faces, played for a few Days, before the other five are written.

The best news from surveying the codebase is how much of what this design needs is already
built and unused. The field-of-view code already blinds a lamp-carrying viewer to the
dark. The map format already carries public, private and high-security zones, named
regions and named paths. The record format already has round-trip tooling. These came from
an earlier stealth-game project the engine borrowed from, and they have been sitting inside
an RPG that had no use for them.
