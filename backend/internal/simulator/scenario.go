package simulator

import (
	"math/rand"
)

// SensorProfile defines the base value and noise for a sensor in a given scenario.
type SensorProfile struct {
	Value float64
	Noise float64 // standard deviation of Gaussian noise
	Unit  string
}

// Scenario defines the expected sensor readings for a particular context.
type Scenario struct {
	Name     string
	Profiles map[string]SensorProfile // keyed by sensor name
}

// Generate produces a noisy reading value from the profile.
func (p SensorProfile) Generate() float64 {
	if p.Noise == 0 {
		return p.Value
	}
	v := p.Value + rand.NormFloat64()*p.Noise
	if v < 0 {
		v = 0
	}
	return v
}

// AllScenarios returns the full set of simulation scenarios, calibrated
// to trigger the corresponding fusion context labels.
func AllScenarios() []Scenario {
	return []Scenario{
		{
			// Comfortable: bright overhead lighting (500 lux) places this above the
			// "moderate/reading" range, so the fuzzy engine selects COMFORTABLE not READING.
			Name: "comfortable",
			Profiles: map[string]SensorProfile{
				"temp_living_room":     {Value: 22, Noise: 0.5, Unit: "°C"},
				"humidity_living_room": {Value: 45, Noise: 2, Unit: "%"},
				"light_living_room":    {Value: 500, Noise: 30, Unit: "lux"},
				"motion_living_room":   {Value: 1, Noise: 0, Unit: ""},
				"light_kitchen":        {Value: 100, Noise: 20, Unit: "lux"},
				"motion_kitchen":       {Value: 0, Noise: 0, Unit: ""},
				"light_bedroom":        {Value: 50, Noise: 10, Unit: "lux"},
				"motion_bedroom":       {Value: 0, Noise: 0, Unit: ""},
			},
		},
		{
			Name: "reading",
			Profiles: map[string]SensorProfile{
				"temp_living_room":     {Value: 21, Noise: 0.5, Unit: "°C"},
				"humidity_living_room": {Value: 42, Noise: 2, Unit: "%"},
				"light_living_room":    {Value: 320, Noise: 30, Unit: "lux"},
				"motion_living_room":   {Value: 1, Noise: 0, Unit: ""},
				"light_kitchen":        {Value: 20, Noise: 10, Unit: "lux"},
				"motion_kitchen":       {Value: 0, Noise: 0, Unit: ""},
				"light_bedroom":        {Value: 10, Noise: 5, Unit: "lux"},
				"motion_bedroom":       {Value: 0, Noise: 0, Unit: ""},
			},
		},
		{
			Name: "watching_tv",
			Profiles: map[string]SensorProfile{
				"temp_living_room":     {Value: 22, Noise: 0.5, Unit: "°C"},
				"humidity_living_room": {Value: 44, Noise: 2, Unit: "%"},
				"light_living_room":    {Value: 50, Noise: 15, Unit: "lux"},
				"motion_living_room":   {Value: 1, Noise: 0, Unit: ""},
				"light_kitchen":        {Value: 10, Noise: 5, Unit: "lux"},
				"motion_kitchen":       {Value: 0, Noise: 0, Unit: ""},
				"light_bedroom":        {Value: 5, Noise: 3, Unit: "lux"},
				"motion_bedroom":       {Value: 0, Noise: 0, Unit: ""},
			},
		},
		{
			Name: "cooking",
			Profiles: map[string]SensorProfile{
				"temp_living_room":     {Value: 24, Noise: 0.5, Unit: "°C"},
				"humidity_living_room": {Value: 58, Noise: 3, Unit: "%"},
				"light_living_room":    {Value: 200, Noise: 30, Unit: "lux"},
				"motion_living_room":   {Value: 0, Noise: 0, Unit: ""},
				"light_kitchen":        {Value: 450, Noise: 30, Unit: "lux"},
				"motion_kitchen":       {Value: 1, Noise: 0, Unit: ""},
				"light_bedroom":        {Value: 10, Noise: 5, Unit: "lux"},
				"motion_bedroom":       {Value: 0, Noise: 0, Unit: ""},
			},
		},
		{
			// Sleeping: cooled to 18°C (inside the fuzzy "sleep temp" range 17–19°C),
			// which distinguishes it from NO_ONE_HOME (temp 22°C, outside sleep range).
			Name: "sleeping",
			Profiles: map[string]SensorProfile{
				"temp_living_room":     {Value: 18, Noise: 0.3, Unit: "°C"},
				"humidity_living_room": {Value: 40, Noise: 2, Unit: "%"},
				"light_living_room":    {Value: 2, Noise: 1, Unit: "lux"},
				"motion_living_room":   {Value: 0, Noise: 0, Unit: ""},
				"light_kitchen":        {Value: 0, Noise: 0, Unit: "lux"},
				"motion_kitchen":       {Value: 0, Noise: 0, Unit: ""},
				"light_bedroom":        {Value: 3, Noise: 1, Unit: "lux"},
				"motion_bedroom":       {Value: 0, Noise: 0, Unit: ""},
			},
		},
		{
			// No-one-home: temp 22°C (outside the fuzzy "sleep temp" range 17–19°C),
			// ensuring this is never confused with SLEEPING even when all lights are off.
			Name: "no_one_home",
			Profiles: map[string]SensorProfile{
				"temp_living_room":     {Value: 22, Noise: 0.3, Unit: "°C"},
				"humidity_living_room": {Value: 45, Noise: 1, Unit: "%"},
				"light_living_room":    {Value: 0, Noise: 0, Unit: "lux"},
				"motion_living_room":   {Value: 0, Noise: 0, Unit: ""},
				"light_kitchen":        {Value: 0, Noise: 0, Unit: "lux"},
				"motion_kitchen":       {Value: 0, Noise: 0, Unit: ""},
				"light_bedroom":        {Value: 0, Noise: 0, Unit: "lux"},
				"motion_bedroom":       {Value: 0, Noise: 0, Unit: ""},
			},
		},
		{
			Name: "alert_too_hot",
			Profiles: map[string]SensorProfile{
				"temp_living_room":     {Value: 29, Noise: 0.5, Unit: "°C"},
				"humidity_living_room": {Value: 55, Noise: 3, Unit: "%"},
				"light_living_room":    {Value: 400, Noise: 30, Unit: "lux"},
				"motion_living_room":   {Value: 1, Noise: 0, Unit: ""},
				"light_kitchen":        {Value: 200, Noise: 20, Unit: "lux"},
				"motion_kitchen":       {Value: 0, Noise: 0, Unit: ""},
				"light_bedroom":        {Value: 100, Noise: 10, Unit: "lux"},
				"motion_bedroom":       {Value: 0, Noise: 0, Unit: ""},
			},
		},
	}
}

// ScenarioNames returns the names of all available scenarios.
func ScenarioNames() []string {
	scenarios := AllScenarios()
	names := make([]string, len(scenarios))
	for i, s := range scenarios {
		names[i] = s.Name
	}
	return names
}
