package game

import (
	"rpg/common"
	"strings"
	"testing"
)

// killByName drains a named Townsperson regardless of where the player is
// standing, so tests can choose their victim rather than walk to them.
func killByName(t *testing.T, game *GameInstance, internalName string) *common.Townsperson {
	t.Helper()
	currentMap := game.GetWorldState().GetMap()
	for _, actor := range currentMap.Actors() {
		if actor.GetInternalName() != internalName {
			continue
		}
		game.GetWorldState().GetPlayer().SetMapPosition(actor.MapPosition())
		if !game.KillTownspersonAt(actor.MapPosition()) {
			t.Fatalf("could not drain %s", internalName)
		}
		return actor.GetTownsperson()
	}
	t.Fatalf("%s is not in the town", internalName)
	return nil
}

func TestBloodDrainsAndAKillRefillsIt(t *testing.T) {
	game := newTestGame(t)
	blood := game.Blood()

	full := blood.Current()
	advanceRealSeconds(game, 12*60) // half a Day
	drained := blood.Current()
	if drained >= full {
		t.Fatalf("Blood did not drain over half a Day: %.1f -> %.1f", full, drained)
	}

	killByName(t, game, "greta")
	if game.Blood().Current() <= drained {
		t.Errorf("a Kill did not refill Blood: %.1f -> %.1f", drained, game.Blood().Current())
	}
}

func TestRunningOutOfBloodEndsTheRun(t *testing.T) {
	game := newTestGame(t)
	for i := 0; i < 5*24*60*oneRealSecond && game.Ending() == RunOngoing; i++ {
		game.Advance(1.0)
	}
	if game.Ending() != RunStarved {
		t.Errorf("expected to starve without feeding, ended as %q", game.Ending())
	}
}

func TestAKilledTownspersonBecomesAbsentNotDead(t *testing.T) {
	// The town cannot tell a murder from someone having gone away, and that gap
	// is the vampire's operating room.
	game := newTestGame(t)
	victim := killByName(t, game, "greta")

	if victim.Presence != common.PresenceAbsent {
		t.Errorf("victim should be Absent, is %v", victim.Presence)
	}
	if victim.Presence == common.PresenceConfirmedDead {
		t.Error("the town should not yet know they are dead")
	}
}

func TestObligationsDoNotMoveOnTheDayOfTheKill(t *testing.T) {
	// Reassignment waits a full Day; the town cannot react to a disappearance it
	// has not noticed.
	game := newTestGame(t)
	victim := killByName(t, game, "hanno")
	held := len(victim.Obligations)
	if held == 0 {
		t.Fatal("victim owed nothing, so this proves nothing")
	}

	// Move to just before midnight without crossing it.
	clock := game.GetWorldState().GetClock()
	for clock.Hour() != 23 {
		game.Advance(1.0)
	}
	if len(victim.Obligations) != held {
		t.Errorf("duties moved before a full Day had passed")
	}
}

func TestObligationsAreReassignedAfterADay(t *testing.T) {
	game := newTestGame(t)
	victim := killByName(t, game, "hanno")
	if len(victim.Obligations) == 0 {
		t.Fatal("victim owed nothing")
	}

	startDay := game.GetWorldState().GetClock().Day()
	for game.GetWorldState().GetClock().Day() < startDay+2 && game.Ending() == RunOngoing {
		game.Advance(1.0)
	}

	if len(victim.Obligations) != 0 {
		t.Errorf("the victim still owes %d duties after two Days", len(victim.Obligations))
	}

	// Every voided duty either found a taker or is on the Unmet ledger.
	accountedFor := len(game.UnmetObligations())
	for _, actor := range townspeople(game) {
		for _, obligation := range actor.GetTownsperson().Obligations {
			if obligation.Owner == "hanno" || obligation.ID == "bread_for_mary" || obligation.ID == "water_for_the_east_quarter" {
				accountedFor++
			}
		}
	}
	if accountedFor == 0 {
		t.Error("the victim's duties vanished without being reassigned or recorded as Unmet")
	}
}

func TestReassignmentPrefersThePersonTheDutyPointedAt(t *testing.T) {
	// Relationship before competence: the widow being fed can fetch her own bread.
	game := newTestGame(t)

	// bread_for_mary is owed by hanno to mary, and mary has thirteen spare hours.
	killByName(t, game, "hanno")
	startDay := game.GetWorldState().GetClock().Day()
	for game.GetWorldState().GetClock().Day() < startDay+2 && game.Ending() == RunOngoing {
		game.Advance(1.0)
	}

	mary := game.townspersonByName("mary")
	if mary == nil {
		t.Fatal("mary is missing")
	}
	for _, obligation := range mary.Obligations {
		if obligation.ID == "bread_for_mary" {
			return // took it herself, as intended
		}
	}
	t.Error("bread_for_mary was not offered to Mary, who it was owed to and who has the hours")
}

func TestADutyNobodyCanTakeGoesUnmetAndIsVisible(t *testing.T) {
	// The absent face has to reach the player, or the town degrades invisibly.
	game := newTestGame(t)

	// unlock_the_town_gate is owed by shekel to the town with no suitable Role,
	// so it dies with him.
	killByName(t, game, "shekel")
	startDay := game.GetWorldState().GetClock().Day()
	for game.GetWorldState().GetClock().Day() < startDay+2 && game.Ending() == RunOngoing {
		game.Advance(1.0)
	}

	var unmet *common.Obligation
	for _, obligation := range game.UnmetObligations() {
		if obligation.ID == "unlock_the_town_gate" {
			unmet = obligation
		}
	}
	if unmet == nil {
		t.Fatalf("unlock_the_town_gate should be Unmet; ledger holds %d", len(game.UnmetObligations()))
	}

	found := false
	for _, line := range game.GetLog() {
		if strings.Contains(line, unmet.Absent) {
			found = true
		}
	}
	if !found {
		t.Error("the town never showed the player what the missing duty looks like")
	}
}

func TestHidingABodyIsPossible(t *testing.T) {
	game := newTestGame(t)
	victim := killByName(t, game, "greta")
	if !game.HideBody() {
		t.Fatal("could not hide the body the player is standing on")
	}
	if !victim.BodyHidden {
		t.Error("the body is not recorded as hidden")
	}
}

func TestLeavingTownEndsTheRun(t *testing.T) {
	game := newTestGame(t)
	game.LeaveTown()
	if game.Ending() != RunLeftTown {
		t.Errorf("expected to have left town, got %q", game.Ending())
	}
}
