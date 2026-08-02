package game

import (
	"rpg/common"
	"rpg/geometry"
	"strings"
	"testing"
)

func allObligations(game *GameInstance) []*common.Obligation {
	var all []*common.Obligation
	for _, actor := range townspeople(game) {
		all = append(all, actor.GetTownsperson().Obligations...)
	}
	return all
}

func TestTownHasObligations(t *testing.T) {
	game := newTestGame(t)
	if len(allObligations(game)) == 0 {
		t.Fatal("no Obligations loaded; the town has nothing to propagate")
	}
}

func TestEveryObligationIsAuthoredTwice(t *testing.T) {
	// The absent face is the one that gets skipped under pressure, and skipping
	// it makes the simulation degrade invisibly.
	game := newTestGame(t)
	for _, obligation := range allObligations(game) {
		if obligation.Performed == "" {
			t.Errorf("%s has no performed appearance", obligation.ID)
		}
		if obligation.Absent == "" {
			t.Errorf("%s has no absent appearance", obligation.ID)
		}
	}
}

func TestObligationsPointAtPeopleAndAtTheTown(t *testing.T) {
	game := newTestGame(t)
	toPerson, toTown := 0, 0
	for _, obligation := range allObligations(game) {
		if obligation.OwedToTheTown() {
			toTown++
		} else {
			toPerson++
		}
	}
	if toPerson == 0 {
		t.Error("no Obligation points at a Townsperson, so Reassignment can never try them first")
	}
	if toTown == 0 {
		t.Error("no Obligation points at the town")
	}
}

func TestSomeObligationsCannotBeInherited(t *testing.T) {
	// At least one duty must die with its holder, or nothing ever fails
	// permanently and the player can never watch the town lose something.
	game := newTestGame(t)
	orphanable := 0
	for _, obligation := range allObligations(game) {
		if obligation.SuitableRole == "" {
			orphanable++
		}
	}
	if orphanable == 0 {
		t.Error("every Obligation has a Role that can inherit it; nothing can be lost for good")
	}
}

func TestSchedulesAreDerivedNotAuthored(t *testing.T) {
	// Giving someone a new Obligation must change their day, with no code
	// written per Role. This is the property ADR 0003 rests on.
	game := newTestGame(t)
	people := townspeople(game)

	var subject *common.Townsperson
	for _, actor := range people {
		if actor.GetTownsperson().FreeSecondsPerDay() > 2*3600 {
			subject = actor.GetTownsperson()
			break
		}
	}
	if subject == nil {
		t.Fatal("nobody has spare capacity to take on a duty")
	}

	before := len(subject.Schedule())
	subject.TakeObligation(&common.Obligation{
		ID:          "test_duty",
		StartSecond: 22 * 3600,
		EndSecond:   23 * 3600,
		At:          geometry.Point{X: 17, Y: 11},
		Performed:   "does the thing",
		Absent:      "the thing goes undone",
	})
	subject.RecomputeSchedule()
	after := len(subject.Schedule())

	if after != before+1 {
		t.Errorf("taking an Obligation should add an entry to the day: %d -> %d", before, after)
	}
	if entry, ok := subject.Schedule().EntryAt(22*3600 + 60); !ok || !entry.IsObligation() {
		t.Error("the new duty does not appear in the day at the hour it was given")
	}
}

func TestObligationsConsumeSpareCapacity(t *testing.T) {
	// Capacity is what will limit Reassignment, so a duty has to cost its holder
	// the hours it takes.
	game := newTestGame(t)
	var subject *common.Townsperson
	for _, actor := range townspeople(game) {
		if actor.GetTownsperson().FreeSecondsPerDay() > 2*3600 {
			subject = actor.GetTownsperson()
			break
		}
	}
	if subject == nil {
		t.Fatal("nobody has spare capacity")
	}

	before := subject.FreeSecondsPerDay()
	subject.TakeObligation(&common.Obligation{
		ID: "one_hour", StartSecond: 22 * 3600, EndSecond: 23 * 3600,
		Performed: "x", Absent: "y",
	})
	after := subject.FreeSecondsPerDay()

	if before-after != 3600 {
		t.Errorf("an hour-long duty should cost an hour of capacity: %d -> %d", before, after)
	}
}

func TestReleasingObligationsClearsTheDay(t *testing.T) {
	// This is what happens when a Townsperson is found to be Absent.
	game := newTestGame(t)
	var subject *common.Townsperson
	for _, actor := range townspeople(game) {
		if len(actor.GetTownsperson().Obligations) > 0 {
			subject = actor.GetTownsperson()
			break
		}
	}
	if subject == nil {
		t.Fatal("nobody owes anything")
	}

	released := subject.ReleaseObligations()
	if len(released) == 0 {
		t.Fatal("releasing returned nothing")
	}
	if len(subject.Obligations) != 0 {
		t.Error("Obligations remain after release")
	}
	for _, entry := range subject.Schedule() {
		if entry.IsObligation() {
			t.Errorf("the day still contains %s after release", entry.Activity)
		}
	}
}

func TestTownspeopleActuallyPerformObligations(t *testing.T) {
	// The point of the derived Schedule: a duty is somewhere the Townsperson
	// physically goes, at the hour it is owed.
	game := newTestGame(t)
	clock := game.GetWorldState().GetClock()

	performed := map[string]bool{}
	for i := 0; i < 26*60*oneRealSecond; i++ {
		game.Advance(1.0)
		for _, actor := range townspeople(game) {
			entry, ok := actor.GetTownsperson().Schedule().EntryAt(clock.SecondsSinceMidnight())
			if ok && entry.IsObligation() && game.hasArrivedAt(actor, entry.At) {
				performed[entry.Obligation.ID] = true
			}
		}
	}

	total := len(allObligations(game))
	if len(performed) < total {
		var missing []string
		for _, obligation := range allObligations(game) {
			if !performed[obligation.ID] {
				missing = append(missing, obligation.ID)
			}
		}
		t.Errorf("%d of %d Obligations were performed over a full Day; never reached: %v",
			len(performed), total, missing)
	}
}

func TestPerformedObligationsReachThePlayer(t *testing.T) {
	// An Obligation being performed is the observable event the design turns on.
	// If it never reaches the player there is nothing to infer from.
	game := newTestGame(t)

	for i := 0; i < 26*60*oneRealSecond; i++ {
		game.Advance(1.0)
	}

	log := game.GetLog()
	if len(log) == 0 {
		t.Fatal("a full Day produced no messages at all")
	}

	seen := 0
	for _, obligation := range allObligations(game) {
		for _, line := range log {
			if strings.Contains(line, obligation.Performed) {
				seen++
				break
			}
		}
	}
	if seen == 0 {
		t.Errorf("no Obligation was ever announced to the player over a full Day (%d log lines)", len(log))
	}
}
