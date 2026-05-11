package hsludata

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/emiliogain/smart-home-backend/internal/replaystate"
)

// ApplyMergedTimelineRow applies one row from preprocess merged_timeline.csv
// (stream=event|periodic). Column layout matches events.csv / periodic_data_merged.csv fields.
func ApplyMergedTimelineRow(rec []string, idx map[string]int, userID string, st *replaystate.VirtualState) (tNano int64, changed bool, err error) {
	si, ok := idx["stream"]
	if !ok {
		return 0, false, fmt.Errorf("missing stream column")
	}
	stream := strings.ToLower(strings.TrimSpace(rec[si]))
	switch stream {
	case "event":
		return ApplyEventRow(rec, idx, userID, st)
	case "periodic":
		return ApplyPeriodicRow(rec, idx, userID, st)
	default:
		return 0, false, fmt.Errorf("unknown stream %q", stream)
	}
}

// ApplyEventRow updates motion from event_data.csv (movement on/off).
// Columns: datetime_utc,id,country,room,sensor,value
func ApplyEventRow(rec []string, idx map[string]int, userID string, st *replaystate.VirtualState) (tUnix int64, changed bool, err error) {
	if userID != "" {
		id := strings.TrimSpace(rec[idx["id"]])
		if id != userID {
			return 0, false, nil
		}
	}
	at, err := ParseDateTimeUTC(rec[idx["datetime_utc"]])
	if err != nil {
		return 0, false, err
	}
	room := rec[idx["room"]]
	loc, skip := NormalizeDatasetRoom(room)
	if skip {
		return at.UnixNano(), false, nil
	}
	sensor := strings.ToLower(strings.TrimSpace(rec[idx["sensor"]]))
	val := strings.ToLower(strings.TrimSpace(rec[idx["value"]]))
	switch sensor {
	case "movement":
		if val == "on" {
			st.SetSensor("motion_"+loc, 1, "")
			return at.UnixNano(), true, nil
		}
		if val == "off" {
			st.SetSensor("motion_"+loc, 0, "")
			return at.UnixNano(), true, nil
		}
		return at.UnixNano(), false, nil
	case "door":
		// Door events are ignored for fusion (no door actuator in the model).
		return at.UnixNano(), false, nil
	default:
		return at.UnixNano(), false, nil
	}
}

// ApplyPeriodicRow updates scalars from periodic_data_*.csv.
// Columns: datetime_utc,id,country,room,sensor,min_value,average_value,max_value
func ApplyPeriodicRow(rec []string, idx map[string]int, userID string, st *replaystate.VirtualState) (tUnix int64, changed bool, err error) {
	if userID != "" {
		id := strings.TrimSpace(rec[idx["id"]])
		if id != userID {
			return 0, false, nil
		}
	}
	at, err := ParseDateTimeUTC(rec[idx["datetime_utc"]])
	if err != nil {
		return 0, false, err
	}
	room := rec[idx["room"]]
	loc, skip := NormalizeDatasetRoom(room)
	if skip {
		return at.UnixNano(), false, nil
	}
	sensor := strings.ToLower(strings.TrimSpace(rec[idx["sensor"]]))
	avgStr := strings.TrimSpace(rec[idx["average_value"]])
	if avgStr == "" {
		return at.UnixNano(), false, fmt.Errorf("empty average_value")
	}
	v, err := strconv.ParseFloat(strings.ReplaceAll(avgStr, ",", "."), 64)
	if err != nil {
		return 0, false, fmt.Errorf("average_value: %w", err)
	}

	switch sensor {
	case "temperature":
		st.SetSensor("temp_"+loc, v, "°C")
		return at.UnixNano(), true, nil
	case "humidity":
		st.SetSensor("humidity_"+loc, v, "%")
		return at.UnixNano(), true, nil
	case "ambient_light":
		st.SetSensor("light_"+loc, v, "lux")
		return at.UnixNano(), true, nil
	case "co2", "voc", "sound_db_average", "sound_db_max":
		return at.UnixNano(), false, nil
	default:
		return at.UnixNano(), false, nil
	}
}
