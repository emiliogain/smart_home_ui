package hsludata

import "strings"

// NormalizeDatasetRoom maps CSV room labels to fusion locations used in sensor names.
// bathroom is ignored for replay (no matching devices in the app).
func NormalizeDatasetRoom(room string) (loc string, skip bool) {
	switch strings.ToLower(strings.TrimSpace(room)) {
	case "livingroom":
		return "living_room", false
	case "kitchen":
		return "kitchen", false
	case "bedroom":
		return "bedroom", false
	case "general":
		// Common area / unspecified — treat as living room for comfort metrics.
		return "living_room", false
	case "bathroom":
		return "", true
	default:
		return "", true
	}
}
