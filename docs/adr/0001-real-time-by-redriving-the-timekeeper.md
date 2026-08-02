# 1. Real-time by re-driving the TimeKeeper

Date: 2026-08-02

## Status

Accepted

## Context

The engine is turn-based on roguelike rules: the world advances only when the player
acts. `TimeKeeper.PlayerTakesTurn(autUsed)` is the sole clock. It is called from
`EndTurnAction`, and it hands every other `TimedBeing` an energy allowance proportional
to the AUT the player just spent, then drains that energy by running actions.

The new game is real-time. NPCs move on their own; time advances whether or not the
player does anything.

The obvious reading of "make it real-time" is to abandon the discrete energy model and
move to continuous simulation — float positions, velocities, time-based interpolation.

## Decision

Keep the AUT/energy model exactly as it is. Change only what drives it.

`PlayerTakesTurn(autUsed)` becomes `Advance(aut)`, called from the fixed-timestep update
rather than from `EndTurnAction`. The player becomes an ordinary `TimedBeing` with its
own energy budget, subject to the same rules as everyone else.

Both existing loops already provide a fixed timestep — `util/gameloop.go` has a lag
accumulator, and ebiten's `Update` runs at a fixed 60 TPS — so no new loop is needed.

## Consequences

Everything built on the energy model survives untouched: per-actor speed stats, action
costs in AUT, status-effect ticking, the distinction between move speed and attack
speed. This is the large majority of the value in `common/`.

Movement stays grid-discrete. Actors occupy tiles, not positions. Smooth motion remains
a presentation concern handled by `MovementAnimation`, exactly as today. Anything
wanting true continuous movement — sliding, knockback with momentum, collision
between moving bodies — is off the table without revisiting this decision.

`Broker.checkForQueuedActions` must change. It currently drains the queued-action list
only when *zero* animations are running anywhere in the world. That was safe when one
actor moved at a time; with a whole town animating continuously the queue never drains
and the game deadlocks. The gate has to become per-actor, or be removed.

The relationship between AUT and wall-clock seconds is currently undefined and must be
chosen. It sets walking speed, and walking speed determines whether a Schedule derived
from Obligations is physically achievable in the hours allotted. Expect to tune this
against the real town rather than to derive it.
