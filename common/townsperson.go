package common

import (
	"fmt"
	"rpg/geometry"
	"sort"
	"strconv"
	"strings"
)

// Role is what a Townsperson does for the town: public, structural, and visible
// from watching them work. The set is closed, and it doubles as Reassignment's
// compatibility rule — an Obligation can pass to a Townsperson of a fitting Role
// and to no one else.
type Role string

const (
	RoleFarmer    Role = "farmer"    // produces food
	RoleSmith     Role = "smith"     // makes and repairs metal tools
	RoleBuilder   Role = "builder"   // carpenter, mason, wheelwright
	RoleTrader    Role = "trader"    // merchant, innkeeper, transport
	RoleAuthority Role = "authority" // lord, priest, guard, judge
)

func RoleFromString(name string) (Role, error) {
	switch Role(strings.ToLower(strings.TrimSpace(name))) {
	case RoleFarmer:
		return RoleFarmer, nil
	case RoleSmith:
		return RoleSmith, nil
	case RoleBuilder:
		return RoleBuilder, nil
	case RoleTrader:
		return RoleTrader, nil
	case RoleAuthority:
		return RoleAuthority, nil
	}
	return "", fmt.Errorf("unknown role %q", name)
}

// RoutineSlot is one block of a Townsperson's fixed day: be at this place,
// between these hours, doing this. Slots never react to anything.
type RoutineSlot struct {
	StartSecond int
	EndSecond   int
	Activity    string
	At          geometry.Point
}

func (s RoutineSlot) Covers(secondsSinceMidnight float64) bool {
	at := int(secondsSinceMidnight)
	if s.StartSecond <= s.EndSecond {
		return at >= s.StartSecond && at < s.EndSecond
	}
	return at >= s.StartSecond || at < s.EndSecond // wraps past midnight
}

func (s RoutineSlot) DurationSeconds() int {
	if s.StartSecond <= s.EndSecond {
		return s.EndSecond - s.StartSecond
	}
	return (SecondsPerDay - s.StartSecond) + s.EndSecond
}

// Routine is the fixed, authored part of a Townsperson's day — sleep, meals,
// home. Publicly legible: a glance at the street shows it. The hours it leaves
// free are that Townsperson's spare capacity, and capacity is the only thing
// that will limit Reassignment.
type Routine []RoutineSlot

func (r Routine) SlotAt(secondsSinceMidnight float64) (RoutineSlot, bool) {
	for _, slot := range r {
		if slot.Covers(secondsSinceMidnight) {
			return slot, true
		}
	}
	return RoutineSlot{}, false
}

// FreeSecondsPerDay is the Townsperson's spare capacity: how much of their day
// the Routine does not already claim. A Townsperson who sleeps long and eats
// slowly is fragile; one with a lean Routine is the town's shock absorber.
func (r Routine) FreeSecondsPerDay() int {
	claimed := 0
	for _, slot := range r {
		claimed += slot.DurationSeconds()
	}
	if claimed > SecondsPerDay {
		return 0
	}
	return SecondsPerDay - claimed
}

// Townsperson is an inhabitant of the town: a Role, an Identity, and a Routine.
// Obligations and the derived Schedule arrive in a later ticket.
type Townsperson struct {
	InternalName string
	Role         Role
	Identity     string
	Routine      Routine
}

func (t *Townsperson) String() string {
	return fmt.Sprintf("%s (%s)", t.InternalName, t.Role)
}

// ParseClockTime reads "HH:MM" into seconds since midnight.
func ParseClockTime(value string) (int, error) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("expected HH:MM, got %q", value)
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, fmt.Errorf("bad hour in %q", value)
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, fmt.Errorf("bad minute in %q", value)
	}
	return hour*3600 + minute*60, nil
}

// ParseRoutineSlot reads one authored line: "22:00-06:00 sleep 12,9".
// The place is either "x,y" or the name of a location on the map, resolved by
// the caller-supplied lookup.
func ParseRoutineSlot(value string, namedLocation func(string) (geometry.Point, bool)) (RoutineSlot, error) {
	fields := strings.Fields(value)
	if len(fields) < 3 {
		return RoutineSlot{}, fmt.Errorf("expected 'HH:MM-HH:MM activity place', got %q", value)
	}
	span := strings.Split(fields[0], "-")
	if len(span) != 2 {
		return RoutineSlot{}, fmt.Errorf("expected 'HH:MM-HH:MM', got %q", fields[0])
	}
	start, err := ParseClockTime(span[0])
	if err != nil {
		return RoutineSlot{}, err
	}
	end, err := ParseClockTime(span[1])
	if err != nil {
		return RoutineSlot{}, err
	}
	place, err := parsePlace(fields[2], namedLocation)
	if err != nil {
		return RoutineSlot{}, err
	}
	return RoutineSlot{StartSecond: start, EndSecond: end, Activity: fields[1], At: place}, nil
}

func parsePlace(value string, namedLocation func(string) (geometry.Point, bool)) (geometry.Point, error) {
	if coords := strings.Split(value, ","); len(coords) == 2 {
		x, xErr := strconv.Atoi(strings.TrimSpace(coords[0]))
		y, yErr := strconv.Atoi(strings.TrimSpace(coords[1]))
		if xErr == nil && yErr == nil {
			return geometry.Point{X: x, Y: y}, nil
		}
	}
	if namedLocation != nil {
		if pos, ok := namedLocation(value); ok {
			return pos, nil
		}
	}
	return geometry.Point{}, fmt.Errorf("place %q is neither coordinates nor a known location", value)
}

// SortRoutine puts slots in clock order so the authored file can be written in
// any order and still read sensibly.
func SortRoutine(routine Routine) {
	sort.SliceStable(routine, func(i, j int) bool {
		return routine[i].StartSecond < routine[j].StartSecond
	})
}
