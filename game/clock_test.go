package game

import (
	"rpg/common"
	"testing"
)

// oneRealSecond is how many fixed-update ticks make a real second, and therefore
// how many in-game seconds pass at the default Time Multiplier.
const oneRealSecond = 60

func advanceRealSeconds(game *GameInstance, seconds int) {
	for i := 0; i < seconds*oneRealSecond; i++ {
		game.Advance(1.0)
	}
}

func TestTimeAdvancesWithoutThePlayerActing(t *testing.T) {
	game := newTestGame(t)
	clock := game.GetWorldState().GetClock()

	before := clock.SecondsSinceMidnight()
	advanceRealSeconds(game, 60) // one real minute == one in-game hour
	after := clock.SecondsSinceMidnight()

	elapsed := after - before
	if after < before {
		elapsed = after + common.SecondsPerDay - before
	}
	if elapsed < 3500 || elapsed > 3700 {
		t.Errorf("one real minute should be one in-game hour, got %.0f in-game seconds", elapsed)
	}
}

func TestADayIsTwentyFourRealMinutes(t *testing.T) {
	game := newTestGame(t)
	clock := game.GetWorldState().GetClock()

	startDay := clock.Day()
	advanceRealSeconds(game, 24*60)

	if clock.Day() != startDay+1 {
		t.Errorf("24 real minutes should be one Day, got %d days", clock.Day()-startDay)
	}
}

func TestNightBecomesDay(t *testing.T) {
	game := newTestGame(t)
	clock := game.GetWorldState().GetClock()

	// A run opens at dusk, so the sun should be down.
	if clock.IsDaylight() {
		t.Fatalf("run should open at night, clock says %s", clock)
	}
	untilDawn := clock.SecondsUntilDawn()
	if untilDawn <= 0 {
		t.Fatal("night should report time remaining until dawn")
	}

	advanceRealSeconds(game, 11*60) // dusk is 20:00, dawn is 06:00 — ten hours
	if !clock.IsDaylight() {
		t.Errorf("sun should be up by %s", clock)
	}
	if clock.SecondsUntilDawn() != 0 {
		t.Error("daylight should report no time remaining until dawn")
	}
}

func TestTimeMultiplierScalesTheWholeSimulation(t *testing.T) {
	// The same number of in-game hours must produce the same world at every
	// multiplier — only the real time taken differs. If the multiplier scaled the
	// calendar without scaling movement, Schedules would break at x4 and nowhere
	// else, which is the worst kind of defect to chase.
	positionsAt := func(multiplier float64) (int, int, map[string]string) {
		game := newTestGame(t)
		game.GetWorldState().SetTimeMultiplier(multiplier)
		clock := game.GetWorldState().GetClock()

		// Six in-game hours, however long that takes in real time.
		ticks := int(float64(6*60*oneRealSecond) / multiplier)
		for i := 0; i < ticks; i++ {
			game.Advance(1.0)
		}
		placed := map[string]string{}
		for _, actor := range game.GetWorldState().GetMap().Actors() {
			placed[actor.Name()] = actor.MapPosition().String()
		}
		return clock.Hour(), clock.Minute(), placed
	}

	baseHour, baseMinute, basePlaces := positionsAt(1)
	for _, multiplier := range []float64{2, 4} {
		hour, minute, places := positionsAt(multiplier)
		if hour != baseHour || minute != baseMinute {
			t.Errorf("x%.0f reached %02d:%02d, x1 reached %02d:%02d", multiplier, hour, minute, baseHour, baseMinute)
		}
		for name, at := range basePlaces {
			if places[name] != at {
				t.Errorf("x%.0f put %s at %s, x1 put them at %s", multiplier, name, places[name], at)
			}
		}
	}
}
