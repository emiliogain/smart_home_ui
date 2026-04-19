package fusion

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	"github.com/emiliogain/smart-home-backend/internal/ports/secondary"
)

// FuzzyPredictor implements FusionPredictor using fuzzy logic.
//
// Unlike the rule-based predictor (which uses hard Boolean thresholds),
// fuzzy logic assigns a membership degree [0,1] to each linguistic term
// (e.g. temperature is "hot" with degree 0.6). Rules combine these degrees
// using AND (min) and OR (max) operators to produce a fire strength for
// each context. The context with the highest fire strength wins.
type FuzzyPredictor struct{}

func NewFuzzyPredictor() *FuzzyPredictor { return &FuzzyPredictor{} }

func (p *FuzzyPredictor) Predict(_ context.Context, window secondary.SensorWindow) (*secondary.FusionResult, error) {
	fs := extractFuzzySnapshot(window)

	log.Printf("[fusion/fuzzy] temp=%.1f hum=%.1f lights=%s motion=%s",
		fs.temp, fs.humidity, fs.formatLights(), fs.formatMotion())

	// ── Membership values ──────────────────────────────────────────────────
	fHot := rampUp(fs.temp, 24, 27)
	fCold := rampDown(fs.temp, 16, 19)
	fComfTemp := trapMF(fs.temp, 16, 19, 24, 27)
	// Sleep temperature: deliberately cooled for sleep (17–19°C flat top).
	// Distinct from NO_ONE_HOME which has normal ambient temp (≥22°C → outside this range).
	fSleepTemp := trapMF(fs.temp, 15, 17, 19, 21)
	fNormHum := trapMF(fs.humidity, 30, 40, 55, 65)
	fHumid := rampUp(fs.humidity, 50, 65)

	lightLiving := fs.lightIn("living_room")
	fDimLiving := trapMF(lightLiving, 5, 30, 80, 120)
	fModLiving := trapMF(lightLiving, 80, 200, 350, 450)

	// Kitchen brightness is a strong cooking signal (bright overhead light when cooking).
	lightKitchen := fs.lightIn("kitchen")
	fBrightKitchen := rampUp(lightKitchen, 350, 450)

	// Motion recency per room (seconds since last motion event).
	fPresLiving := fs.presenceIn("living_room")
	fAbsLiving := 1 - fPresLiving
	fPresKitchen := fs.presenceIn("kitchen")
	fAbsKitchen := 1 - fPresKitchen
	fPresBedroom := fs.presenceIn("bedroom")
	fAbsBedroom := 1 - fPresBedroom

	maxLight := fs.maxLight()
	fDarkAll := rampDown(maxLight, 5, 20)
	fPresAny := math.Max(fPresLiving, math.Max(fPresKitchen, fPresBedroom))
	fAbsAll := math.Min(fAbsLiving, math.Min(fAbsKitchen, fAbsBedroom))

	// ── Rules (priority order: first highest-strength rule wins on ties) ──
	type rule struct {
		context string
		fire    float64
	}

	rules := []rule{
		// R1: temperature alert — hot (safety, highest priority)
		{"ALERT_TOO_HOT", fHot},
		// R2: temperature alert — cold
		{"ALERT_TOO_COLD", fCold},
		// R3: sleeping — MUST come before NO_ONE_HOME because sleeping is a
		// special case of absence (no motion + dark + sleep temperature).
		// fSleepTemp distinguishes it from NO_ONE_HOME (normal ambient temp).
		{"SLEEPING", fmin(fAbsAll, fmin(fDarkAll, fSleepTemp))},
		// R4: no one home — absent everywhere (catches non-sleeping absence)
		{"NO_ONE_HOME", fAbsAll},
		// R5: cooking — kitchen presence AND (humid OR hot OR bright kitchen light)
		{"COOKING_KITCHEN", fmin(fPresKitchen, fmax(fHumid, fmax(fHot, fBrightKitchen)))},
		// R6: watching TV — living room presence AND dim light
		{"WATCHING_TV_LIVING_ROOM", fmin(fPresLiving, fDimLiving)},
		// R7: reading — living room presence AND moderate light
		// comfortable scenario uses 500 lux (bright, not moderate) to avoid this rule
		{"READING_LIVING_ROOM", fmin(fPresLiving, fModLiving)},
		// R8: comfortable — someone present AND comfortable temp AND normal humidity
		// (fallback when no specific activity is detected)
		{"COMFORTABLE", fmin(fPresAny, fmin(fComfTemp, fNormHum))},
	}

	// Pick the rule with the highest fire strength.
	best := rule{context: "UNKNOWN", fire: 0}
	for _, r := range rules {
		if r.fire > best.fire {
			best = r
		}
	}

	if best.fire < 0.1 {
		log.Printf("[fusion/fuzzy] → UNKNOWN (fire strength too low: %.3f)", best.fire)
		return &secondary.FusionResult{Label: "UNKNOWN", Confidence: 0.5}, nil
	}

	// Log all rules with non-trivial fire strength for interpretability.
	var fired []string
	for _, r := range rules {
		if r.fire > 0.05 {
			fired = append(fired, fmt.Sprintf("%s=%.2f", r.context, r.fire))
		}
	}
	log.Printf("[fusion/fuzzy] → %s (%.0f%%) fired: %s",
		best.context, best.fire*100, strings.Join(fired, " "))

	return &secondary.FusionResult{Label: best.context, Confidence: best.fire}, nil
}

// ── Fuzzy snapshot ───────────────────────────────────────────────────────────

type fuzzySnapshot struct {
	temp        float64
	humidity    float64
	lightByLoc  map[string]float64
	motionByLoc map[string]time.Time // location → last motion timestamp
}

func extractFuzzySnapshot(w secondary.SensorWindow) fuzzySnapshot {
	s := fuzzySnapshot{
		lightByLoc:  make(map[string]float64),
		motionByLoc: make(map[string]time.Time),
	}
	if temps := w.ByType[sensor.TypeTemperature]; len(temps) > 0 {
		s.temp = temps[0].Value
	}
	if hums := w.ByType[sensor.TypeHumidity]; len(hums) > 0 {
		s.humidity = hums[0].Value
	}
	for _, r := range w.ByType[sensor.TypeLight] {
		loc := r.Location
		if loc == "" {
			loc = "unknown"
		}
		if _, exists := s.lightByLoc[loc]; !exists {
			s.lightByLoc[loc] = r.Value
		}
	}
	for _, r := range w.ByType[sensor.TypeMotion] {
		if r.Value > 0 {
			loc := r.Location
			if loc == "" {
				loc = "unknown"
			}
			if t, ok := s.motionByLoc[loc]; !ok || r.Timestamp.After(t) {
				s.motionByLoc[loc] = r.Timestamp
			}
		}
	}
	return s
}

// presenceIn returns the fuzzy presence degree for a room [0,1] based on
// how recently motion was detected. Fully present within 1 min, zero after 5 min.
func (s fuzzySnapshot) presenceIn(loc string) float64 {
	t, ok := s.motionByLoc[loc]
	if !ok {
		return 0
	}
	age := time.Since(t).Seconds()
	return rampDown(age, 60, 300)
}

func (s fuzzySnapshot) lightIn(loc string) float64 {
	v, ok := s.lightByLoc[loc]
	if !ok {
		return -1
	}
	return v
}

func (s fuzzySnapshot) maxLight() float64 {
	max := 0.0
	for _, v := range s.lightByLoc {
		if v > max {
			max = v
		}
	}
	return max
}

func (s fuzzySnapshot) formatLights() string {
	parts := make([]string, 0)
	for loc, v := range s.lightByLoc {
		parts = append(parts, fmt.Sprintf("%s=%.0f", loc, v))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, " ")
}

func (s fuzzySnapshot) formatMotion() string {
	if len(s.motionByLoc) == 0 {
		return "none"
	}
	parts := make([]string, 0)
	now := time.Now()
	for loc, t := range s.motionByLoc {
		parts = append(parts, fmt.Sprintf("%s=%s_ago", loc, now.Sub(t).Round(time.Second)))
	}
	return strings.Join(parts, " ")
}

// ── Membership functions ─────────────────────────────────────────────────────

// trapMF returns the trapezoid membership: 0 below a, rises a→b, flat b→c, falls c→d, 0 above d.
func trapMF(x, a, b, c, d float64) float64 {
	if x <= a || x >= d {
		return 0
	}
	if x >= b && x <= c {
		return 1
	}
	if x < b {
		return (x - a) / (b - a)
	}
	return (d - x) / (d - c)
}

// rampUp returns 0 at x≤a, 1 at x≥b, linear in between (right-shoulder).
func rampUp(x, a, b float64) float64 {
	if x <= a {
		return 0
	}
	if x >= b {
		return 1
	}
	return (x - a) / (b - a)
}

// rampDown returns 1 at x≤a, 0 at x≥b, linear in between (left-shoulder).
func rampDown(x, a, b float64) float64 {
	if x <= a {
		return 1
	}
	if x >= b {
		return 0
	}
	return (b - x) / (b - a)
}

// fuzzy AND (minimum T-norm)
func fmin(a, b float64) float64 { return math.Min(a, b) }

// fuzzy OR (maximum S-norm)
func fmax(a, b float64) float64 { return math.Max(a, b) }
