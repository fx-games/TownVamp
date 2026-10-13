package game

import (
	"testing"
)

// Daywalker is a playtesting switch: the vampire can stand in the sun and walk
// into someone's house, and nothing comes of either.
func TestDaywalkerIsUnseenAndUnkillable(t *testing.T) {
	game := newTestGame(t)
	game.SetDaywalker(true)

	resident := standInside(t, game, "greta")
	game.checkForTrespass()
	if resident.GetTownsperson().HasWithdrawn {
		t.Error("a Townsperson saw the Daywalker trespassing")
	}
	if game.canSeePlayer(resident) {
		t.Error("a Townsperson standing by the Daywalker can see him")
	}
	for _, watcher := range game.whoCanSee(resident.MapPosition(), resident) {
		if watcher == game.GetWorldState().GetPlayer() {
			t.Error("the Daywalker counts as a witness")
		}
	}

	clock := game.GetWorldState().GetClock()
	for i := 0; i < 24*60*oneRealSecond && !clock.IsDaylight(); i++ {
		game.Advance(1.0)
	}
	stepIntoTheOpen(t, game)
	before := game.Blood().Current()
	for i := 0; i < 3*24*60*oneRealSecond && game.Ending() == RunOngoing; i++ {
		game.Advance(1.0)
	}
	if game.Ending() != RunOngoing {
		t.Fatalf("three Days in the open ended the Daywalker's run as %q", game.Ending())
	}
	if game.Blood().Current() < before {
		t.Errorf("the Daywalker lost Blood: %.1f -> %.1f", before, game.Blood().Current())
	}
}
