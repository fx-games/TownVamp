# 2. Player knowledge is not modelled

Date: 2026-08-02

## Status

Accepted

## Context

The game is about a vampire observing a town in order to choose victims. The obvious
implementation is a knowledge model: facts about each Townsperson, discovered by
observation, surfaced in a dossier or journal, with the UI showing what is known and
what is still hidden.

Every design in this space tends toward that. It is also where such designs die: the
fact schema grows, the discovery rules multiply, the UI becomes a checklist, and
"spying" degenerates into farming proximity until the panel fills in.

## Decision

The engine models nothing about what the player knows. No dossier, no journal, no fact
database, no discovered/undiscovered state. The engine has no concept of spying at all.

Secrecy is modelled entirely on the NPC side. Actions are Public or Private; a Private
Action carries a Witness Whitelist naming who may observe it. An NPC who notices an
unwhitelisted observer refrains and seeks privacy. NPC actions are deliberately
abstract — opening a door, carrying an object, using a thing, talking within earshot —
and are open to interpretation.

Inference happens in the player's head, not in the program.

## Consequences

The entire class of knowledge-tracking code, and its UI, never gets written. What the
player learns is bounded by what they were physically present to see, with no
bookkeeping.

Nothing can be gated on the player "knowing" something — no dialogue options unlocked
by discovered facts, no objective list, no hint system, no quest log keyed to
knowledge. Any such feature would require reversing this decision.

The design burden moves onto NPC behaviour. If NPCs are not legible enough to support
inference, the player has nothing to reason about and there is no fallback: no journal
will paper over unreadable behaviour. Legibility of Private Actions and their tells
becomes the thing that must be gotten right.

It also means the game cannot tell the player whether a decision was good. It does not
know. This is consistent with the wider intent — the simulation has no notion of a
right or wrong victim, only of consequences that propagate.
