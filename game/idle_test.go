package game

import (
	"rpg/common"
	"rpg/geometry"
	"testing"
)

// An hour nobody has claimed is still an hour the player is watching. These
// tests are about the two halves of that: the town has to keep moving, and the
// moving must not cost it a single appointment.

func TestTownspeopleGoSomewhereInTheirUnclaimedHours(t *testing.T) {
	game := newTestGame(t)
	clock := game.GetWorldState().GetClock()

	drifted := map[string]bool{}
	previous := map[string]geometry.Point{}
	for _, actor := range townspeople(game) {
		previous[actor.GetInternalName()] = actor.MapPosition()
	}

	// A day of it. Somebody, at some point, has to leave the spot their last
	// duty left them standing on.
	for hour := 0; hour < 24; hour++ {
		advanceRealSeconds(game, 60)
		for _, actor := range townspeople(game) {
			person := actor.GetTownsperson()
			name := actor.GetInternalName()
			at := actor.MapPosition()
			if _, claimed := person.Schedule().EntryAt(clock.SecondsSinceMidnight()); !claimed {
				if at != previous[name] {
					drifted[name] = true
				}
			}
			previous[name] = at
		}
	}

	if len(drifted) == 0 {
		t.Error("nobody moved during an unclaimed hour; the town stands still whenever it is off duty")
	}
}

func TestAWithdrawnTownspersonStopsWandering(t *testing.T) {
	// Withdrawal is the price of being seen somewhere the vampire does not
	// belong. It biases the Routine indoors; the hours the Routine does not
	// cover have to follow it there, or the bias buys the player nothing.
	game := newTestGame(t)
	var subject *common.Townsperson
	for _, person := range game.townspeople {
		if _, hasHome := person.HomeLocation(); hasHome && person.FreeSecondsPerDay() > 2*3600 {
			subject = person
			break
		}
	}
	if subject == nil {
		t.Fatal("nobody has both a home and hours to spare")
	}
	home, _ := subject.HomeLocation()
	subject.Withdraw()

	actor := actorFor(game, subject.InternalName)
	if actor == nil {
		t.Fatalf("%s is not on the map", subject.InternalName)
	}
	where, found := game.pickSomewhereToBe(actor, subject)
	if !found || where != home {
		t.Errorf("a withdrawn Townsperson drifts to %v, not home to %v", where, home)
	}
}

func actorFor(game *GameInstance, internalName string) *common.Actor {
	for _, actor := range game.GetWorldState().GetMap().Actors() {
		if actor.GetInternalName() == internalName {
			return actor
		}
	}
	return nil
}
