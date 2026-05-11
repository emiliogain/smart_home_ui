// Command hslu14243471-replay pushes preprocessed HSLU data into the smart-home API using
// the same batch endpoint as the simulators.
//
// Preprocess with scripts/preprocess_hslu_14243471.py, then pass -data-dir to the folder
// user_<id>/ which must contain merged_timeline.csv (and optionally sensors_manifest.json).
//
// Replay reads merged_timeline.csv in order. The first row in each batch sets a dataset-time
// anchor; further rows are applied while timestamp <= anchor + -timeline-delta, then one
// batch POST runs and the process sleeps -timeline-wait before continuing.
//
// Run the backend with simulator_enabled: false while replaying.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/hsludata"
	"github.com/emiliogain/smart-home-backend/internal/replaystate"
	"github.com/emiliogain/smart-home-backend/internal/simulator"
	"github.com/emiliogain/smart-home-backend/pkg/client"
)

func main() {
	backendURL := flag.String("backend", "http://localhost:8080", "Backend base URL")
	dataDir := flag.String("data-dir", "", "Preprocessed folder user_<id>/ containing merged_timeline.csv (required)")
	userID := flag.String("user-id", "", "Participant id as in the CSV id column (required)")
	timelineDeltaStr := flag.String("timeline-delta", "90s", "Dataset-time span from batch start; rows with t <= start+delta stay in the same batch")
	timelineWaitStr := flag.String("timeline-wait", "10s", "Wall sleep after each batch POST before reading the next row")
	maxBatches := flag.Int("max-batches", 0, "Stop after this many successful batch writes (0 = unlimited)")
	flag.Parse()

	if strings.TrimSpace(*userID) == "" {
		log.Fatal("flag -user-id is required (must match the id column in merged_timeline.csv)")
	}
	dataDirExpanded := strings.TrimSpace(*dataDir)
	if dataDirExpanded == "" {
		log.Fatal("flag -data-dir is required (folder with merged_timeline.csv)")
	}
	var err error
	dataDirExpanded, err = hsludata.ExpandUser(dataDirExpanded)
	if err != nil {
		log.Fatal(err)
	}
	st, err := os.Stat(dataDirExpanded)
	if err != nil || !st.IsDir() {
		log.Fatalf("-data-dir must be an existing directory: %s", dataDirExpanded)
	}
	mergedPath := filepath.Join(dataDirExpanded, "merged_timeline.csv")
	if _, err := os.Stat(mergedPath); err != nil {
		log.Fatalf("merged_timeline.csv missing under -data-dir: %s: %v", mergedPath, err)
	}
	log.Printf("replay merged timeline: %s", mergedPath)

	c := client.New(strings.TrimRight(*backendURL, "/"))

	if err := c.ResetDB(); err != nil {
		log.Fatalf("reset db: %v", err)
	}
	log.Printf("db reset: sensors and readings truncated")

	mp := filepath.Join(dataDirExpanded, "sensors_manifest.json")
	replayDefs, err := loadSensorManifest(mp)
	if err != nil {
		log.Fatalf("sensors_manifest.json missing or invalid under -data-dir: %v", err)
	}
	log.Printf("loaded %d sensors from %s", len(replayDefs), mp)

	sensorIDs, err := ensureSensors(c, replayDefs)
	if err != nil {
		log.Fatalf("sensors: %v", err)
	}

	state := replaystate.ReplayBaseline(replayDefs)
	batches := 0

	submit := func(at time.Time) error {
		ingestAt := time.Now().UTC()
		batch := state.ReadingsBatch(sensorIDs, replayDefs, ingestAt)
		if len(batch) == 0 {
			return nil
		}
		var br []client.BatchReading
		for _, rd := range batch {
			br = append(br, client.BatchReading{
				SensorID:  rd.SensorID,
				Value:     rd.Value,
				Unit:      rd.Unit,
				Timestamp: rd.Timestamp,
			})
		}
		if err := c.SubmitReadingsBatch(br); err != nil {
			return fmt.Errorf("batch: %w", err)
		}
		batches++
		if batches == 1 || batches%500 == 0 {
			log.Printf("hslu replay: batch #%d dataset t=%s", batches, at.Format(time.RFC3339Nano))
		}
		if *maxBatches > 0 && batches >= *maxBatches {
			return errStopReplay
		}
		return nil
	}

	td, e1 := time.ParseDuration(strings.TrimSpace(*timelineDeltaStr))
	if e1 != nil || td <= 0 {
		log.Fatalf("-timeline-delta: need a positive duration (e.g. 90s, 2m): %v", e1)
	}
	tw, e2 := time.ParseDuration(strings.TrimSpace(*timelineWaitStr))
	if e2 != nil || tw < 0 {
		log.Fatalf("-timeline-wait: need a non-negative duration: %v", e2)
	}
	log.Printf("batch window=%v wall wait after each batch=%v", td, tw)

	replayErr := hsludata.ReplayMergedTimeline(mergedPath, strings.TrimSpace(*userID), td, tw, &state, submit)
	if errors.Is(replayErr, errStopReplay) {
		log.Printf("stopped after %d batches (-max-batches)", batches)
	} else if replayErr != nil {
		log.Fatalf("replay: %v", replayErr)
	}
	log.Printf("hslu14243471-replay finished (%d batches)", batches)
}

var errStopReplay = errors.New("stop replay")

type sensorsManifestFile struct {
	Sensors []struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Location string `json:"location"`
	} `json:"sensors"`
}

func loadSensorManifest(path string) ([]simulator.SensorDef, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m sensorsManifestFile
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	out := make([]simulator.SensorDef, 0, len(m.Sensors))
	for _, s := range m.Sensors {
		name := strings.TrimSpace(s.Name)
		if name == "" {
			continue
		}
		out = append(out, simulator.SensorDef{
			Name:     name,
			Type:     strings.TrimSpace(s.Type),
			Location: strings.TrimSpace(s.Location),
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty or invalid sensors list")
	}
	return out, nil
}

func ensureSensors(c *client.Client, defs []simulator.SensorDef) (map[string]string, error) {
	seen := make(map[string]struct{})
	var uniq []simulator.SensorDef
	for _, d := range defs {
		name := strings.TrimSpace(d.Name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		uniq = append(uniq, d)
	}
	defs = uniq
	if len(defs) == 0 {
		return nil, fmt.Errorf("no sensors to register")
	}

	list, err := c.ListSensors()
	if err != nil {
		return nil, err
	}
	byName := make(map[string]string)
	for _, s := range list {
		if _, exists := byName[s.Name]; !exists {
			byName[s.Name] = s.ID
		}
	}
	out := make(map[string]string)
	for _, def := range defs {
		if id, ok := byName[def.Name]; ok {
			out[def.Name] = id
			log.Printf("sensor %q already registered (id=%s)", def.Name, id)
			continue
		}
		resp, err := c.RegisterSensor(def.Name, def.Type, def.Location)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", def.Name, err)
		}
		out[def.Name] = resp.ID
		log.Printf("registered sensor %q (id=%s)", def.Name, resp.ID)
	}
	if len(out) != len(defs) {
		return nil, fmt.Errorf("expected %d sensors, got %d", len(defs), len(out))
	}
	return out, nil
}
