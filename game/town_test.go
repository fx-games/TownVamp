package game

import (
	"rpg/common"
	"rpg/geometry"
	"testing"
)

func townspeople(game *GameInstance) []*common.Actor {
	var people []*common.Actor
	for _, actor := range game.GetWorldState().GetMap().Actors() {
		if actor.IsTownsperson() {
			people = append(people, actor)
		}
	}
	return people
}

func TestTownHasSixTownspeople(t *testing.T) {
	game := newTestGame(t)
	people := townspeople(game)
	if len(people) != 6 {
		t.Fatalf("expected 6 Townspeople, got %d", len(people))
	}
}

func TestAtLeastFourRolesAreRepresented(t *testing.T) {
	// Reassignment falls back to Role once relationship has been tried, so the
	// fallback needs somewhere to go.
	game := newTestGame(t)
	roles := map[common.Role]int{}
	for _, actor := range townspeople(game) {
		roles[actor.GetTownsperson().Role]++
	}
	if len(roles) < 4 {
		t.Errorf("expected at least 4 of the 5 Roles, got %d: %v", len(roles), roles)
	}
	for role := range roles {
		if _, err := common.RoleFromString(string(role)); err != nil {
			t.Errorf("%v is not one of the five Roles", role)
		}
	}
}

func TestRoutineDestinationsAreReachable(t *testing.T) {
	// A Routine is only real if its places exist and can be walked to. This is
	// the cheap ancestor of the Schedule feasibility check ticket 08 needs.
	game := newTestGame(t)
	currentMap := game.GetWorldState().GetMap()
	if len(townspeople(game)) == 0 {
		t.Fatal("no Townspeople loaded")
	}

	for _, actor := range townspeople(game) {
		person := actor.GetTownsperson()
		for _, slot := range person.Routine {
			if !currentMap.Contains(slot.At) {
				t.Errorf("%s: %s at %v is off the map", person.InternalName, slot.Activity, slot.At)
				continue
			}
			if !currentMap.IsWalkable(slot.At) {
				t.Errorf("%s: %s at %v is not walkable", person.InternalName, slot.Activity, slot.At)
				continue
			}
			path := currentMap.GetJPSPath(actor.MapPosition(), slot.At, func(p geometry.Point) bool {
				return currentMap.Contains(p) && currentMap.IsWalkable(p)
			})
			if len(path) == 0 && actor.MapPosition() != slot.At {
				t.Errorf("%s: cannot reach %s at %v from %v", person.InternalName, slot.Activity, slot.At, actor.MapPosition())
			}
		}
	}
}

func TestRoutineDensityVaries(t *testing.T) {
	// One Townsperson must be the town's shock absorber and one must be fragile,
	// or Reassignment has nothing interesting to discover later.
	game := newTestGame(t)
	leanest, busiest := -1, common.SecondsPerDay+1
	for _, actor := range townspeople(game) {
		free := actor.GetTownsperson().Routine.FreeSecondsPerDay()
		if free > leanest {
			leanest = free
		}
		if free < busiest {
			busiest = free
		}
	}
	if leanest < 4*3600 {
		t.Errorf("no Townsperson has meaningful spare capacity; leanest has %ds free", leanest)
	}
	if busiest > 2*3600 {
		t.Errorf("no Townsperson is genuinely busy; busiest still has %ds free", busiest)
	}
}

func TestTownspeopleWalkTheirRoutinesUnattended(t *testing.T) {
	game := newTestGame(t)
	people := townspeople(game)
	if len(people) == 0 {
		t.Fatal("no Townspeople loaded")
	}

	// Run to just before the town wakes; before this everyone is asleep at home
	// and standing still is correct behaviour, not a stalled simulation.
	clock := game.GetWorldState().GetClock()
	for clock.Hour() != 4 {
		game.Advance(1.0)
	}

	start := map[string]geometry.Point{}
	for _, actor := range people {
		start[actor.Name()] = actor.MapPosition()
	}

	advanceRealSeconds(game, 8*60) // through the morning, when Routines send people out

	moved := 0
	for _, actor := range people {
		if actor.MapPosition() != start[actor.Name()] {
			moved++
		}
	}
	if moved < len(people)-1 {
		t.Errorf("only %d of %d Townspeople moved over ten in-game hours", moved, len(people))
	}
}

func TestTownspeopleReachWhereTheirRoutineSendsThem(t *testing.T) {
	game := newTestGame(t)
	clock := game.GetWorldState().GetClock()

	// Run to mid-morning, when most Routines have people at work.
	for clock.Hour() != 11 {
		game.Advance(1.0)
	}
	advanceRealSeconds(game, 20) // twenty more in-game minutes to finish walking

	arrived, expected := 0, 0
	for _, actor := range townspeople(game) {
		slot, ok := actor.GetTownsperson().Routine.SlotAt(clock.SecondsSinceMidnight())
		if !ok {
			continue
		}
		expected++
		if game.hasArrivedAt(actor, slot.At) {
			arrived++
		} else {
			t.Logf("%s is at %v, their Routine says %s at %v",
				actor.Name(), actor.MapPosition(), slot.Activity, slot.At)
		}
	}
	if expected == 0 {
		t.Fatal("no Townsperson has a Routine slot at 11:00")
	}
	if arrived < expected {
		t.Errorf("%d of %d Townspeople reached where their Routine sends them", arrived, expected)
	}
}
