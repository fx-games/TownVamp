package game

import (
	"fmt"
	"rpg/common"
	"rpg/defs"
	"rpg/geometry"
	"rpg/util"
)

// checkForBodyDiscovery tests whether anyone's line of sight crosses a corpse.
//
// It is evaluated per body rather than per observer. Corpses are few and
// Townspeople are many, so the cost stays bounded by how many bodies are lying
// about — which is also why no general NPC vision is needed. A body in the
// square is found within the hour; one in a cellar may hold for weeks.
func (g *GameInstance) checkForBodyDiscovery() {
	if len(g.bodies) == 0 {
		return
	}
	for name, at := range g.bodies {
		person := g.townspersonByName(name)
		if person == nil || person.BodyHidden {
			continue
		}
		for _, finder := range g.whoCanSee(at, nil) {
			if finder == g.state.GetPlayer() {
				continue // the killer finding their own work proves nothing
			}
			if finder.GetTownsperson() == nil || finder.GetTownsperson().IsAbsent() {
				continue
			}
			g.confirmDeath(person, finder, at)
			break
		}
	}
}

// confirmDeath is the moment the town stops wondering and starts knowing. The
// duties have already moved by now — what changes is that a Hunter takes to the
// streets.
func (g *GameInstance) confirmDeath(person *common.Townsperson, finder *common.Actor, at geometry.Point) {
	person.ConfirmDead()
	delete(g.bodies, person.InternalName)

	g.PlayerMessage(fmt.Sprintf("%s finds %s's body.", finder.Name(), displayName(person.InternalName)))
	g.spawnHunter(at)
}

// spawnHunter puts someone on the streets who will not run. A Hunter is the
// single exception to the rule that Townspeople flee: it pursues, and catching
// the vampire ends the run.
//
// There is no fight. No health, no damage, no armour, no attack rolls — the
// combat systems were deleted and none of them come back for this. A Hunter is a
// chase and a proximity rule.
func (g *GameInstance) spawnHunter(near geometry.Point) {
	if g.hunter != nil {
		return // one is enough to change how the town feels
	}
	currentMap := g.state.GetMap()

	spawnAt := near
	if currentMap.IsActorAt(spawnAt) || !currentMap.IsWalkable(spawnAt) {
		free := currentMap.GetFreeCardinalNeighbors(near)
		if len(free) == 0 {
			return
		}
		spawnAt = free[0]
	}

	hunter := common.NewActor(int32('H'))
	hunter.SetInternalName("hunter")
	hunter.SetName("A hunter")
	hunter.SetAsciiIcon('H')
	hunter.SetCP437Icon(int32(util.UnicodeToCP437Byte('H')))
	hunter.SetColor(defs.White)
	hunter.SetDrawPosition(spawnAt.ToPointF())
	currentMap.AddActor(hunter, spawnAt)
	g.state.GetTimeKeeper().AddTimed(hunter)
	g.hunter = hunter

	g.PlayerMessage("Someone begins walking the streets at night, carrying a lantern.")
}

// getHunterAction is the Hunter's whole behaviour: walk the town at night, and
// pursue the vampire on sight.
//
// The lantern it carries is what limits it. A Hunter sees only the pool its own
// light casts, while the vampire sees the whole street — so it hunts him with
// the one thing that keeps it from finding him.
func (g *GameInstance) getHunterAction(actor *common.Actor) (float64, func()) {
	stepCost := common.SecondsPerStep(actor.GetMoveSpeed())
	player := g.state.GetPlayer()
	if player == nil || !player.IsAlive() || g.ending != RunOngoing {
		return waitOneStep(stepCost)
	}

	// A Hunter walks the town at night. By day there is nothing to find, and the
	// vampire is not out in it.
	if g.state.IsDaylight() {
		return waitOneStep(stepCost)
	}

	currentMap := g.state.GetMap()
	if currentMap.MoveDistance(actor.MapPosition(), player.MapPosition()) <= 1 {
		return stepCost, func() { g.endRun(RunCaught) }
	}

	if g.canSee(actor, player.MapPosition()) {
		next, found := g.nextStepTowards(actor, player.MapPosition())
		if found {
			return stepCost, func() {
				g.actionBroker.Execute(common.NewMoveAction(common.AbstractCause("Hunt"), actor, next, false))
			}
		}
		return waitOneStep(stepCost)
	}

	// Not in sight: patrol towards the last place a body was found, or wander.
	target := g.hunterPatrolTarget()
	if target == actor.MapPosition() {
		return waitOneStep(stepCost)
	}
	next, found := g.nextStepTowards(actor, target)
	if !found {
		return waitOneStep(stepCost)
	}
	return stepCost, func() {
		g.actionBroker.Execute(common.NewMoveAction(common.AbstractCause("Patrol"), actor, next, false))
	}
}

// hunterPatrolTarget sends the Hunter wandering between the map's named places,
// so it covers ground rather than standing still.
func (g *GameInstance) hunterPatrolTarget() geometry.Point {
	locations := g.state.GetMap().GetNamedLocations()
	if len(locations) == 0 {
		return g.hunter.MapPosition()
	}
	names := make([]string, 0, len(locations))
	for name := range locations {
		names = append(names, name)
	}
	// Deterministic given the seed: the world's random source, not the global one.
	return locations[names[g.randomSource.Intn(len(names))]]
}

func (g *GameInstance) isHunter(actor *common.Actor) bool {
	return g.hunter != nil && actor == g.hunter
}
