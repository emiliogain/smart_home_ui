package context

import "time"

// ContextType identifies the inferred user/environment context.
// Values match the frontend ContextType enum exactly.
type ContextType string

const (
	NoOneHome         ContextType = "NO_ONE_HOME"
	ReadingLivingRoom ContextType = "READING_LIVING_ROOM"
	WatchingTVLiving  ContextType = "WATCHING_TV_LIVING_ROOM"
	CookingKitchen    ContextType = "COOKING_KITCHEN"
	Sleeping          ContextType = "SLEEPING"
	AlertTooHot       ContextType = "ALERT_TOO_HOT"
	AlertTooCold      ContextType = "ALERT_TOO_COLD"
	Comfortable       ContextType = "COMFORTABLE"
	Unknown           ContextType = "UNKNOWN"
)

// SensorReading is a snapshot of a single sensor value for the frontend.
type SensorReading struct {
	SensorID string  `json:"sensorId"`
	Value    float64 `json:"value"`
	Unit     string  `json:"unit,omitempty"`
	At       string  `json:"at"`
}

// SensorSnapshot groups the latest readings across all sensors.
type SensorSnapshot struct {
	Readings []SensorReading `json:"readings"`
}

// ContextUpdate is the payload sent to the frontend via WebSocket or REST.
type ContextUpdate struct {
	CurrentContext ContextType     `json:"currentContext"`
	Confidence     float64         `json:"confidence"`
	LastUpdated    string          `json:"lastUpdated"`
	SensorSnapshot *SensorSnapshot `json:"sensorSnapshot"`
}

// NewContextUpdate creates a ContextUpdate with the current timestamp.
func NewContextUpdate(ctx ContextType, confidence float64, snapshot *SensorSnapshot) ContextUpdate {
	return ContextUpdate{
		CurrentContext: ctx,
		Confidence:     confidence,
		LastUpdated:    time.Now().UTC().Format(time.RFC3339),
		SensorSnapshot: snapshot,
	}
}
