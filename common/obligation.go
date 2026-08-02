package common

import (
	"fmt"
	"rpg/geometry"
)

// Obligation is a duty a Townsperson owes, pointing either at another
// Townsperson or at the town itself. It is the only thing this simulation
// propagates: when a Townsperson goes Absent their Obligations are voided, and
// what happens to them afterwards is the whole of the consequence model.
//
// Every Obligation is authored twice — what it looks like being performed, and
// what its absence looks like. The second face is the one that gets skipped
// under pressure, and skipping it makes the simulation degrade invisibly.
type Obligation struct {
	ID    string
	Owner string // internal name of the Townsperson who currently holds it

	// Towards is the internal name of the Townsperson this is owed to, or empty
	// when it is owed to the town. Reassignment tries this person first.
	Towards string

	// SuitableRole is who may inherit it once the person it points at has been
	// tried. Empty means nobody can: the Obligation dies with its holder.
	SuitableRole Role

	At          geometry.Point
	StartSecond int
	EndSecond   int

	// Performed is what a watcher sees when it happens. Absent is what the town
	// looks like when it does not.
	Performed string
	Absent    string
}

func (o *Obligation) DurationSeconds() int {
	if o.StartSecond <= o.EndSecond {
		return o.EndSecond - o.StartSecond
	}
	return (SecondsPerDay - o.StartSecond) + o.EndSecond
}

func (o *Obligation) Covers(secondsSinceMidnight float64) bool {
	at := int(secondsSinceMidnight)
	if o.StartSecond <= o.EndSecond {
		return at >= o.StartSecond && at < o.EndSecond
	}
	return at >= o.StartSecond || at < o.EndSecond
}

// OwedToTheTown reports whether nobody in particular is waiting on this.
func (o *Obligation) OwedToTheTown() bool {
	return o.Towards == ""
}

func (o *Obligation) String() string {
	if o.OwedToTheTown() {
		return fmt.Sprintf("%s (owed to the town)", o.ID)
	}
	return fmt.Sprintf("%s (owed to %s)", o.ID, o.Towards)
}
