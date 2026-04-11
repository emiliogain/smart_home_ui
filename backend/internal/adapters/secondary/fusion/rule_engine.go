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

// Thresholds holds configurable parameters for the rule-based fusion engine.
type Thresholds struct {
	TempHot          float64       // above this → ALERT_TOO_HOT (default 27)
	TempCold         float64       // below this → ALERT_TOO_COLD (default 17)
	MotionTimeout    time.Duration // no motion for this long → NO_ONE_HOME (default 5m)
	SleepTimeout     time.Duration // no motion for this long → candidate for SLEEPING (default 10m)
	CookingTemp      float64       // kitchen temp above this → cooking signal (default 23)
	CookingHumidity  float64       // kitchen humidity above this → cooking signal (default 50)
	DimLight         float64       // light below this in living room → TV (default 100)
	ReadingLightLow  float64       // reading light lower bound (default 200)
	ReadingLightHigh float64       // reading light upper bound (default 500)
	SleepLight       float64       // light below this → sleeping candidate (default 10)
}

// DefaultThresholds returns production thresholds calibrated to match the
// frontend's mockContextProvider values.
func DefaultThresholds() Thresholds {
	return Thresholds{
		TempHot:          27,
		TempCold:         17,
		MotionTimeout:    5 * time.Minute,
		SleepTimeout:     10 * time.Minute,
		CookingTemp:      23,
		CookingHumidity:  50,
		DimLight:         100,
		ReadingLightLow:  200,
		ReadingLightHigh: 500,
		SleepLight:       10,
	}
}

// RuleBasedPredictor implements FusionPredictor using priority-ordered rules.
type RuleBasedPredictor struct {
	th Thresholds
}

func NewRuleBasedPredictor(th Thresholds) *RuleBasedPredictor {
	return &RuleBasedPredictor{th: th}
}

func (p *RuleBasedPredictor) Predict(_ context.Context, window secondary.SensorWindow) (*secondary.FusionResult, error) {
	snap := extractSnapshot(window)

	log.Printf("[fusion] snapshot: temp=%.1f hum=%.1f lights=%s motion=%s",
		snap.temp, snap.humidity, snap.formatLights(), snap.formatMotion())

	// Priority 1: Temperature alerts (safety first).
	if snap.hasTemp {
		if snap.temp > p.th.TempHot {
			excess := snap.temp - p.th.TempHot
			conf := clamp(0.7+excess*0.05, 0.7, 1.0)
			return p.decide("ALERT_TOO_HOT", conf, "temp %.1f > %.1f", snap.temp, p.th.TempHot), nil
		}
		if snap.temp < p.th.TempCold {
			deficit := p.th.TempCold - snap.temp
			conf := clamp(0.7+deficit*0.05, 0.7, 1.0)
			return p.decide("ALERT_TOO_COLD", conf, "temp %.1f < %.1f", snap.temp, p.th.TempCold), nil
		}
	}

	// Priority 2: No one home.
	if !snap.anyMotionRecent(p.th.MotionTimeout) {
		conf := 0.85
		// Use the brightest light across all rooms as the check.
		if snap.maxLight() < 10 {
			conf = 0.95
		}
		return p.decide("NO_ONE_HOME", conf, "no motion in any room for >%s", p.th.MotionTimeout), nil
	}

	// Priority 3: Cooking in kitchen.
	if snap.motionInLocation("kitchen", p.th.MotionTimeout) {
		cooking := false
		conf := 0.6
		if snap.hasTemp && snap.temp > p.th.CookingTemp {
			cooking = true
			conf += 0.15
		}
		if snap.hasHumidity && snap.humidity > p.th.CookingHumidity {
			cooking = true
			conf += 0.15
		}
		if cooking {
			return p.decide("COOKING_KITCHEN", clamp(conf, 0.6, 0.95),
				"kitchen motion + temp=%.1f hum=%.1f", snap.temp, snap.humidity), nil
		}
	}

	livingLight := snap.lightInLocation("living_room")

	// Priority 4: Watching TV (dim light in living room + motion).
	if snap.motionInLocation("living_room", p.th.MotionTimeout) && livingLight >= 0 && livingLight < p.th.DimLight {
		conf := clamp(0.75+(p.th.DimLight-livingLight)/p.th.DimLight*0.2, 0.7, 0.95)
		return p.decide("WATCHING_TV_LIVING_ROOM", conf,
			"living_room motion + light=%.0f < %.0f", livingLight, p.th.DimLight), nil
	}

	// Priority 5: Reading (moderate light in living room + motion).
	if snap.motionInLocation("living_room", p.th.MotionTimeout) && livingLight >= p.th.ReadingLightLow && livingLight <= p.th.ReadingLightHigh {
		return p.decide("READING_LIVING_ROOM", 0.75,
			"living_room motion + light=%.0f in [%.0f,%.0f]", livingLight, p.th.ReadingLightLow, p.th.ReadingLightHigh), nil
	}

	// Priority 6: Sleeping.
	if !snap.anyMotionRecent(p.th.SleepTimeout) && snap.maxLight() < p.th.SleepLight {
		if snap.lastMotionLocation == "bedroom" || snap.lastMotionLocation == "" {
			conf := 0.8
			if snap.hasTemp && snap.temp >= 17 && snap.temp <= 21 {
				conf = 0.9
			}
			return p.decide("SLEEPING", conf,
				"no motion >%s + light<%.0f + last_motion=%s", p.th.SleepTimeout, p.th.SleepLight, snap.lastMotionLocation), nil
		}
	}

	// Priority 7: Comfortable (someone present, no specific activity).
	if snap.anyMotionRecent(p.th.MotionTimeout) {
		conf := 0.7
		if snap.hasTemp && snap.temp >= 20 && snap.temp <= 25 {
			conf = 0.85
		}
		return p.decide("COMFORTABLE", conf, "motion present, temp=%.1f", snap.temp), nil
	}

	return p.decide("UNKNOWN", 0.5, "no rules matched"), nil
}

func (p *RuleBasedPredictor) decide(label string, confidence float64, reason string, args ...interface{}) *secondary.FusionResult {
	msg := fmt.Sprintf(reason, args...)
	log.Printf("[fusion] → %s (%.0f%%) reason: %s", label, confidence*100, msg)
	return &secondary.FusionResult{
		Label:      label,
		Confidence: confidence,
	}
}

// snapshot aggregates the latest readings for rule evaluation.
type snapshot struct {
	temp        float64
	hasTemp     bool
	humidity    float64
	hasHumidity bool

	// Light per location. Key is location, value is the most recent lux reading.
	lightByLocation map[string]float64

	// Motion tracks the most recent motion-detected time per location.
	motionByLocation    map[string]time.Time
	lastMotionLocation  string
	lastMotionTimestamp time.Time
}

func extractSnapshot(w secondary.SensorWindow) snapshot {
	s := snapshot{
		lightByLocation:  make(map[string]float64),
		motionByLocation: make(map[string]time.Time),
	}

	// Take the most recent reading for each scalar sensor type.
	if temps := w.ByType[sensor.TypeTemperature]; len(temps) > 0 {
		s.temp = temps[0].Value
		s.hasTemp = true
	}
	if hums := w.ByType[sensor.TypeHumidity]; len(hums) > 0 {
		s.humidity = hums[0].Value
		s.hasHumidity = true
	}

	// Track light per location (most recent reading per location).
	for _, r := range w.ByType[sensor.TypeLight] {
		loc := r.Location
		if loc == "" {
			loc = "unknown"
		}
		// ByType is ordered by timestamp DESC, so first per location wins.
		if _, exists := s.lightByLocation[loc]; !exists {
			s.lightByLocation[loc] = r.Value
		}
	}

	// Track motion per location.
	for _, r := range w.ByType[sensor.TypeMotion] {
		if r.Value > 0 {
			loc := r.Location
			if loc == "" {
				loc = "unknown"
			}
			if existing, ok := s.motionByLocation[loc]; !ok || r.Timestamp.After(existing) {
				s.motionByLocation[loc] = r.Timestamp
			}
			if r.Timestamp.After(s.lastMotionTimestamp) {
				s.lastMotionTimestamp = r.Timestamp
				s.lastMotionLocation = loc
			}
		}
	}

	return s
}

// lightInLocation returns the light level for a specific room, or -1 if unknown.
func (s snapshot) lightInLocation(loc string) float64 {
	v, ok := s.lightByLocation[loc]
	if !ok {
		return -1
	}
	return v
}

// maxLight returns the highest light reading across all locations.
func (s snapshot) maxLight() float64 {
	max := -1.0
	for _, v := range s.lightByLocation {
		if v > max {
			max = v
		}
	}
	return max
}

func (s snapshot) anyMotionRecent(timeout time.Duration) bool {
	now := time.Now()
	for _, t := range s.motionByLocation {
		if now.Sub(t) < timeout {
			return true
		}
	}
	return false
}

func (s snapshot) motionInLocation(loc string, timeout time.Duration) bool {
	t, ok := s.motionByLocation[loc]
	if !ok {
		return false
	}
	return time.Since(t) < timeout
}

func (s snapshot) formatLights() string {
	parts := make([]string, 0, len(s.lightByLocation))
	for loc, v := range s.lightByLocation {
		parts = append(parts, fmt.Sprintf("%s=%.0f", loc, v))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, " ")
}

func (s snapshot) formatMotion() string {
	if len(s.motionByLocation) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(s.motionByLocation))
	now := time.Now()
	for loc, t := range s.motionByLocation {
		parts = append(parts, fmt.Sprintf("%s=%s_ago", loc, now.Sub(t).Round(time.Second)))
	}
	return strings.Join(parts, " ")
}

func clamp(v, lo, hi float64) float64 {
	return math.Min(hi, math.Max(lo, v))
}
