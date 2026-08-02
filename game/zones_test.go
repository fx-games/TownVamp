package game

import (
	"rpg/common"
	"rpg/geometry"
	"strings"
	"testing"
)

func TestTownHasPrivateZones(t *testing.T) {
	game := newTestGame(t)
	private := 0
	for _, zone := range game.zones {
		if zone.Private {
			private++
		}
	}
	if private == 0 {
		t.Fatal("no private Zones, so trespass can never happen")
	}
}

func TestAZoneBelongsToItsResidentsAndNobodyElse(t *testing.T) {
	game := newTestGame(t)
	for _, zone := range game.zones {
		if !zone.Private || len(zone.BelongsTo) == 0 {
			continue
		}
		if !zone.MayEnter(zone.BelongsTo[0]) {
			t.Errorf("%s excludes its own resident", zone.Name)
		}
		if zone.MayEnter("someone_else") {
			t.Errorf("%s admits a stranger", zone.Name)
		}
		if zone.MayEnter("") {
			t.Errorf("%s admits the vampire", zone.Name)
		}
		return
	}
	t.Skip("no private Zone with residents")
}

// standInside puts the player in a private Zone belonging to the named
// Townsperson, next to them, in daylight so they are certain to be noticed.
func standInside(t *testing.T, game *GameInstance, internalName string) *common.Actor {
	t.Helper()
	clock := game.GetWorldState().GetClock()
	for !clock.IsDaylight() {
		game.Advance(1.0)
	}
	for _, zone := range game.zones {
		if !zone.Private || !zone.MayEnter(internalName) {
			continue
		}
		for _, actor := range game.GetWorldState().GetMap().Actors() {
			if actor.GetInternalName() != internalName {
				continue
			}
			inside, ok := walkableTileIn(game, zone)
			if !ok {
				continue // this zone is all wall; try another
			}
			game.GetWorldState().GetPlayer().SetMapPosition(inside)
			actor.SetMapPosition(inside)
			return actor
		}
	}
	t.Fatalf("no private Zone belonging to %s", internalName)
	return nil
}

func TestBeingSeenTrespassingMakesThemWithdraw(t *testing.T) {
	game := newTestGame(t)
	resident := standInside(t, game, "greta")
	person := resident.GetTownsperson()
	if person.HasWithdrawn {
		t.Fatal("already withdrawn before anything happened")
	}

	game.checkForTrespass()

	if !person.HasWithdrawn {
		t.Fatal("a Townsperson who saw the vampire in their own house did not Withdraw")
	}
	found := false
	for _, line := range game.GetLog() {
		if strings.Contains(line, "hurries away") {
			found = true
		}
	}
	if !found {
		t.Error("the player was never told they had been seen")
	}
}

func TestWithdrawalMovesHoursIndoorsWithoutStoppingWork(t *testing.T) {
	// A bias, not confinement: they keep doing their job, but the loose hours
	// spent out in the open come home.
	// Abel: his day has a two-hour stop at the tavern, which is exactly the kind
	// of loose hour Withdrawal is meant to take away. Greta's day is one long
	// shift at the forge and has nothing to give up, which is the point.
	game := newTestGame(t)
	resident := standInside(t, game, "abel")
	person := resident.GetTownsperson()

	home, hasHome := homeOf(person)
	if !hasHome {
		t.Fatal("no home to withdraw to")
	}
	outdoorsBefore := slotsAwayFrom(person, home)
	longestBefore := longestSlotSeconds(person)

	game.checkForTrespass()

	outdoorsAfter := slotsAwayFrom(person, home)
	if outdoorsAfter >= outdoorsBefore {
		t.Errorf("Withdrawal did not move any hours indoors: %d -> %d", outdoorsBefore, outdoorsAfter)
	}
	if longestSlotSeconds(person) != longestBefore {
		t.Error("Withdrawal took away their actual work; it is meant to be a bias, not confinement")
	}
}

func TestWithdrawalIsPermanentAndSilent(t *testing.T) {
	game := newTestGame(t)
	resident := standInside(t, game, "greta")
	game.checkForTrespass()

	// Anyone else in the room may legitimately have seen it too. What must not
	// happen is word spreading: someone across town withdraws only if they saw it
	// themselves.
	for _, actor := range townspeople(game) {
		other := actor.GetTownsperson()
		if actor == resident || !other.HasWithdrawn {
			continue
		}
		if !game.canSee(actor, game.GetWorldState().GetPlayer().MapPosition()) {
			t.Errorf("%s withdrew without seeing anything; word is spreading", other.InternalName)
		}
	}

	advanceRealSeconds(game, 24*60)
	if !resident.GetTownsperson().HasWithdrawn {
		t.Error("Withdrawal wore off; it is meant to last the run")
	}
}

func homeOf(person *common.Townsperson) (geometry.Point, bool) {
	for _, slot := range person.Routine {
		if slot.Activity == "sleep" {
			return slot.At, true
		}
	}
	return geometry.Point{}, false
}

func slotsAwayFrom(person *common.Townsperson, home geometry.Point) int {
	away := 0
	for _, slot := range person.Routine {
		if slot.At != home {
			away++
		}
	}
	return away
}

func longestSlotSeconds(person *common.Townsperson) int {
	longest := 0
	for _, slot := range person.Routine {
		if d := slot.DurationSeconds(); d > longest {
			longest = d
		}
	}
	return longest
}

// walkableTileIn finds somewhere inside a Zone that anyone could actually stand.
func walkableTileIn(game *GameInstance, zone *TownZone) (geometry.Point, bool) {
	currentMap := game.GetWorldState().GetMap()
	for x := zone.Rect.Min.X; x < zone.Rect.Max.X; x++ {
		for y := zone.Rect.Min.Y; y < zone.Rect.Max.Y; y++ {
			at := geometry.Point{X: x, Y: y}
			if currentMap.Contains(at) && currentMap.IsWalkable(at) {
				return at, true
			}
		}
	}
	return geometry.Point{}, false
}

func TestEveryZoneHasSomewhereToStand(t *testing.T) {
	// A Zone drawn over solid rock can never be trespassed in, so it may as well
	// not exist.
	game := newTestGame(t)
	for _, zone := range game.zones {
		if _, ok := walkableTileIn(game, zone); !ok {
			t.Errorf("zone %q at %v contains no walkable tile", zone.Name, zone.Rect)
		}
	}
}
