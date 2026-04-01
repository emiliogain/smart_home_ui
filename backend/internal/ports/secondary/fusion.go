package secondary

import (
	"context"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
)

// FusionResult holds the output of the sensor-fusion ML model.
type FusionResult struct {
	Label      string            // e.g. "user_away", "sleeping", "cooking"
	Confidence float64           // 0-1
	Actions    map[string]string // suggested UI adaptations
}

// FusionPredictor abstracts the sensor-fusion model.
// The implementation may call a Python service over HTTP/gRPC.
type FusionPredictor interface {
	Predict(ctx context.Context, readings []sensor.Reading) (*FusionResult, error)
}
