package fusion

import (
	"context"

	"github.com/emiliogain/smart-home-backend/internal/ports/secondary"
)

// StubPredictor is a placeholder that always returns "COMFORTABLE".
// Used for testing when the rule engine is not needed.
type StubPredictor struct{}

func NewStubPredictor() *StubPredictor {
	return &StubPredictor{}
}

func (p *StubPredictor) Predict(_ context.Context, _ secondary.SensorWindow) (*secondary.FusionResult, error) {
	return &secondary.FusionResult{
		Label:      "COMFORTABLE",
		Confidence: 1.0,
		Actions:    nil,
	}, nil
}
