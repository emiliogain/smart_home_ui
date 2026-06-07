package hsludata

import "strings"

// HeaderIndex maps lower-cased trimmed header → column index.
func HeaderIndex(header []string) map[string]int {
	idx := make(map[string]int)
	for i, h := range header {
		key := strings.ToLower(strings.TrimSpace(h))
		if key != "" {
			idx[key] = i
		}
	}
	return idx
}
