package common

// WaitCondition names a moment worth skipping to. Waiting runs the simulation
// forward without drawing and stops the instant the condition holds, so skipping
// the dull hours never costs the player the thing they were waiting for.
type WaitCondition string

const (
	WaitUntilDawn WaitCondition = "dawn"
	WaitUntilDusk WaitCondition = "dusk"
)

// IsMet reports whether the moment has arrived.
func (w WaitCondition) IsMet(clock *WorldClock) bool {
	switch w {
	case WaitUntilDawn:
		return clock.IsDaylight()
	case WaitUntilDusk:
		return !clock.IsDaylight()
	}
	return true
}
