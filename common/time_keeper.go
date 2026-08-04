package common

// BaseSecondsPerStep is how long one cell costs someone of ordinary speed, and
// so what finally decides how fast anyone crosses the town: a minute a cell,
// about an hour end to end. Schedules are only achievable if this and the
// distances on the map agree, so expect to tune it against the real town.
const BaseSecondsPerStep = 60.0

// SecondsPerStep is what a step costs the given walker. Speed is a multiplier
// on the ordinary pace, where ten is ordinary.
func SecondsPerStep(moveSpeed int) float64 {
	if moveSpeed <= 0 {
		return BaseSecondsPerStep
	}
	return BaseSecondsPerStep * 10.0 / float64(moveSpeed)
}

// TimedBeing is anyone the world moves without being asked. There is no action
// economy behind this any more: the world runs on a clock, so a being simply
// says when it may act again and the TimeKeeper honours it.
type TimedBeing interface {
	IsAlive() bool

	// ReadyAt is the moment, in in-game seconds since the run began, from which
	// this being may act again.
	ReadyAt() float64
	SetReadyAt(second float64)
}

type TimeKeeper struct {
	elapsedSeconds float64
	timed          []TimedBeing
	nextAction     func(being TimedBeing) (secondsUsed float64, act func())
	onRemoveTimed  func(thing TimedBeing)
}

func NewTimeKeeper(nextAction func(being TimedBeing) (secondsUsed float64, act func())) *TimeKeeper {
	return &TimeKeeper{
		timed:      make([]TimedBeing, 0),
		nextAction: nextAction,
	}
}

func (t *TimeKeeper) SetTimed(timedBeings []TimedBeing) {
	t.timed = timedBeings
	for _, being := range timedBeings {
		being.SetReadyAt(t.elapsedSeconds)
	}
}

// AdvanceBySeconds moves the world forward by some number of in-game seconds.
// This is the only clock: it runs from the fixed-timestep update, not from a
// player action, so the town keeps moving whether or not the player does.
func (t *TimeKeeper) AdvanceBySeconds(inGameSeconds float64) {
	t.elapsedSeconds += inGameSeconds

	// Iterated backwards so that a being dying mid-loop can be removed without
	// disturbing the indices still to be visited.
	for i := len(t.timed) - 1; i >= 0; i-- {
		being := t.timed[i]
		if !being.IsAlive() {
			t.removeAtIndex(i)
			continue
		}
		for being.ReadyAt() <= t.elapsedSeconds {
			secondsUsed, act := t.nextAction(being)
			if act == nil {
				break // nothing they can do yet; ask again next slice
			}
			act()
			if !being.IsAlive() {
				t.removeAtIndex(i)
				break
			}
			if secondsUsed <= 0 {
				break // a free action would spin here forever
			}
			being.SetReadyAt(being.ReadyAt() + secondsUsed)
		}
	}
}

func (t *TimeKeeper) AddTimed(actor TimedBeing) {
	// From now, not from the beginning of the run, or a being added to a world
	// already hours old would owe the loop every step it missed.
	actor.SetReadyAt(t.elapsedSeconds)
	t.timed = append(t.timed, actor)
}

func (t *TimeKeeper) removeAtIndex(i int) {
	timedBeing := t.timed[i]
	t.timed = append(t.timed[:i], t.timed[i+1:]...)
	if t.onRemoveTimed != nil {
		t.onRemoveTimed(timedBeing)
	}
}

func (t *TimeKeeper) SetOnRemoveTimed(onRemoveTimed func(thing TimedBeing)) {
	t.onRemoveTimed = onRemoveTimed
}

func (t *TimeKeeper) ClearTrackedBeings(shouldRemoveBeing func(being TimedBeing) bool) []TimedBeing {
	var removed []TimedBeing
	for i := len(t.timed) - 1; i >= 0; i-- {
		if shouldRemoveBeing(t.timed[i]) {
			removed = append(removed, t.timed[i])
			t.removeAtIndex(i)
		}
	}
	return removed
}

func (t *TimeKeeper) GetBeingsByFilter(keep func(being TimedBeing) bool) []TimedBeing {
	var result []TimedBeing
	for i := len(t.timed) - 1; i >= 0; i-- {
		timedBeing := t.timed[i]
		if timedBeing.IsAlive() && keep(timedBeing) {
			result = append(result, timedBeing)
		}
	}
	return result
}
