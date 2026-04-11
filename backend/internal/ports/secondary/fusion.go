package secondary

import (
	"context"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
)

// SensorWindow groups enriched readings by type and location for fusion.
type SensorWindow struct {
	All        []sensor.EnrichedReading
	ByType     map[sensor.SensorType][]sensor.EnrichedReading
	ByLocation map[string][]sensor.EnrichedReading
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
