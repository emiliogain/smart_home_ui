package fusion

import (
	"context"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	"github.com/emiliogain/smart-home-backend/internal/ports/secondary"
)

// StubPredictor is a placeholder that always returns "normal".
// Replace with an HTTP/gRPC client that calls the real Python fusion model.
type StubPredictor struct{}

func NewStubPredictor() *StubPredictor {
	return &StubPredictor{}
}

func (p *StubPredictor) Predict(_ context.Context, _ []sensor.Reading) (*secondary.FusionResult, error) {
	return &secondary.FusionResult{
		Label:      "normal",
		Confidence: 1.0,
		Actions:    nil,
	}, nil
}
