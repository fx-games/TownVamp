# Context

Glossary for the town simulation. Terms only — no implementation detail.

The proof of concept town holds six Townspeople, each carrying one to three Obligations.
Every Obligation is authored twice over: what it looks like being performed, and what
its absence looks like.

## Core

**Townsperson** — a simulated inhabitant of the town. Has a Role, an Identity, a set of
Obligations, and a Schedule.

**Role** — what a Townsperson does *for the town*. Public and structural: visible from
watching them work, and other people depend on it. A Role is drawn from a closed set of
five:

- **Farmer** — produces food.
- **Smith** — makes and repairs metal tools.
- **Builder** — carpenter, mason, wheelwright.
- **Trader** — merchant, innkeeper, transport.
- **Authority** — lord, priest, guard, judge.

Role is what Reassignment falls back to once relationship has been tried, so the closed
set is also the compatibility rule: an Obligation can pass to a Townsperson of a fitting
Role and to no one else.

**Identity** — who a Townsperson *is*, beneath the Role. Secrets, inner life, private
relationships, what they want and fear. Not visible from the Role, and never stated by
the game — only inferable from what they can be caught doing.

**Obligation** — a duty a Townsperson owes, pointing either at another Townsperson
("feeds Kessa's children") or at the town itself ("opens the gate at dawn").
Obligations are the substrate along which a death propagates: when a Townsperson dies,
their Obligations are voided. A voided Obligation is offered for Reassignment; if no
one takes it, it goes Unmet and stays that way.

**Reassignment** — the attempt to move a voided Obligation to another Townsperson.
Candidates are tried in order: first whoever the Obligation pointed at (the widow being
fed can fetch her own bread), then anyone whose Role suits it, and if neither yields a
taker the Obligation goes Unmet. Relationship before competence — which is why early
deaths get absorbed and later ones cannot be.

**Unmet Obligation** — an Obligation that was voided and could not be reassigned. The
unit of consequence in the simulation, and the raw material of the epilogue.

## Presence

A Townsperson is in exactly one of three states:

**Alive** — present and meeting their Obligations.

**Absent** — not where they should be. Covers the murdered, the vanished, and the
merely away; the town cannot tell these apart. After one full day of Absence, the
Townsperson's Obligations are offered for Reassignment.

**Confirmed Dead** — the town has proof of death.

## Time and behaviour

**Routine** — the fixed, authored part of a Townsperson's day: sleep, meals, home. Never
reacts to anything, and publicly legible — a glance at the street shows it.

The hours a Routine leaves free are that Townsperson's **spare capacity**, and capacity
is the only thing limiting Reassignment: a taker who has no free hours cannot take the
Obligation, and it goes Unmet. So a Townsperson who sleeps long and eats slowly is
fragile, and a Townsperson with a lean Routine is the town's shock absorber. Killing the
shock absorber is what makes everything else fall apart two days later — which is the
deepest thing observation can teach the player.

**Schedule** — a Townsperson's actual day: their Routine, with their Obligations
resolved into the hours the Routine leaves free. Recomputed from the Obligation set,
so a Reassignment changes behaviour without bespoke code. What the vampire learns by
*watching* is the Obligation part, not the Routine part.

## Actions and secrecy

**Public Action** — something a Townsperson does without caring who sees. Opening a
shop, drawing water, arguing in the square.

**Private Action** — something a Townsperson does only when the wrong people cannot see
them. Carries a **Witness Whitelist**: the Townspeople who are permitted to observe it.
A group secret is simply a Private Action whose whitelist has several names on it.

**Witness Whitelist** — the set of Townspeople allowed to observe a given Private
Action. Anyone else observing it is an intruder.

**Seeking Privacy** — what a Townsperson does when they notice someone outside the
Witness Whitelist can see them: they refrain from the Private Action and move somewhere
unobserved, if such a place is reachable.

## The vampire

**Blood** — the vampire's sustaining resource. Drains continuously and never stops. The
hunger clock, and the only thing forcing the player to act.

**Kill** — drain a Townsperson entirely, and the only way to refill Blood. The victim
becomes Absent and their Obligations enter the Reassignment path. Leaves a body, which
may be found unless hidden.

**Day** — twenty-four minutes of real time at the default Time Multiplier. Higher
multipliers shorten it (x2 = twelve minutes, x4 = six).

**Sunlight** — fatal. A vampire caught outdoors at dawn burns. The hard deadline on
every night's plan: how far from shelter is it safe to go, and can you get back.

## Threats

**Withdrawal** — how an ordinary Townsperson responds to having seen the vampire where
he did not belong. Their Routine shifts indoors: more time in private Zones, shutters
closed. No alarm is raised and nothing is reported — they simply become harder to reach.
Enough Withdrawal and the town starves the vampire out without ever knowing why.

**Hunter** — someone who walks the town at night once a death has been Confirmed. Unlike
an ordinary Townsperson, a Hunter who sees the vampire pursues rather than flees, and
catching him kills him. There is no fight: no health, no damage, no armour. Hunters
carry light, which by the same rule that protects them from the dark prevents them from
seeing into it.

## Places

**Zone** — a region of the map with a permission rule attached. Public zones are open to
everyone. Private zones belong to particular Townspeople. A Zone's permitted occupants
are the same kind of rule as a Private Action's Witness Whitelist: *may this person be
here, seeing this?* — asked of a place in one case and of an act in the other.

**Trespass** — being seen in a Zone where one does not belong. The Townsperson who
notices runs away, and postpones any Private Action they were about to perform. They do
not challenge the intruder and do not raise an alarm.

## Player knowledge

The engine has **no model of what the player knows.** There is no dossier, no journal,
no fact database, no unlocked-clue state. NPCs perform actions that are abstract and
open to interpretation — opening a door, carrying something, using an object, talking
where they can be overheard. All inference happens in the player's head.

Consequently the engine also has no concept of "spying." Spying is what the player is
doing; it is not something the simulation represents.

## Ending a run

**Leaving Town** — the player may quit the town at any moment. Ends the run. A run also
ends when the vampire dies: starved of Blood, caught by Sunlight, or caught by a Hunter.

**Epilogue** — what is shown once a run ends. One slide per Townsperson, stating their
end state and what followed from it, plus one slide summarising the town as a whole. In
the proof of concept these are raw data readouts rather than prose; the simulation is
what generates the story, and the Epilogue only reports it.

## Deferred

**Feed** — biting a Townsperson without killing them, refilling a little Blood and
leaving the victim confused enough to forget one Obligation or Routine slot. Cut from
the proof of concept. Worth knowing why it was interesting: a forgotten Obligation
produces no Absence, so nothing triggers Reassignment and nobody registers the gap —
making Feeding a quieter and potentially faster way to hollow out a town than Killing.

## Numbers still to be found

These are tuning, not design. They want playtesting rather than reasoning, and each
should stay adjustable without a rebuild.

- The rate Blood drains, and how much one Townsperson restores. Together these set the
  length of a run and how many of the six die in it.
- What an AUT is worth in seconds. This sets walking speed, and walking speed decides
  whether a Schedule is physically achievable in the hours allowed.
- How much time a Routine leaves free, which decides how many voided Obligations the
  town can absorb before it begins failing them.
