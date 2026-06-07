package main

import (
	"testing"
	"time"
)

// TestBenchmarkUserFixture runs the offline benchmark against the committed
// synthetic fixture (testdata/bench/user_99) and asserts the harness produces
// sensible label-free metrics and reproduces the key rule-vs-fuzzy divergence:
// the rule engine's priority-2 NO_ONE_HOME short-circuits before its priority-6
// SLEEPING, so it never labels sleep, whereas the fuzzy engine's sleep-temp band
// does. Regenerate the fixture with: python3 testdata/bench/gen_fixture.py
func TestBenchmarkUserFixture(t *testing.T) {
	cfg := benchConfig{grid: time.Minute, windowDur: 0, smootherN: 2}
	res, err := benchmarkUser("testdata/bench/user_99", "99", cfg)
	if err != nil {
		t.Fatalf("benchmarkUser: %v", err)
	}
	if res.Decisions < 100 {
		t.Fatalf("expected >100 decisions, got %d", res.Decisions)
	}
	if res.AgreementPct < 0 || res.AgreementPct > 100 {
		t.Fatalf("agreement out of range: %.2f", res.AgreementPct)
	}
	if res.Rule.LatencyMeanUs <= 0 || res.Fuzzy.LatencyMeanUs <= 0 {
		t.Fatalf("latency means must be positive: rule=%.2f fuzzy=%.2f", res.Rule.LatencyMeanUs, res.Fuzzy.LatencyMeanUs)
	}
	// Sleep/away divergence: fuzzy labels SLEEPING, rule does not.
	if res.Fuzzy.ContextCounts["SLEEPING"] == 0 {
		t.Errorf("expected fuzzy to emit SLEEPING, got distribution %v", res.Fuzzy.ContextCounts)
	}
	if res.Rule.ContextCounts["SLEEPING"] != 0 {
		t.Errorf("expected rule to never emit SLEEPING (NO_ONE_HOME short-circuits), got %d", res.Rule.ContextCounts["SLEEPING"])
	}
	// Both models should agree that the hot phase is an alert.
	if res.Rule.ContextCounts["ALERT_TOO_HOT"] == 0 || res.Fuzzy.ContextCounts["ALERT_TOO_HOT"] == 0 {
		t.Errorf("expected both models to emit ALERT_TOO_HOT")
	}
	// Smoothing must not increase flicker.
	if res.Rule.SmoothedTransitions > res.Rule.Transitions {
		t.Errorf("smoothing increased rule transitions: %d > %d", res.Rule.SmoothedTransitions, res.Rule.Transitions)
	}
}

func TestSmoothStream(t *testing.T) {
	// A single-tick spike (B) between A runs is suppressed at N=2.
	raw := []string{"A", "A", "B", "A", "A"}
	got := smoothStream(raw, 2)
	if countTransitions(got) != 0 {
		t.Errorf("expected spike suppressed (0 transitions), got %v", got)
	}
	// A genuine 2-tick change is accepted.
	raw = []string{"A", "A", "B", "B", "B"}
	got = smoothStream(raw, 2)
	if countTransitions(got) != 1 {
		t.Errorf("expected one accepted transition, got %v", got)
	}
}
