package hsludata

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestForEachMergedRow_fixture(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Skip("caller")
	}
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "data", "hslu14243471")
	ev := filepath.Join(root, "fixture_events.csv")
	per := filepath.Join(root, "fixture_periodic.csv")
	// Single-file "dir": glob expects periodic_data_*.csv — copy name pattern.
	perDir := t.TempDir()
	perPath := filepath.Join(perDir, "periodic_data_2023_05.csv")
	src, err := os.ReadFile(per)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(perPath, src, 0o644); err != nil {
		t.Fatal(err)
	}
	var count int
	err = ForEachMergedRow(ev, perDir, "99", func(src *Stream, tNano int64) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 5 {
		t.Fatalf("expected 5 merged rows, got %d", count)
	}
}
