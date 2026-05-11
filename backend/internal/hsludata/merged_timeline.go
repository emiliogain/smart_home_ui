package hsludata

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/replaystate"
)

// ReplayMergedTimeline reads merged_timeline.csv in time order, applies each row to st,
// and groups rows into batches by dataset time: the first row in a batch sets an anchor;
// subsequent rows are applied in the same batch while row_time <= anchor + batchDelta.
// When a row would exceed that window, onBatch is invoked with the last timestamp of the
// completed batch, then waitAfterBatch wall sleep runs (unless final is true on EOF),
// then the overflow row starts a new batch.
func ReplayMergedTimeline(
	path string,
	userID string,
	batchDelta time.Duration,
	waitAfterBatch time.Duration,
	st *replaystate.VirtualState,
	onBatch func(lastRowTime time.Time) error,
) error {
	if batchDelta <= 0 {
		return fmt.Errorf("batchDelta must be positive")
	}
	if waitAfterBatch < 0 {
		return fmt.Errorf("waitAfterBatch must be non-negative")
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	r := csv.NewReader(f)
	r.LazyQuotes = true
	r.FieldsPerRecord = -1

	hdr, err := r.Read()
	if err != nil {
		return fmt.Errorf("merged_timeline header: %w", err)
	}
	idx := HeaderIndex(hdr)
	for _, col := range []string{"datetime_utc", "stream"} {
		if _, ok := idx[col]; !ok {
			return fmt.Errorf("merged_timeline.csv missing column %q", col)
		}
	}

	var batchStart int64 = -1
	var lastInBatch int64

	flush := func(final bool) error {
		if batchStart < 0 {
			return nil
		}
		if err := onBatch(time.Unix(0, lastInBatch).UTC()); err != nil {
			return err
		}
		if waitAfterBatch > 0 && !final {
			time.Sleep(waitAfterBatch)
		}
		batchStart = -1
		return nil
	}

	for {
		rec, err := r.Read()
		if err == io.EOF {
			return flush(true)
		}
		if err != nil {
			return err
		}

		ti, ok := idx["datetime_utc"]
		if !ok || ti >= len(rec) {
			continue
		}
		at, err := ParseDateTimeUTC(strings.TrimSpace(rec[ti]))
		if err != nil {
			continue
		}
		tNano := at.UnixNano()

		if batchStart < 0 {
			_, _, aerr := ApplyMergedTimelineRow(rec, idx, userID, st)
			if aerr != nil {
				continue
			}
			batchStart = tNano
			lastInBatch = tNano
			continue
		}

		limit := batchStart + batchDelta.Nanoseconds()
		if tNano <= limit {
			_, _, aerr := ApplyMergedTimelineRow(rec, idx, userID, st)
			if aerr != nil {
				continue
			}
			lastInBatch = tNano
			continue
		}

		if err := flush(false); err != nil {
			return err
		}

		_, _, aerr := ApplyMergedTimelineRow(rec, idx, userID, st)
		if aerr != nil {
			continue
		}
		batchStart = tNano
		lastInBatch = tNano
	}
}
