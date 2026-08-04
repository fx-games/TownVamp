package gui

import (
	"strings"
	"testing"

	"rpg/common"
)

// The marker is the whole point of the window: it says which line of someone's
// day the hour on the clock is in. Exactly one line may carry it.
func TestTheCurrentActivityIsTheOnlyMarkedOne(t *testing.T) {
	person := &common.Townsperson{
		InternalName: "marta",
		Number:       4,
		Role:         common.RoleBuilder,
		Routine: common.Routine{
			{StartSecond: 22 * 3600, EndSecond: 6 * 3600, Activity: "sleep"},
			{StartSecond: 6 * 3600, EndSecond: 12 * 3600, Activity: "work"},
			{StartSecond: 12 * 3600, EndSecond: 22 * 3600, Activity: "home"},
		},
	}
	person.RecomputeSchedule()

	lines := npcWindowLines(person, "Marta", 8*3600)

	var marked []string
	for _, line := range lines {
		if strings.HasPrefix(line.Text, "> ") {
			marked = append(marked, line.Text)
		}
	}
	if len(marked) != 1 {
		t.Fatalf("%d lines are marked as current, want exactly 1: %v", len(marked), marked)
	}
	if !strings.Contains(marked[0], "work") {
		t.Errorf("at 08:00 the marked line is %q, want the one covering work", marked[0])
	}
}

// A Routine slot that wraps past midnight still covers the small hours, and the
// header has to survive being off the front of the day.
func TestTheMarkerFollowsAWrappingSlotPastMidnight(t *testing.T) {
	person := &common.Townsperson{
		InternalName: "martin",
		Number:       5,
		Role:         common.RoleFarmer,
		Routine: common.Routine{
			{StartSecond: 22 * 3600, EndSecond: 6 * 3600, Activity: "sleep"},
			{StartSecond: 6 * 3600, EndSecond: 22 * 3600, Activity: "mill"},
		},
	}
	person.RecomputeSchedule()

	for _, line := range npcWindowLines(person, "Martin", 2*3600) {
		if strings.HasPrefix(line.Text, "> ") {
			if !strings.Contains(line.Text, "sleep") {
				t.Errorf("at 02:00 the marked line is %q, want sleep", line.Text)
			}
			return
		}
	}
	t.Error("nothing is marked current at 02:00")
}

// Long text wraps at words and stays inside the window's columns.
func TestLinesWrapAtWordsWithinTheWindow(t *testing.T) {
	person := &common.Townsperson{
		InternalName: "greta",
		Number:       3,
		Role:         common.RoleTrader,
		Routine: common.Routine{
			{
				StartSecond: 12 * 3600,
				EndSecond:   13 * 3600,
				Activity:    "carries bowls out to every table in the common room without ever being asked for them",
			},
		},
	}
	person.RecomputeSchedule()

	lines := npcWindowLines(person, "Greta", 12*3600)
	wrapped := 0
	for _, line := range lines {
		if len(line.Text) > npcWindowColumns {
			t.Errorf("line runs to %d characters, window is %d: %q", len(line.Text), npcWindowColumns, line.Text)
		}
		if strings.HasPrefix(line.Text, strings.Repeat(" ", npcWindowActivityIndent)) {
			wrapped++
		}
	}
	if wrapped == 0 {
		t.Error("a line longer than the window never wrapped")
	}
}
