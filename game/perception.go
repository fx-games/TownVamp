package game

import (
	"rpg/common"
	"rpg/geometry"
)

const (
	// WitnessRange is how far a Townsperson can recognise what someone else is
	// doing. Beyond it a figure is a figure, not a witnessable act.
	WitnessRange = 8

	// After dark a human sees very little, and a lantern only buys the pool it
	// casts. The vampire is under no such limit — that asymmetry is the whole
	// reason he can watch without being watched, and it is what makes standing
	// in the dark a decision rather than flavour.
	//
	// ponytail: expressed as ranges because the town marks no lit tiles, so the
	// engine's own occlusion rule has nothing to work with. Replace with the
	// map's lighting once the town has them.
	NightSightRange = 2
	LanternRange    = 3
)

// canSee is the one perception query in the game. It is deliberately narrow:
// nothing computes a field of view for every Townsperson every tick. It runs
// only where secrecy is at stake — a Private Action in progress, a Hunter
// looking for the vampire, a body lying in the open — so its cost is bounded by
// how many of those exist rather than by the size of the town.
func (g *GameInstance) canSee(observer *common.Actor, target geometry.Point) bool {
	currentMap := g.state.GetMap()
	from := observer.MapPosition()

	if currentMap.MoveDistance(from, target) > g.sightRangeOf(observer) {
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

// sightRangeOf is how far an observer can make anything out. The vampire sees
// as well at night as by day; everyone else does not, which is the asymmetry the
// design depends on.
func (g *GameInstance) sightRangeOf(observer *common.Actor) int {
	if observer == nil {
		return 0
	}
	if observer == g.state.GetPlayer() {
		return WitnessRange
	}
	if g.state.IsDaylight() {
		return WitnessRange
	}
	if g.carriesLight(observer) {
		return LanternRange
	}
	return NightSightRange
}

// carriesLight reports whether an observer has a light of their own. Hunters
// carry one by definition.
func (g *GameInstance) carriesLight(observer *common.Actor) bool {
	if observer == nil {
		return false
	}
	if g.isHunter(observer) {
		return true
	}
	return observer.GetEquipment() != nil && observer.GetEquipment().HasLightSourceEquipped()
}

// forbiddenWatcher returns the first person able to see a place who is not on the
// given whitelist. Callers pass the actor doing the thing so they are not counted
// as their own witness.
func (g *GameInstance) forbiddenWatcher(place geometry.Point, actor *common.Actor, obligation *common.Obligation) (*common.Actor, bool) {
	for _, watcher := range g.whoCanSee(place, actor) {
		if !obligation.MayWitness(g.whitelistNameOf(watcher)) {
			return watcher, true
		}
	}
	return nil, false
}

// whitelistNameOf is how an actor is known to a Witness Whitelist. The vampire
// has no name there and never will: he is on nobody's list, which is what makes
// observing a secret require being hidden rather than merely being nearby.
func (g *GameInstance) whitelistNameOf(actor *common.Actor) string {
	if actor == g.state.GetPlayer() {
		return ""
	}
	return actor.GetInternalName()
}
