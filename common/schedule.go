package common

import (
	"fmt"
	"rpg/geometry"
	"sort"
)

// ScheduleEntry is one block of a Townsperson's actual day. Entries coming from
// the Routine are fixed and react to nothing; entries coming from an Obligation
// are what change when a Townsperson dies and their duties move.
type ScheduleEntry struct {
	StartSecond int
	EndSecond   int
	Activity    string
	At          geometry.Point
	Obligation  *Obligation // nil for Routine entries
}

func (e ScheduleEntry) Covers(secondsSinceMidnight float64) bool {
	at := int(secondsSinceMidnight)
	if e.StartSecond <= e.EndSecond {
		return at >= e.StartSecond && at < e.EndSecond
	}
	return at >= e.StartSecond || at < e.EndSecond
}

func (e ScheduleEntry) IsObligation() bool { return e.Obligation != nil }

// Schedule is a Townsperson's day: their Routine, with their Obligations
// resolved into the hours the Routine leaves free. It is derived rather than
// authored, which is why moving an Obligation changes what the town visibly does
// without any code being written per Role.
type Schedule []ScheduleEntry

func (s Schedule) EntryAt(secondsSinceMidnight float64) (ScheduleEntry, bool) {
	// Obligations win over Routine where they overlap: a duty is what a
	// Townsperson has agreed to interrupt their own day for.
	var fallback ScheduleEntry
	foundFallback := false
	for _, entry := range s {
		if !entry.Covers(secondsSinceMidnight) {
			continue
		}
		if entry.IsObligation() {
			return entry, true
		}
		fallback = entry
		foundFallback = true
	}
	return fallback, foundFallback
}

// DeriveSchedule resolves Obligations into the hours a Routine leaves free.
// An Obligation that will not fit is reported rather than silently dropped —
// an unperformable duty is an authoring mistake, not a simulation outcome.
func DeriveSchedule(routine Routine, obligations []*Obligation) (Schedule, []error) {
	var schedule Schedule
	var problems []error

	for _, slot := range routine {
		schedule = append(schedule, ScheduleEntry{
			StartSecond: slot.StartSecond,
			EndSecond:   slot.EndSecond,
			Activity:    slot.Activity,
			At:          slot.At,
		})
	}

	for _, obligation := range obligations {
		if clash, found := routine.SlotAt(float64(obligation.StartSecond)); found {
			problems = append(problems, fmt.Errorf(
				"obligation %s at %s runs into the routine slot %q",
				obligation.ID, clockString(obligation.StartSecond), clash.Activity))
			continue
		}
		schedule = append(schedule, ScheduleEntry{
			StartSecond: obligation.StartSecond,
			EndSecond:   obligation.EndSecond,
			Activity:    obligation.ID,
			At:          obligation.At,
			Obligation:  obligation,
		})
	}

	sort.SliceStable(schedule, func(i, j int) bool {
		return schedule[i].StartSecond < schedule[j].StartSecond
	})
	return schedule, problems
}

// CheckWalkable reports the places in a Schedule a Townsperson cannot physically
// get to in the time allowed. Because Schedules are derived rather than authored,
// nothing otherwise guarantees the day is achievable — this is the cost of
// deriving them, and it is a load-time concern rather than a runtime surprise.
func (s Schedule) CheckWalkable(travelSeconds func(from, to geometry.Point) (float64, bool)) []error {
	var problems []error
	for i := range s {
		from := s[i]
		to := s[(i+1)%len(s)]
		if from.At == to.At {
			continue
		}
		needed, reachable := travelSeconds(from.At, to.At)
		if !reachable {
			problems = append(problems, fmt.Errorf(
				"cannot walk from %s at %v to %s at %v at all", from.Activity, from.At, to.Activity, to.At))
			continue
		}
		// A Townsperson sets off when the previous entry ends and may arrive part
		// way into the next one — being a little late is ordinary. The day is only
		// infeasible if they could not arrive before the entry is over.
		available := float64(gapSeconds(from.EndSecond, to.StartSecond) + durationSeconds(to.StartSecond, to.EndSecond))
		if needed > available {
			problems = append(problems, fmt.Errorf(
				"walking from %s to %s needs %.0fs, but %s is over after %.0fs",
				from.Activity, to.Activity, needed, to.Activity, available))
		}
	}
	return problems
}

// gapSeconds is the time between one entry ending and the next beginning. An
// entry that starts the moment the previous ends leaves no travel time, so the
// walk has to be instant — which is exactly the mistake worth reporting.
func gapSeconds(endOfPrevious, startOfNext int) int {
	if startOfNext >= endOfPrevious {
		return startOfNext - endOfPrevious
	}
	return (SecondsPerDay - endOfPrevious) + startOfNext
}

func durationSeconds(start, end int) int {
	if start <= end {
		return end - start
	}
	return (SecondsPerDay - start) + end
}

func clockString(secondsSinceMidnight int) string {
	return fmt.Sprintf("%02d:%02d", secondsSinceMidnight/3600, (secondsSinceMidnight%3600)/60)
}
