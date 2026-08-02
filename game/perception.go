package game

import (
	"rpg/common"
	"rpg/geometry"
)

const (
	// WitnessRange is how far a Townsperson can recognise what someone else is
	// doing. Beyond it a figure is a figure, not a witnessable act.
	WitnessRange = 8

	// LanternRange is how far anyone carrying a light can see at night. It is
	// short on purpose: the light that lets them move is the same light that
	// ruins their night vision, so a lantern shrinks the world to the pool it
	// casts. This is the asymmetry the whole design rests on, expressed as a
	// range rather than read off the map — the town marks no lit tiles, so the
	// engine's own occlusion rule has nothing to work with here.
	//
	// ponytail: replace with the map's lighting once the town has lit tiles.
	LanternRange = 3
)

// canSee is the one perception query in the game. It is deliberately narrow:
// nothing computes a field of view for every Townsperson every tick. It runs
// only where secrecy is at stake — a Private Action in progress, a Hunter
// looking for the vampire, a body lying in the open — so its cost is bounded by
// how many of those exist rather than by the size of the town.
func (g *GameInstance) canSee(observer *common.Actor, target geometry.Point) bool {
	currentMap := g.state.GetMap()
	from := observer.MapPosition()

	reach := WitnessRange
	if !g.state.IsDaylight() && g.carriesLight(observer) {
		reach = LanternRange
	}
	if currentMap.MoveDistance(from, target) > reach {
		return false
	}

	// The engine's own rule, kept for when the map gains lit tiles: a viewer
	// standing in light cannot see into the dark.
	if currentMap.IsTileLit(from) && !currentMap.IsTileLit(target) && !g.state.IsDaylight() {
		return false
	}

	blocked := geometry.LineOfSight(from, target, func(p geometry.Point) bool {
		return currentMap.Contains(p) && currentMap.IsTransparent(p)
	})
	if len(blocked) == 0 {
		return false
	}
	return blocked[len(blocked)-1] == target
}

// whoCanSee returns the Townspeople — and the player, when relevant — able to
// observe a place right now. Callers pass the actor doing the thing so they are
// not counted as their own witness.
func (g *GameInstance) whoCanSee(place geometry.Point, except *common.Actor) []*common.Actor {
	var watchers []*common.Actor
	for _, actor := range g.state.GetMap().Actors() {
		if actor == except {
			continue
		}
		if g.canSee(actor, place) {
			watchers = append(watchers, actor)
		}
	}
	return watchers
}

// carriesLight reports whether an observer is lit, and therefore blind beyond
// the pool their own lantern casts. Hunters carry one by definition; ordinary
// Townspeople are indoors after dark and are not out looking for anyone.
func (g *GameInstance) carriesLight(observer *common.Actor) bool {
	if observer == nil {
		return false
	}
	if g.isHunter(observer) {
		return true
	}
	return observer.GetEquipment() != nil && observer.GetEquipment().HasLightSourceEquipped()
}
