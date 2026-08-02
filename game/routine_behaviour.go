package game

import (
	"fmt"
	"rpg/common"
	"rpg/geometry"
)

// getScheduledAction is what a Townsperson does with their time: go to wherever
// their Routine says they should be, and stay there once they arrive.
//
// This replaces the dungeon crawler's idle-or-attacking AI entirely. Behaviour
// comes from authored data, so changing what the town does means editing a
// record file, not writing code — which is the property Obligations and
// Reassignment will depend on in later tickets.
func (g *GameInstance) getScheduledAction(actor *common.Actor) (int, func()) {
	person := actor.GetTownsperson()
	if person == nil {
		return 0, nil // the player, and anyone else without a Routine
	}

	stepCost := 100 / actor.GetMoveSpeed()
	if actor.GetEnergy() < stepCost {
		return 0, nil // not enough time has passed for another step
	}

	slot, hasSlot := person.Schedule().EntryAt(g.state.GetClock().SecondsSinceMidnight())
	if !hasSlot {
		return waitOneStep(stepCost) // unclaimed hour: spare capacity, spent standing still
	}

	if g.hasArrivedAt(actor, slot.At) {
		g.noteObligationPerformed(actor, slot)
		return waitOneStep(stepCost)
	}

	next, found := g.nextStepTowards(actor, slot.At)
	if !found {
		return waitOneStep(stepCost)
	}

	return stepCost, func() {
		if g.state.IsClosedDoorAt(next) {
			g.actionBroker.Execute(common.NewOpenDoorAction(common.AbstractCause("Routine"), actor, next))
			return
		}
		g.actionBroker.Execute(common.NewMoveAction(common.AbstractCause("Routine"), actor, next, false))
	}
}

// hasArrivedAt decides whether a Townsperson is where their day says they should
// be. Standing on the exact tile counts; so does standing next to it when someone
// else is already there. Duties are frequently owed *to* a person — bread left at
// a door, a dispute heard at a stall — and the recipient is standing on the spot,
// so insisting on the exact tile means the duty can never be performed at all.
func (g *GameInstance) hasArrivedAt(actor *common.Actor, destination geometry.Point) bool {
	here := actor.MapPosition()
	if here == destination {
		return true
	}
	currentMap := g.state.GetMap()
	occupied := currentMap.IsActorAt(destination) || !currentMap.IsWalkable(destination)
	return occupied && currentMap.MoveDistance(here, destination) <= 1
}

// waitOneStep burns the same time a step would have taken. Without it a
// Townsperson standing at their destination spins the energy loop forever.
func waitOneStep(stepCost int) (int, func()) {
	return stepCost, func() {}
}

// nextStepTowards is one step along a path to the destination, or false when
// there is no way through.
func (g *GameInstance) nextStepTowards(actor *common.Actor, destination geometry.Point) (geometry.Point, bool) {
	currentMap := g.state.GetMap()
	here := actor.MapPosition()

	path := currentMap.GetJPSPath(here, destination, func(p geometry.Point) bool {
		return currentMap.Contains(p) && currentMap.IsWalkableFor(p, actor)
	})
	for _, step := range path {
		if step == here {
			continue
		}
		if currentMap.IsActorAt(step) {
			return geometry.Point{}, false // someone is in the way; wait rather than shove
		}
		return step, true
	}
	return geometry.Point{}, false
}

// WaitUntil runs the simulation forward without drawing until the named moment
// arrives. It stops early if the vampire dies, and is capped at two Days so a
// condition that can never hold cannot hang the game.
func (g *GameInstance) WaitUntil(condition common.WaitCondition) {
	wasPaused := g.isPaused
	g.isPaused = false
	defer func() { g.isPaused = wasPaused }()

	clock := g.state.GetClock()
	maxTicks := int(2*common.SecondsPerDay/g.state.GetTimeMultiplier()) + 1
	for i := 0; i < maxTicks; i++ {
		g.Advance(1.0)
		if condition.IsMet(clock) {
			return
		}
		if player := g.state.GetPlayer(); player != nil && !player.IsAlive() {
			return
		}
	}
}

// noteObligationPerformed shows the player what a Townsperson is doing, once,
// when they arrive to do it — and only if the player is there to see it. Per ADR
// 0002 nothing is recorded about what was witnessed; what the player makes of it
// is theirs.
func (g *GameInstance) noteObligationPerformed(actor *common.Actor, entry common.ScheduleEntry) {
	if !entry.IsObligation() {
		return
	}
	key := actor.GetInternalName() + ":" + entry.Obligation.ID
	if g.announcedToday[key] {
		return
	}
	g.announcedToday[key] = true

	if !g.state.GetPlayer().CanSee(actor.MapPosition()) {
		return
	}
	g.PlayerMessage(fmt.Sprintf("%s %s.", actor.Name(), entry.Obligation.Performed))
}

// showUnmetObligations is the absent face: what the town looks like where a duty
// has stopped being done. Without this the simulation degrades invisibly and the
// player has nothing to read.
func (g *GameInstance) showUnmetObligations() {
	for _, obligation := range g.unmetObligations {
		g.PlayerMessage(obligation.Absent + ".")
	}
}
