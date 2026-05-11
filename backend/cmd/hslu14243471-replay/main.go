// Command hslu14243471-replay merges HSLU 14243471-style exports:
//   - event_data.csv (movement, door)
//   - periodic_data_monthly_csv/periodic_data_*.csv (temperature, humidity, ambient_light, …)
//
// into the smart-home API using the same batch endpoint as the simulators.
//
// Preprocess raw exports with scripts/preprocess_hslu_14243471.py, then point
// -data-dir at datasets/hslu_processed/user_<id> (contains events.csv + periodic_data_merged.csv).
//
// Use the same participant id in both streams (see the "id" column). Run the backend
// with simulator_enabled: false while replaying.
//
// Pacing: default -playback 1 replays in real time vs dataset timestamps (5 min between
// rows → 5 min wall). Use -every 10 for exactly one batch every 10 wall seconds.
package main

import (
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
	"github.com/emiliogain/smart-home-backend/internal/replay"
	"github.com/emiliogain/smart-home-backend/internal/simulator"
	"github.com/emiliogain/smart-home-backend/pkg/client"
)

func main() {
	backendURL := flag.String("backend", "http://localhost:8080", "Backend base URL")
	eventPath := flag.String("events", "", "Path to event_data.csv (empty = ~/Downloads/event_data.csv)")
	periodicDir := flag.String("periodic-dir", "", "Directory with periodic_data_*.csv (empty = ~/Downloads/periodic_data_monthly_csv)")
	dataDir := flag.String("data-dir", "", "Preprocessed folder: .../user_<id> with events.csv and periodic_data_*.csv (overrides -events/-periodic-dir)")
	userID := flag.String("user-id", "", "Participant id as in the CSV id column (required)")
	playback := flag.Float64("playback", 1, "Dataset-time vs wall-time: sleep=datasetDelta/playback when -every=0 (1=real-time, 2=2× faster). 0=no delay unless -every>0")
	every := flag.Float64("every", 0, "If >0, wait this many wall seconds after each batch (fixed pace, e.g. 10 = one update every 10s). 0=use -playback with dataset timestamps")
	maxWait := flag.Float64("max-wait", 0, "Optional cap in seconds on sleep from -playback only (0=no cap; use for very long dataset gaps)")
	maxBatches := flag.Int("max-batches", 0, "Stop after this many successful batch writes (0 = unlimited)")
	useDefaults := flag.Bool("defaults", true, "When -events/-periodic-dir are empty, use ~/Downloads paths (ignored if -data-dir is set)")
	flag.Parse()

	if strings.TrimSpace(*userID) == "" {
		log.Fatal("flag -user-id is required (must match rows in both event and periodic files)")
	}

	evPath := strings.TrimSpace(*eventPath)
	perDir := strings.TrimSpace(*periodicDir)
	dataDirExpanded := strings.TrimSpace(*dataDir)
	if dataDirExpanded != "" {
		var err error
		dataDirExpanded, err = hsludata.ExpandUser(dataDirExpanded)
		if err != nil {
			log.Fatal(err)
		}
		st, err := os.Stat(dataDirExpanded)
		if err != nil || !st.IsDir() {
			log.Fatalf("-data-dir must be an existing directory: %s", dataDirExpanded)
		}
		evPath = filepath.Join(dataDirExpanded, "events.csv")
		perDir = dataDirExpanded
		if _, err := os.Stat(evPath); err != nil {
			log.Fatalf("preprocessed events.csv missing under -data-dir: %v", err)
		}
		log.Printf("using -data-dir layout: events=%s periodic_dir=%s", evPath, perDir)
	} else {
		if evPath == "" && *useDefaults {
			var err error
			evPath, perDir, err = hsludata.DefaultDownloadsPaths()
			if err != nil {
				log.Fatalf("defaults: %v (pass -events and -periodic-dir, or -data-dir)", err)
			}
			log.Printf("using defaults: events=%s periodic=%s", evPath, perDir)
		}
		if evPath == "" || perDir == "" {
			log.Fatal("need -data-dir, or -events and -periodic-dir, or -defaults with files in ~/Downloads/")
		}
		var err error
		evPath, err = hsludata.ExpandUser(evPath)
		if err != nil {
			log.Fatal(err)
		}
		perDir, err = hsludata.ExpandUser(perDir)
		if err != nil {
			log.Fatal(err)
		}
	}

	c := client.New(strings.TrimRight(*backendURL, "/"))
	sensorIDs, err := ensureDefaultSensors(c)
	if err != nil {
		log.Fatalf("sensors: %v", err)
	}

	state := replaystate.ComfortableDefaults()
	var prevDataset int64
	first := true
	batches := 0

	err = hsludata.ForEachMergedRow(evPath, perDir, strings.TrimSpace(*userID), func(src *hsludata.Stream, tNano int64) error {
		var changed bool
		var applyErr error
		switch src.Kind {
		case "event":
			_, changed, applyErr = hsludata.ApplyEventRow(src.Cur, src.Idx, "", &state)
		case "periodic":
			_, changed, applyErr = hsludata.ApplyPeriodicRow(src.Cur, src.Idx, "", &state)
		}
		if applyErr != nil {
			log.Printf("skip %s row: %v", src.Path, applyErr)
			return nil
		}
		if !changed {
			return nil
		}

		at := time.Unix(0, tNano).UTC()
		if first {
			prevDataset = tNano
			first = false
		} else {
			delta := time.Duration(tNano - prevDataset)
			replay.SleepBetweenRows(delta, *every, *playback, *maxWait)
		}
		prevDataset = tNano

		// Stamp readings at ingest time so fusion rules (which use time.Now()) see motion
		// as recent. Dataset timestamps are only used for pacing above, not for DB time.
		ingestAt := time.Now().UTC()
		batch := state.ReadingsBatch(sensorIDs, ingestAt)
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
			log.Printf("hslu replay: batch #%d source=%s kind=%s t=%s", batches, src.Path, src.Kind, at.Format(time.RFC3339Nano))
		}
		if *maxBatches > 0 && batches >= *maxBatches {
			return errStopReplay
		}
		return nil
	})
	if errors.Is(err, errStopReplay) {
		log.Printf("stopped after %d batches (-max-batches)", batches)
	} else if err != nil {
		log.Fatalf("replay: %v", err)
	}
	log.Printf("hslu14243471-replay finished (%d batches)", batches)
}

var errStopReplay = errors.New("stop replay")

func ensureDefaultSensors(c *client.Client) (map[string]string, error) {
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
	for _, def := range simulator.DefaultSensors {
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
	if len(out) != len(simulator.DefaultSensors) {
		return nil, fmt.Errorf("expected %d sensors, got %d", len(simulator.DefaultSensors), len(out))
	}
	return out, nil
}
