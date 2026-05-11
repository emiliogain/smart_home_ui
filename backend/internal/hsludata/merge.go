package hsludata

import (
	"container/heap"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// heapItem references an active stream by index.
type heapItem struct {
	tNano int64
	src   int
}

type timeHeap []heapItem

func (h timeHeap) Len() int { return len(h) }
func (h timeHeap) Less(i, j int) bool {
	return h[i].tNano < h[j].tNano || (h[i].tNano == h[j].tNano && h[i].src < h[j].src)
}
func (h timeHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *timeHeap) Push(x interface{}) { *h = append(*h, x.(heapItem)) }

func (h *timeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// OpenPeriodicStreams opens every periodic_data_*.csv under dir (sorted by name).
func OpenPeriodicStreams(dir, userID string) ([]*Stream, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "periodic_data_*.csv"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no periodic_data_*.csv under %s", dir)
	}
	sort.Strings(matches)
	var out []*Stream
	for _, p := range matches {
		s, err := OpenStream(p, userID, "periodic")
		if err != nil {
			CloseAll(out)
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		out = append(out, s)
	}
	return out, nil
}

// CloseAll closes non-nil streams.
func CloseAll(streams []*Stream) {
	for _, s := range streams {
		if s != nil {
			_ = s.Close()
		}
	}
}

// ForEachMergedRow merges one event CSV with all periodic_data_*.csv files in time order
// and invokes fn for each row (after user-id filtering).
func ForEachMergedRow(eventPath, periodicDir, userID string, fn func(src *Stream, tNano int64) error) error {
	ev, err := OpenStream(eventPath, userID, "event")
	if err != nil {
		return fmt.Errorf("events: %w", err)
	}
	defer func() { _ = ev.Close() }()

	perStreams, err := OpenPeriodicStreams(periodicDir, userID)
	if err != nil {
		return err
	}
	defer CloseAll(perStreams)

	streams := append([]*Stream{ev}, perStreams...)
	h := &timeHeap{}
	heap.Init(h)
	for i, s := range streams {
		nano, ok, err := s.PeekTime()
		if err != nil {
			return fmt.Errorf("stream %d: %w", i, err)
		}
		if ok {
			heap.Push(h, heapItem{tNano: nano, src: i})
		}
	}

	for h.Len() > 0 {
		it := heap.Pop(h).(heapItem)
		src := streams[it.src]
		if err := fn(src, it.tNano); err != nil {
			return err
		}
		src.advance()
		nano, ok, err := src.PeekTime()
		if err != nil {
			return fmt.Errorf("%s: %w", src.Path, err)
		}
		if ok {
			heap.Push(h, heapItem{tNano: nano, src: it.src})
		}
	}
	return nil
}

// DefaultDownloadsPaths returns (event_csv, periodic_dir) under the user's home Downloads folder.
func DefaultDownloadsPaths() (event string, periodicDir string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	dl := filepath.Join(home, "Downloads")
	ev := filepath.Join(dl, "event_data.csv")
	per := filepath.Join(dl, "periodic_data_monthly_csv")
	if _, err := os.Stat(ev); err != nil {
		return "", "", fmt.Errorf("event_data.csv not found at %s: %w", ev, err)
	}
	st, err := os.Stat(per)
	if err != nil {
		return "", "", fmt.Errorf("periodic_data_monthly_csv: %w", err)
	}
	if !st.IsDir() {
		return "", "", fmt.Errorf("%s is not a directory", per)
	}
	return ev, per, nil
}

// ExpandUser replaces a leading "~" with the home directory.
func ExpandUser(p string) (string, error) {
	if strings.HasPrefix(p, "~/") {
		h, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(h, p[2:]), nil
	}
	return p, nil
}
