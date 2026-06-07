package fusion

import (
	"math"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
)

// latestValueAtLocation returns the first reading (newest-first slice) for a location.
func latestValueAtLocation(readings []sensor.EnrichedReading, loc string) (float64, bool) {
	for _, r := range readings {
		if r.Location == loc {
			return r.Value, true
		}
	}
	return 0, false
}

func fuseMax(a float64, hasA bool, b float64, hasB bool) (float64, bool) {
	switch {
	case hasA && hasB:
		return math.Max(a, b), true
	case hasA:
		return a, true
	case hasB:
		return b, true
	default:
		return 0, false
	}
}
