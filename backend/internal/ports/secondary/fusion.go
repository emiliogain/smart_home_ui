package secondary

import (
	"context"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
)

// SensorWindow groups enriched readings by type and location for fusion.
type SensorWindow struct {
	All        []sensor.EnrichedReading
	ByType     map[sensor.SensorType][]sensor.EnrichedReading
	ByLocation map[string][]sensor.EnrichedReading

	// Now is the reference time used for motion-recency / presence decay.
	// Zero means "use wall clock" (time.Now()), which is the production path:
	// buildSensorWindow leaves this unset so live behavior is unchanged.
	// The offline benchmark sets it to each grid tick's dataset timestamp so
	// recency is measured in dataset time rather than wall-clock time.
	Now time.Time
}

// FusionResult holds the output of the sensor-fusion model.
type FusionResult struct {
	Label      string            // e.g. "NO_ONE_HOME", "SLEEPING", "COOKING_KITCHEN"
	Confidence float64           // 0-1
	Actions    map[string]string // suggested UI adaptations
}

// FusionPredictor abstracts the sensor-fusion model.
type FusionPredictor interface {
	Predict(ctx context.Context, window SensorWindow) (*FusionResult, error)
}
