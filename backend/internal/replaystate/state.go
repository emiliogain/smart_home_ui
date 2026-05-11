package replaystate

import (
	"time"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	"github.com/emiliogain/smart-home-backend/internal/simulator"
	"github.com/google/uuid"
)

// VirtualState holds the latest scalar for each project sensor name so replay
// can submit full batches (same pattern as the embedded simulator).
type VirtualState struct {
	byName map[string]scalar
}

type scalar struct {
	value float64
	unit  string
}

// ComfortableDefaults matches the "comfortable" scenario baseline.
func ComfortableDefaults() VirtualState {
	s := VirtualState{byName: make(map[string]scalar)}
	for name, prof := range simulator.AllScenarios()[0].Profiles {
		s.byName[name] = scalar{value: prof.Value, unit: prof.Unit}
	}
	return s
}

// SetSensor writes the latest value for a project sensor name (e.g. temp_living_room).
func (v *VirtualState) SetSensor(name string, value float64, unit string) {
	v.byName[name] = scalar{value: value, unit: unit}
}

// ReadingsBatch builds one persisted reading per registered DefaultSensors entry.
func (v *VirtualState) ReadingsBatch(sensorIDs map[string]string, at time.Time) []sensor.Reading {
	var out []sensor.Reading
	for _, def := range simulator.DefaultSensors {
		id := sensorIDs[def.Name]
		if id == "" {
			continue
		}
		sc, ok := v.byName[def.Name]
		if !ok {
			continue
		}
		out = append(out, sensor.Reading{
			ID:        uuid.NewString(),
			SensorID:  id,
			Value:     sc.value,
			Unit:      sc.unit,
			Timestamp: at,
		})
	}
	return out
}
