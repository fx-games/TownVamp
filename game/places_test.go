package game

import (
	"fmt"
	"strings"
	"testing"
)

// TestProbeADayInTheTown is a readable trace of the town's day: where everyone is,
// hour by hour. Not an assertion — a way to look at the simulation.
func TestProbeADayInTheTown(t *testing.T) {
	game := newTestGame(t)
	clock := game.GetWorldState().GetClock()
	people := townspeople(game)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n%-8s", "time"))
	for _, actor := range people {
		b.WriteString(fmt.Sprintf("%-22s", actor.Name()))
	}
	lastHour := -1
	for i := 0; i < 25*60*oneRealSecond; i++ {
		game.Advance(1.0)
		if clock.Hour() == lastHour || clock.Minute() != 0 {
			continue
		}
		lastHour = clock.Hour()
		b.WriteString(fmt.Sprintf("\n%02d:00   ", clock.Hour()))
		for _, actor := range people {
			doing := "-"
			if slot, ok := actor.GetTownsperson().Routine.SlotAt(clock.SecondsSinceMidnight()); ok {
				doing = slot.Activity
			}
			b.WriteString(fmt.Sprintf("%-10s%-12v", doing, actor.MapPosition()))
		}
	}
	t.Log(b.String())
}
