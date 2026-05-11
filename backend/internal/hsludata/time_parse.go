package hsludata

import (
	"fmt"
	"time"
)

// layouts for datetime_utc like 2023-05-30T13:30:02.403000000 (no zone).
var datetimeLayouts = []string{
	"2006-01-02T15:04:05.999999999",
	"2006-01-02T15:04:05.999999",
	"2006-01-02T15:04:05.999",
	time.RFC3339Nano,
	time.RFC3339,
}

// ParseDateTimeUTC parses the HSLU dataset timestamp column (datetime_utc).
func ParseDateTimeUTC(s string) (time.Time, error) {
	s = trimBOM(s)
	var last error
	for _, layout := range datetimeLayouts {
		t, err := time.ParseInLocation(layout, s, time.UTC)
		if err == nil {
			return t, nil
		}
		last = err
	}
	return time.Time{}, fmt.Errorf("parse time %q: %w", s, last)
}

func trimBOM(s string) string {
	if len(s) >= 3 && s[0] == '\xef' && s[1] == '\xbb' && s[2] == '\xbf' {
		return s[3:]
	}
	return s
}
