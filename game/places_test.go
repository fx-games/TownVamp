package game

import (
	"fmt"
	"strings"
	"testing"
)

// TestProbeARunWithAKilling plays the core loop end to end and prints what the
// player would have seen. Not an assertion — a way to read the simulation.
func TestProbeARunWithAKilling(t *testing.T) {
	game := newTestGame(t)
	clock := game.GetWorldState().GetClock()

	logFrom := 0
	report := func(label string) {
		log := game.GetLog()
		var b strings.Builder
		b.WriteString("\n=== " + label + " (" + clock.String() + ") ===")
		for _, line := range log[logFrom:] {
			b.WriteString("\n  " + line)
		}
		logFrom = len(log)
		t.Log(b.String())
	}

	// Night one: watch until the town is asleep, then take the farmer, who feeds
	// the inn and keeps the mill in grain.
	for clock.Hour() != 2 {
		game.Advance(1.0)
	}
	victim := killByName(t, game, "elias")
	// Back to the lair before dawn, or the sun ends the run here.
	if lair, ok := game.lairPosition(); ok {
		game.GetWorldState().GetPlayer().SetMapPosition(lair)
	}
	report("night one — drained " + victim.InternalName)

	// Three more days of town.
	startDay := clock.Day()
	for clock.Day() < startDay+3 && game.Ending() == RunOngoing {
		game.Advance(1.0)
	}
	report("three days later")

	if game.Ending() == RunOngoing {
		game.LeaveTown()
	}
	var e strings.Builder
	e.WriteString("\n--- epilogue ---")
	for _, slide := range game.Epilogue() {
		e.WriteString("\n\n  == " + slide.Title + " ==")
		for _, line := range slide.Lines {
			e.WriteString("\n  " + line)
		}
	}
	t.Log(e.String())

	var b strings.Builder
	b.WriteString("\n--- the town now ---")
	for _, person := range game.townspeople {
		b.WriteString(fmt.Sprintf("\n  %-10s %-10s %-16s %d duties, %.1fh spare",
			person.InternalName, person.Role, person.Presence,
			len(person.Obligations), float64(person.FreeSecondsPerDay())/3600))
	}
	b.WriteString(fmt.Sprintf("\n  unmet obligations: %d", len(game.UnmetObligations())))
	for _, obligation := range game.UnmetObligations() {
		b.WriteString("\n    - " + obligation.ID + ": " + obligation.Absent)
	}
	b.WriteString(fmt.Sprintf("\n  blood: %.0f%%  ending: %q", game.Blood().Current(), game.Ending()))
	t.Log(b.String())
}
