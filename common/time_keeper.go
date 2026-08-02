package common

type TimedBeing interface {
	AddEnergy(energy int)
	RemoveEnergy(used int)
	IsAlive() bool
	GetBaseSpeed() int
	GetEnergy() int
}

// SecondsPerAut fixes what an AUT is worth on the wall clock, which is what
// finally decides how fast anyone walks. A step costs 10 AUT, so at six seconds
// per AUT a Townsperson covers one tile per in-game minute — about a minute to
// cross a room, an hour to cross the town. Schedules are only achievable if this
// and the distances on the map agree, so expect to tune it against the real town.
const SecondsPerAut = 6.0

type TimeKeeper struct {
	autCounter         uint64
	timed              []TimedBeing
	hasEnergyForAction func(being TimedBeing) (int, func())
	onRemoveTimed      func(thing TimedBeing)

	// Time arrives in fractions of an AUT — a tick is one in-game second, an AUT
	// is six — while energy is whole numbers. The remainder is carried here.
	pendingAut float64
}

func NewTimeKeeper(hasEnergyForAction func(being TimedBeing) (int, func())) *TimeKeeper {
	return &TimeKeeper{
		timed:              make([]TimedBeing, 0),
		hasEnergyForAction: hasEnergyForAction,
	}
}

func (t *TimeKeeper) SetTimed(timedBeings []TimedBeing) {
	t.timed = timedBeings
}

// AdvanceBySeconds moves the world forward by some number of in-game seconds.
// This is the only clock: it runs from the fixed-timestep update, not from a
// player action, so the town keeps moving whether or not the player does.
func (t *TimeKeeper) AdvanceBySeconds(inGameSeconds float64) {
	t.pendingAut += inGameSeconds / SecondsPerAut
	wholeAut := int(t.pendingAut)
	if wholeAut <= 0 {
		return
	}
	t.pendingAut -= float64(wholeAut)
	t.advance(wholeAut)
}

func (t *TimeKeeper) advance(autElapsed int) {
	t.autCounter += uint64(autElapsed)

	// Iterated backwards so that a being dying mid-loop can be removed without
	// disturbing the indices still to be visited.
	for i := len(t.timed) - 1; i >= 0; i-- {
		if i >= len(t.timed) {
			continue
		}
		timedBeing := t.timed[i]
		if !timedBeing.IsAlive() {
			t.removeAtIndex(i)
			continue
		}
		addedEnergy := int((float64(autElapsed) / 10.0) * float64(timedBeing.GetBaseSpeed()))
		timedBeing.AddEnergy(addedEnergy)
		for energyUsed, action := t.hasEnergyForAction(timedBeing); timedBeing.IsAlive() && action != nil; energyUsed, action = t.hasEnergyForAction(timedBeing) {
			timedBeing.RemoveEnergy(energyUsed)
			if !timedBeing.IsAlive() {
				t.removeAtIndex(i)
				break
			}
			action()
		}
	}
}

// ElapsedAut is the total simulated time since the run began.
func (t *TimeKeeper) ElapsedAut() uint64 {
	return t.autCounter
}

func (t *TimeKeeper) AddTimed(actor TimedBeing) {
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
