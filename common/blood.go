package common

import "fmt"

// Blood is the vampire's hunger clock. It drains continuously and never stops,
// and Killing is the only thing that refills it. It is the only mechanism
// forcing the player to act — without it they could watch the town forever and
// never choose anyone.
type Blood struct {
	current float64
	maximum float64

	// DrainPerDay is how much is lost over a full Day. Together with what a
	// Townsperson restores, this sets the length of a run and how many of the six
	// die in it. Expect to find it by playing rather than by reasoning.
	DrainPerDay float64

	// RestoredByAKill is what draining one Townsperson is worth.
	RestoredByAKill float64
}

func NewBlood() *Blood {
	return &Blood{
		current:         100,
		maximum:         100,
		DrainPerDay:     40, // roughly two and a half Days from full
		RestoredByAKill: 60,
	}
}

func (b *Blood) Drain(inGameSeconds float64) {
	b.current -= b.DrainPerDay * (inGameSeconds / float64(SecondsPerDay))
	if b.current < 0 {
		b.current = 0
	}
}

func (b *Blood) Feed() {
	b.current += b.RestoredByAKill
	if b.current > b.maximum {
		b.current = b.maximum
	}
}

func (b *Blood) Current() float64 { return b.current }
func (b *Blood) Maximum() float64 { return b.maximum }
func (b *Blood) IsEmpty() bool    { return b.current <= 0 }

// DaysRemaining is what the player needs in order to judge how many more nights
// of watching they can afford.
func (b *Blood) DaysRemaining() float64 {
	if b.DrainPerDay <= 0 {
		return 0
	}
	return b.current / b.DrainPerDay
}

func (b *Blood) String() string {
	return fmt.Sprintf("Blood: %.0f%%", (b.current/b.maximum)*100)
}
