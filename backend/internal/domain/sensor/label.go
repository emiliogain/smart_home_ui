package sensor

import "strings"

// FormatDisplayLabel returns a short human-readable line for UI (type · room, or name).
func FormatDisplayLabel(name string, typ SensorType, location string) string {
	typePart := titleSegments(string(typ))
	locPart := titleSegments(location)
	switch {
	case typePart != "" && locPart != "":
		return typePart + " · " + locPart
	case name != "":
		return titleSegments(name)
	case typePart != "":
		return typePart
	default:
		return name
	}
}

func titleSegments(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "-", "_"))
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "_")
	for i := range parts {
		p := strings.ToLower(parts[i])
		if p == "" {
			continue
		}
		r := []rune(p)
		if r[0] >= 'a' && r[0] <= 'z' {
			r[0] -= 'a' - 'A'
		}
		parts[i] = string(r)
	}
	return strings.Join(parts, " ")
}
