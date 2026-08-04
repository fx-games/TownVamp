package util

import (
	"testing"
	"time"
)

// The loop used to convert fractional seconds straight to a time.Duration,
// which truncated every sleep to zero and spun a core flat. Three frames at
// 50 FPS cannot take less than two frame periods unless the sleep is gone.
func TestLoopSleepsBetweenFrames(t *testing.T) {
	frames := 0
	terminationCode := 0
	loop := CreateLoop(LoopConfig{
		TargetFPS:       50,
		IdleThreshold:   5,
		CurrentTimeFunc: func() float64 { return float64(time.Now().UnixNano()) * 1e-9 },
		ProcessInputFunc: func() int {
			frames++
			if frames > 3 {
				return 1
			}
			return 0
		},
		UpdateFunc: func(dt float64) {},
		RenderFunc: func() {},
	}, &terminationCode)

	start := time.Now()
	loop()

	if elapsed := time.Since(start); elapsed < 40*time.Millisecond {
		t.Fatalf("three frames at 50 FPS took %v, the loop is not sleeping", elapsed)
	}
}
