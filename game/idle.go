package game

import (
	"rpg/common"
	"rpg/geometry"
	"sort"
)

// An hour nothing is authored into is spare capacity, and spare capacity is
// what the town repairs itself out of. It is not, however, an hour a person
// spends standing still: a town where everyone freezes between appointments
// reads as a mechanism waiting for input rather than a place, and the whole
// design rests on the player being willing to watch it.
//
// So an unclaimed hour is spent going somewhere and standing about there for a
// while. Three rules keep this from costing anything:
//
//   - It never makes anyone late. A destination is only chosen if the walk out
//     and the walk on to the next appointment both still fit in the time left.
//     What the player learns by watching a Routine stays exactly true.
//   - It never consumes capacity. Nothing is added to the Routine or the
//     Schedule, so Reassignment sees the same free hours it saw before.
//   - It never takes anyone anywhere they do not belong. The same predicate
//     that governs trespass governs where someone will idle.
type idleDrift struct {
	to      geometry.Point
	arrived bool
	linger  float64 // in-game seconds left to stand here before moving on
}

// minimumLinger stops a Townsperson ricocheting between two places: having
// gone somewhere, they are there for a while.
const minimumLinger = 8 * 60.0

// getIdleAction is what someone does with an hour nobody has claimed.
func (g *GameInstance) getIdleAction(actor *common.Actor, person *common.Townsperson, stepCost float64) (float64, func()) {
	drift, known := g.idleDrift[person.InternalName]
	if !known || drift.arrived && drift.linger <= 0 {
		destination, found := g.pickSomewhereToBe(actor, person)
		if !found {
			return waitOneStep(stepCost)
		}
		drift = &idleDrift{to: destination, linger: g.lingerDuration()}
		g.idleDrift[person.InternalName] = drift
	}

	if g.hasArrivedAt(actor, drift.to) {
		drift.arrived = true
		drift.linger -= stepCost
		return waitOneStep(stepCost)
	}

	next, reachable := g.nextStepTowards(actor, drift.to)
	if !reachable {
		// Someone in the doorway, or nowhere to go at all. Wait it out and try
		// somewhere else once the lingering is up.
		drift.linger -= stepCost
		if drift.linger <= 0 {
			drift.arrived = true
		}
		return waitOneStep(stepCost)
	}
	return stepCost, func() {
		if g.state.IsClosedDoorAt(next) {
			g.actionBroker.Execute(common.NewOpenDoorAction(common.AbstractCause("Idle"), actor, next))
			return
		}
		g.actionBroker.Execute(common.NewMoveAction(common.AbstractCause("Idle"), actor, next, false))
	}
}

// pickSomewhereToBe chooses a place to spend part of an unclaimed hour: near
// enough to get to and back from before the next appointment, and somewhere this
// person is entitled to be.
func (g *GameInstance) pickSomewhereToBe(actor *common.Actor, person *common.Townsperson) (geometry.Point, bool) {
	// Someone who has seen the vampire where he did not belong stops wandering.
	// Withdrawal is a Routine bias; this is the same bias applied to the hours
	// the Routine does not cover, and without it a withdrawn Townsperson would
	// spend their free hours back out in the open.
	if person.HasWithdrawn {
		if home, found := person.HomeLocation(); found {
			return home, true
		}
		return geometry.Point{}, false
	}

	here := actor.MapPosition()
	budget := g.secondsUntilDue(person)
	next, hasNext, _ := g.nextDue(person)

	// Straight-line distance first, which costs nothing and can only ever be
	// shorter than the walk: anything that fails here cannot fit, and the
	// pathfinder is never asked about it.
	currentMap := g.state.GetMap()
	affordableSteps := int((budget - minimumLinger) / common.BaseSecondsPerStep)
	var plausible []geometry.Point
	for _, candidate := range g.idleCandidates(person) {
		if candidate == here {
			continue
		}
		steps := currentMap.MoveDistance(here, candidate)
		if hasNext {
			steps += currentMap.MoveDistance(candidate, next.At)
		}
		if steps > affordableSteps {
			continue
		}
		plausible = append(plausible, candidate)
	}

	// Then pay for the real path, on a handful of them at most.
	for attempt := 0; attempt < 4 && len(plausible) > 0; attempt++ {
		pick := g.randomSource.Intn(len(plausible))
		candidate := plausible[pick]
		if g.fitsBeforeDue(here, candidate, person, budget) {
			return candidate, true
		}
		plausible = append(plausible[:pick], plausible[pick+1:]...)
	}
	return geometry.Point{}, false
}

// nextDue is where this person is next expected, and how long they have to get
// there. A Schedule with nothing in it leaves them the whole day.
func (g *GameInstance) nextDue(person *common.Townsperson) (common.ScheduleEntry, bool, float64) {
	entry, until, found := person.Schedule().NextEntryAfter(g.state.GetClock().SecondsSinceMidnight())
	if !found {
		return common.ScheduleEntry{}, false, float64(common.SecondsPerDay)
	}
	return entry, true, until
}

func (g *GameInstance) secondsUntilDue(person *common.Townsperson) float64 {
	_, _, until := g.nextDue(person)
	return until
}

// fitsBeforeDue asks the only question that matters about an idle errand: can
// they go there, stand about, and still be where they are next due on time.
func (g *GameInstance) fitsBeforeDue(here, candidate geometry.Point, person *common.Townsperson, budget float64) bool {
	next, found, _ := g.nextDue(person)
	out, canGo := g.travelSeconds(here, candidate)
	if !canGo {
		return false
	}
	if !found {
		return out+minimumLinger <= budget
	}
	onwards, canReturn := g.travelSeconds(candidate, next.At)
	if !canReturn {
		return false
	}
	return out+minimumLinger+onwards <= budget
}

func (g *GameInstance) travelSeconds(from, to geometry.Point) (float64, bool) {
	if from == to {
		return 0, true
	}
	currentMap := g.state.GetMap()
	path := currentMap.GetJPSPath(from, to, func(p geometry.Point) bool {
		return currentMap.Contains(p) && currentMap.IsWalkable(p)
	})
	if len(path) == 0 {
		return 0, false
	}
	return float64(len(path)) * common.BaseSecondsPerStep, true
}

// idleCandidates are the places on the map this person might drift to: every
// named place they are entitled to stand in. Sorted, because the choice is made
// with the seeded random source and has to fall the same way every run.
func (g *GameInstance) idleCandidates(person *common.Townsperson) []geometry.Point {
	named := g.state.GetMap().GetNamedLocations()
	names := make([]string, 0, len(named))
	for name := range named {
		names = append(names, name)
	}
	sort.Strings(names)

	candidates := make([]geometry.Point, 0, len(names))
	for _, name := range names {
		at := named[name]
		if zone := g.zoneAt(at); zone != nil && !zone.MayEnter(person.InternalName) {
			continue
		}
		candidates = append(candidates, at)
	}
	return candidates
}

// lingerDuration is how long someone stands about once they get there. Varied,
// so the town does not visibly tick over all at once.
func (g *GameInstance) lingerDuration() float64 {
	return minimumLinger + float64(g.randomSource.Intn(25))*60.0
}
