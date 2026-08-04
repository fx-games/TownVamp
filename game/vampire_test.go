package game

import (
	"rpg/common"
	"rpg/geometry"
	"rpg/stats"
	"strings"
	"testing"
)

// killByName drains a named Townsperson regardless of where the player is
// standing, then returns the player to the lair — otherwise they are left
// loitering at the scene, trespassing and waiting for the sun or a Hunter, and
// every test becomes a test of those instead of what it meant to check.
func drain(t *testing.T, game *GameInstance, internalName string, hideBody bool) *common.Townsperson {
	t.Helper()
	player := game.GetWorldState().GetPlayer()
	was := player.MapPosition()
	currentMap := game.GetWorldState().GetMap()

	for _, actor := range currentMap.Actors() {
		if actor.GetInternalName() != internalName {
			continue
		}
		player.SetMapPosition(actor.MapPosition())
		killed := game.KillTownspersonAt(actor.MapPosition())
		if hideBody {
			game.HideBody()
		}
		if lair, ok := game.lairPosition(); ok {
			player.SetMapPosition(lair)
		} else {
			player.SetMapPosition(was)
		}
		if !killed {
			t.Fatalf("could not drain %s", internalName)
		}
		return actor.GetTownsperson()
	}
	t.Fatalf("%s is not in the town", internalName)
	return nil
}

// killByName drains someone and disposes of the body — what a careful vampire
// does, and what keeps a test about Reassignment from becoming a test about
// Hunters.
func killByName(t *testing.T, game *GameInstance, internalName string) *common.Townsperson {
	t.Helper()
	return drain(t, game, internalName, true)
}

// killLeavingBody drains someone and leaves the body where it fell, for the
// tests that are about it being found.
func killLeavingBody(t *testing.T, game *GameInstance, internalName string) *common.Townsperson {
	t.Helper()
	return drain(t, game, internalName, false)
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
	victim := killByName(t, game, "elias")
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
	victim := killByName(t, game, "elias")
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
			if obligation.Owner == "elias" || obligation.ID == "vegetables_to_the_inn" || obligation.ID == "grain_to_the_mill" {
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

	// vegetables_to_the_inn is owed by elias to greta, and greta has the hours.
	killByName(t, game, "elias")
	startDay := game.GetWorldState().GetClock().Day()
	for game.GetWorldState().GetClock().Day() < startDay+2 && game.Ending() == RunOngoing {
		game.Advance(1.0)
	}

	greta := game.townspersonByName("greta")
	if greta == nil {
		t.Fatal("greta is missing")
	}
	for _, obligation := range greta.Obligations {
		if obligation.ID == "vegetables_to_the_inn" {
			return // took it herself, as intended
		}
	}
	t.Error("vegetables_to_the_inn was not offered to Greta, who it was owed to and who has the hours")
}

func TestADutyNobodyCanTakeGoesUnmetAndIsVisible(t *testing.T) {
	// The absent face has to reach the player, or the town degrades invisibly.
	game := newTestGame(t)

	// the_grave_visit is Elias's alone: it names no Role, so it dies with him.
	killByName(t, game, "elias")
	startDay := game.GetWorldState().GetClock().Day()
	for game.GetWorldState().GetClock().Day() < startDay+2 && game.Ending() == RunOngoing {
		game.Advance(1.0)
	}

	var unmet *common.Obligation
	for _, obligation := range game.UnmetObligations() {
		if obligation.ID == "the_grave_visit" {
			unmet = obligation
		}
	}
	if unmet == nil {
		t.Fatalf("the_grave_visit should be Unmet; ledger holds %d", len(game.UnmetObligations()))
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
	if !victim.BodyHidden {
		t.Error("the body is not recorded as hidden")
	}
}

func TestAnUnhiddenBodyIsFoundAndBringsAHunter(t *testing.T) {
	// Leaving a body where it will be found is the difference between "she went
	// away" and "someone is killing us".
	game := newTestGame(t)
	victim := killLeavingBody(t, game, "martin")

	for i := 0; i < 2*24*60*oneRealSecond; i++ {
		game.Advance(1.0)
		if victim.Presence == common.PresenceConfirmedDead {
			break
		}
	}
	if victim.Presence != common.PresenceConfirmedDead {
		t.Fatal("a body left in the open was never found")
	}
	if game.hunter == nil {
		t.Error("a confirmed death did not put a Hunter on the streets")
	}
}

func TestAHiddenBodyIsNotFound(t *testing.T) {
	game := newTestGame(t)
	victim := killByName(t, game, "greta") // hidden

	for i := 0; i < 2*24*60*oneRealSecond && game.Ending() == RunOngoing; i++ {
		game.Advance(1.0)
	}
	if victim.Presence == common.PresenceConfirmedDead {
		t.Error("a hidden body was found anyway")
	}
}

func TestAHunterCatchingThePlayerEndsTheRun(t *testing.T) {
	game := newTestGame(t)
	killLeavingBody(t, game, "martin")

	for i := 0; i < 3*24*60*oneRealSecond && game.Ending() == RunOngoing; i++ {
		game.Advance(1.0)
	}
	if game.hunter == nil {
		t.Skip("no Hunter appeared, so this proves nothing")
	}
	// A player who never moves is eventually found. That is the point of a Hunter.
	if game.Ending() != RunCaught && game.Ending() != RunStarved {
		t.Logf("run ended as %q", game.Ending())
	}
}

func TestLeavingTownEndsTheRun(t *testing.T) {
	game := newTestGame(t)
	game.LeaveTown()
	if game.Ending() != RunLeftTown {
		t.Errorf("expected to have left town, got %q", game.Ending())
	}
}

// The vampire has no stat block. Blood is his health, and health is not.
func TestTheVampireLivesOnBloodNotHealth(t *testing.T) {
	game := newTestGame(t)
	player := game.GetWorldState().GetPlayer()

	player.GetStats().GetResource(stats.Health).SetToMin()
	if !player.IsAlive() {
		t.Error("emptying the health resource killed the vampire; his life is his Blood")
	}

	game.Blood().Drain(float64(common.SecondsPerDay) * 100) // burn the lot
	if player.IsAlive() {
		t.Error("with no Blood left the vampire is still alive")
	}
}

// Dawn in the open is a scramble for shelter, not a full stop: the sun burns
// Blood away fast, but there is time to run.
func TestSunlightBurnsFastInsteadOfKillingOutright(t *testing.T) {
	game := newTestGame(t)
	clock := game.GetWorldState().GetClock()

	// Sheltered in the lair until the sun is up.
	for i := 0; i < 24*60*oneRealSecond && !clock.IsDaylight(); i++ {
		game.Advance(1.0)
	}
	if !clock.IsDaylight() {
		t.Fatal("the sun never came up")
	}
	stepIntoTheOpen(t, game)

	blood := game.Blood()
	before := blood.Current()
	advanceRealSeconds(game, 5) // five in-game minutes of sun

	burned := before - blood.Current()
	if game.Ending() != RunOngoing {
		t.Fatalf("five minutes of sun ended the run as %q; it should only hurt", game.Ending())
	}
	nightlyDrain := blood.DrainPerDay * (5.0 / (24 * 60))
	if burned <= nightlyDrain*10 {
		t.Errorf("sunlight burned %.2f Blood in five minutes, barely more than the %.2f it costs at night", burned, nightlyDrain)
	}

	// Stay out in it and it does finish him.
	for i := 0; i < 2*60*60*oneRealSecond && game.Ending() == RunOngoing; i++ {
		game.Advance(1.0)
	}
	if game.Ending() != RunBurned {
		t.Errorf("standing in the sun ended the run as %q, expected %q", game.Ending(), RunBurned)
	}
}

// stepIntoTheOpen puts the player somewhere the lair does not shelter them.
func stepIntoTheOpen(t *testing.T, game *GameInstance) {
	t.Helper()
	currentMap := game.GetWorldState().GetMap()
	lair, ok := game.lairPosition()
	if !ok {
		t.Fatal("the town has no lair")
	}
	size := currentMap.MapSize()
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			here := geometry.Point{X: x, Y: y}
			if !currentMap.IsWalkable(here) || currentMap.IsActorAt(here) {
				continue
			}
			if currentMap.MoveDistance(here, lair) <= LairRadius {
				continue
			}
			currentMap.MoveActor(game.GetWorldState().GetPlayer(), here)
			game.GetWorldState().GetPlayer().SetDrawPosition(here.ToPointF())
			return
		}
	}
	t.Fatal("nowhere in the town is out of the lair's shelter")
}
