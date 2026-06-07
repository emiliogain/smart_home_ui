package replay

import (
	"testing"
	"time"
)

func TestSleepBetweenRows_fixedEvery(t *testing.T) {
	start := time.Now()
	SleepBetweenRows(time.Minute, 0.05, 1, 0) // every 50ms wall
	elapsed := time.Since(start)
	if elapsed < 30*time.Millisecond || elapsed > 250*time.Millisecond {
		t.Fatalf("expected ~50ms sleep, got %v", elapsed)
	}
}

func TestSleepBetweenRows_playback(t *testing.T) {
	start := time.Now()
	SleepBetweenRows(100*time.Millisecond, 0, 2, 0) // half of 100ms
	elapsed := time.Since(start)
	if elapsed < 30*time.Millisecond || elapsed > 120*time.Millisecond {
		t.Fatalf("expected ~50ms, got %v", elapsed)
	}
}
