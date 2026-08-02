package common

import "fmt"

const (
	secondsPerMinute = 60
	secondsPerHour   = 60 * secondsPerMinute
	SecondsPerDay    = 24 * secondsPerHour

	// Dawn and dusk bound the daylight hours. They are what makes Sunlight
	// dangerous and what a Routine is written against.
	DawnHour = 6
	DuskHour = 20
)

// WorldClock is the town's time of day. One tick of the fixed update is one
// in-game second at the default Time Multiplier, which makes a Day twenty-four
// minutes of real time.
type WorldClock struct {
	secondsSinceMidnight float64
	day                  int
}

func NewWorldClock(startHour int) *WorldClock {
	return &WorldClock{secondsSinceMidnight: float64(startHour * secondsPerHour)}
}

// Advance moves the clock on, rolling over at midnight and counting the Day.
// It reports how many day boundaries were crossed, because that is when
// Schedules are recomputed and Absence is counted.
func (c *WorldClock) Advance(inGameSeconds float64) int {
	c.secondsSinceMidnight += inGameSeconds
	daysPassed := 0
	for c.secondsSinceMidnight >= SecondsPerDay {
		c.secondsSinceMidnight -= SecondsPerDay
		c.day++
		daysPassed++
	}
	return daysPassed
}

func (c *WorldClock) Day() int { return c.day }

func (c *WorldClock) Hour() int {
	return int(c.secondsSinceMidnight) / secondsPerHour
}

func (c *WorldClock) Minute() int {
	return (int(c.secondsSinceMidnight) % secondsPerHour) / secondsPerMinute
}

// SecondsSinceMidnight is what a Routine slot is compared against.
func (c *WorldClock) SecondsSinceMidnight() float64 {
	return c.secondsSinceMidnight
}

func (c *WorldClock) IsDaylight() bool {
	hour := c.Hour()
	return hour >= DawnHour && hour < DuskHour
}

// SecondsUntilDawn is how long a vampire has before Sunlight kills him. It is
// zero while the sun is already up.
func (c *WorldClock) SecondsUntilDawn() float64 {
	if c.IsDaylight() {
		return 0
	}
	dawn := float64(DawnHour * secondsPerHour)
	if c.secondsSinceMidnight < dawn {
		return dawn - c.secondsSinceMidnight
	}
	return (SecondsPerDay - c.secondsSinceMidnight) + dawn
}

func (c *WorldClock) String() string {
	return fmt.Sprintf("Day %d, %02d:%02d", c.day+1, c.Hour(), c.Minute())
}
