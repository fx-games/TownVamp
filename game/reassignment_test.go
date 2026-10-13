package game

import (
	"rpg/common"
	"testing"
)

// Reassignment is the town repairing itself, and the only thing a player ever
// sees of it is the message saying it happened. So the thing worth testing is
// not that a duty changes hands — it is that the message and the world agree.

func TestADutyHandedOverIsActuallyInTheTakersDay(t *testing.T) {
	// Capacity was once counted in aggregate, so a duty could be handed to
	// somebody whose Routine already covered its hours. DeriveSchedule then
	// dropped it, and the duty was owed by one person, done by nobody, and absent
	// from the unmet ledger too — the town announcing a repair that never
	// happened, which is worse than the town admitting a loss.
	game := newTestGame(t)
	clock := game.GetWorldState().GetClock()

	day := clock.Day()
	for _, person := range game.townspeople {
		if person.InternalName == "elias" || person.InternalName == "bram" {
			person.GoAbsent(day)
		}
	}

	for clock.Day() == day {
		advanceRealSeconds(game, 30)
	}
	advanceRealSeconds(game, 5)

	for _, person := range game.townspeople {
		for _, obligation := range person.Obligations {
			found := false
			for _, entry := range person.Schedule() {
				if entry.Obligation == obligation {
					found = true
				}
			}
			if !found {
				t.Errorf("%s owes %s but it is nowhere in their day, so nobody will do it",
					person.InternalName, obligation.ID)
			}
		}
	}
}

func TestNothingVanishesBetweenItsHolderAndTheUnmetLedger(t *testing.T) {
	// Every Obligation the town starts with must end the run either held by
	// somebody or recorded as unmet. A duty in neither place is one the Epilogue
	// will never mention, which makes a consequence the player cannot find out
	// about.
	game := newTestGame(t)
	clock := game.GetWorldState().GetClock()

	authored := map[string]bool{}
	for _, person := range game.townspeople {
		for _, obligation := range person.Obligations {
			authored[obligation.ID] = true
		}
	}

	day := clock.Day()
	for _, person := range game.townspeople {
		if person.Role == common.RoleFarmer {
			person.GoAbsent(day)
		}
	}
	for clock.Day() == day {
		advanceRealSeconds(game, 30)
	}
	advanceRealSeconds(game, 5)

	accounted := map[string]bool{}
	for _, person := range game.townspeople {
		if person.IsAbsent() {
			continue
		}
		for _, obligation := range person.Obligations {
			accounted[obligation.ID] = true
		}
	}
	for _, obligation := range game.UnmetObligations() {
		accounted[obligation.ID] = true
	}

	for id := range authored {
		if !accounted[id] {
			t.Errorf("%s is held by nobody and recorded as unmet by nobody", id)
		}
	}
}

func TestTheFoodChainCanSurviveLosingAnyOneProducer(t *testing.T) {
	// The design wants early deaths absorbed and later ones catastrophic. That
	// only works if a duty has somewhere to go, and having free hours somewhere in
	// the day is not the same as being free during the hour the duty needs. Four
	// farmers all working 06:00 to 12:00 have plenty of spare capacity and no
	// ability to cover each other at all.
	game := newTestGame(t)

	for _, person := range game.townspeople {
		for _, obligation := range person.Obligations {
			if obligation.OwedToTheTown() || obligation.SuitableRole == "" {
				continue // private acts are meant to die with their owner
			}
			holders := 0
			for _, other := range game.townspeople {
				if other.Role == obligation.SuitableRole {
					holders++
				}
			}
			if holders < 2 {
				continue // the only smith in town cannot be covered by anybody
			}

			// The same two doors Reassignment tries: the person it was owed to,
			// then anyone of the Role.
			covered := false
			for _, candidate := range game.townspeople {
				if candidate == person {
					continue
				}
				if candidate.Role != obligation.SuitableRole &&
					candidate.InternalName != obligation.Towards {
					continue
				}
				if candidate.CanFitObligation(obligation) &&
					candidate.FreeSecondsPerDay() >= obligation.DurationSeconds() {
					covered = true
				}
			}
			if !covered {
				t.Errorf("no other %s is free during %s, so losing %s ends it permanently",
					obligation.SuitableRole, obligation.ID, person.InternalName)
			}
		}
	}
}
