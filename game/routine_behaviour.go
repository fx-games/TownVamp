package game

import (
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

	slot, hasSlot := person.Routine.SlotAt(g.state.GetClock().SecondsSinceMidnight())
	if !hasSlot {
		return waitOneStep(stepCost) // unclaimed hour: spare capacity, spent standing still
	}

	here := actor.MapPosition()
	if here == slot.At || g.state.GetMap().MoveDistance(here, slot.At) <= 0 {
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
