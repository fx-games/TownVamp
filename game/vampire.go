package game

import (
	"fmt"
	"rpg/common"
	"rpg/geometry"
)

// RunEnding says how a run finished. The Epilogue reads from it.
type RunEnding string

const (
	RunOngoing  RunEnding = ""
	RunLeftTown RunEnding = "left town"
	RunStarved  RunEnding = "starved"
	RunBurned   RunEnding = "caught by the sun"
	RunCaught   RunEnding = "caught by a hunter"
)

// KillTownspersonAt drains whoever is standing next to the given tile. The victim
// becomes Absent rather than dead as far as the town is concerned: nobody can
// yet tell a murder from someone having simply gone away, and the gap between
// the two is where a vampire lives.
func (g *GameInstance) KillTownspersonAt(target geometry.Point) bool {
	currentMap := g.state.GetMap()
	if !currentMap.IsActorAt(target) {
		return false
	}
	victim := currentMap.ActorAt(target)
	person := victim.GetTownsperson()
	if person == nil || person.IsAbsent() {
		return false
	}

	player := g.state.GetPlayer()
	if currentMap.MoveDistance(player.MapPosition(), target) > 1 {
		g.PlayerMessage("They are too far away.")
		return false
	}

	person.GoAbsent(g.state.GetClock().Day())
	g.blood.Feed()

	// The body stays where it fell. Whether it is found is a matter of who walks
	// past it, which is ticket 13.
	currentMap.RemoveActor(victim)
	g.state.GetTimeKeeper().ClearTrackedBeings(func(being common.TimedBeing) bool {
		return being == victim
	})
	g.bodies[person.InternalName] = target

	g.PlayerMessage(fmt.Sprintf("You drain %s. %s", victim.Name(), g.blood.String()))
	return true
}

// KillAdjacent is the player's verb: drain whoever is next to them. With only
// one Townsperson in reach there is nothing to choose between, so it takes the
// first neighbour it finds.
func (g *GameInstance) KillAdjacent() bool {
	player := g.state.GetPlayer()
	currentMap := g.state.GetMap()
	for _, neighbour := range currentMap.GetAllCardinalNeighbors(player.MapPosition()) {
		if currentMap.IsActorAt(neighbour) && currentMap.ActorAt(neighbour).IsTownsperson() {
			return g.KillTownspersonAt(neighbour)
		}
	}
	g.PlayerMessage("There is nobody here to drain.")
	return false
}

// HideBody moves a body out of sight, turning a killing into a disappearance.
// This is the player's central strategic verb: it is the difference between
// "she went away" and "someone is killing us".
func (g *GameInstance) HideBody() bool {
	player := g.state.GetPlayer()
	currentMap := g.state.GetMap()
	for name, at := range g.bodies {
		if currentMap.MoveDistance(player.MapPosition(), at) > 1 {
			continue
		}
		if person := g.townspersonByName(name); person != nil {
			person.BodyHidden = true
			delete(g.bodies, name)
			g.PlayerMessage("You put the body somewhere it will not be found.")
			return true
		}
	}
	g.PlayerMessage("There is no body here.")
	return false
}

func (g *GameInstance) townspersonByName(internalName string) *common.Townsperson {
	for _, person := range g.townspeople {
		if person.InternalName == internalName {
			return person
		}
	}
	return nil
}

// tickVampire drains Blood and applies the two things that can kill the player
// outright. It runs from the world clock, so it keeps happening whether or not
// the player does anything.
func (g *GameInstance) tickVampire(inGameSeconds float64) {
	if g.ending != RunOngoing || g.blood == nil {
		return
	}
	g.blood.Drain(inGameSeconds)

	if g.blood.IsEmpty() {
		g.endRun(RunStarved)
		return
	}

	// Sunlight. Being caught outside the lair at dawn is the hard deadline that
	// structures every night: how far it is safe to roam, and whether there is
	// time to get back.
	clock := g.state.GetClock()
	if clock.IsDaylight() && !g.isPlayerSheltered() {
		g.endRun(RunBurned)
		return
	}
	g.warnAboutDawn()
}

// LairRadius is how much of the map around the lair counts as shelter.
//
// ponytail: the town map marks no interiors — it has no lit tiles and no private
// zones — so there is nothing in the data to tell inside from outside. One
// authored lair is enough for the proof of concept and gives exactly the tension
// the design wants: be home before dawn. When the map gains zones (ticket 10),
// shelter should become "any private Zone with a roof" instead.
const LairRadius = 2

func (g *GameInstance) lairPosition() (geometry.Point, bool) {
	pos, ok := g.state.GetMap().GetNamedLocations()["spawn"]
	return pos, ok
}

func (g *GameInstance) isPlayerSheltered() bool {
	lair, ok := g.lairPosition()
	if !ok {
		return true // no lair authored; do not kill the player for the map's sake
	}
	return g.state.GetMap().MoveDistance(g.state.GetPlayer().MapPosition(), lair) <= LairRadius
}

// warnAboutDawn gives the player enough to plan around the deadline rather than
// be ambushed by it.
func (g *GameInstance) warnAboutDawn() {
	if g.isPlayerSheltered() {
		g.warnedAboutDawn = false
		return
	}
	remaining := g.state.GetClock().SecondsUntilDawn()
	if remaining > 0 && remaining < 3600 && !g.warnedAboutDawn {
		g.warnedAboutDawn = true
		g.PlayerMessage("The sky is beginning to grey. You should get back.")
	}
	if remaining > 3600 {
		g.warnedAboutDawn = false
	}
}

func (g *GameInstance) endRun(ending RunEnding) {
	if g.ending != RunOngoing {
		return
	}
	g.ending = ending
	defer g.ShowEpilogue()
	switch ending {
	case RunStarved:
		g.PlayerMessage("The hunger takes you. You are done.")
	case RunBurned:
		g.PlayerMessage("Dawn finds you in the open. You burn.")
	case RunCaught:
		g.PlayerMessage("A hand closes on your shoulder. They knew what you were.")
	case RunLeftTown:
		g.PlayerMessage("You leave the town behind.")
	}
}

// LeaveTown ends the run by choice, at any moment.
func (g *GameInstance) LeaveTown() {
	g.endRun(RunLeftTown)
}

func (g *GameInstance) Ending() RunEnding { return g.ending }
func (g *GameInstance) Blood() *common.Blood {
	return g.blood
}
