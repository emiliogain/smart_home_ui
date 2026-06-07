package hsludata

import (
	"encoding/csv"
	"io"
	"os"
	"strings"
)

// Stream is one time-ordered CSV (event_data or periodic_data_YYYY_MM).
type Stream struct {
	Kind   string // "event" or "periodic"
	Path   string
	userID string
	f      *os.File
	r      *csv.Reader
	Idx    map[string]int
	Cur    []string
	EOF    bool
	Err    error
}

// OpenStream opens a CSV, reads the header row, and buffers the first matching data row.
func OpenStream(path, userID, kind string) (*Stream, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(f)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	hdr, err := r.Read()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	s := &Stream{
		Kind:   kind,
		Path:   path,
		userID: userID,
		f:      f,
		r:      r,
		Idx:    HeaderIndex(hdr),
	}
	s.advance()
	return s, nil
}

// Close releases the file handle.
func (s *Stream) Close() error {
	if s.f == nil {
		return nil
	}
	err := s.f.Close()
	s.f = nil
	return err
}

func (s *Stream) advance() {
	s.Cur = nil
	for {
		rec, err := s.r.Read()
		if err == io.EOF {
			s.EOF = true
			return
		}
		if err != nil {
			s.Err = err
			return
		}
		if s.userID != "" {
			idIdx, ok := s.Idx["id"]
			if !ok {
				s.Err = io.ErrUnexpectedEOF
				return
			}
			if strings.TrimSpace(rec[idIdx]) != strings.TrimSpace(s.userID) {
				continue
			}
		}
		if ti, ok := s.Idx["datetime_utc"]; ok && ti < len(rec) {
			if _, err := ParseDateTimeUTC(rec[ti]); err != nil {
				continue
			}
		}
		s.Cur = rec
		return
	}
}

// PeekTime returns the current row's timestamp, or ok=false if no row.
func (s *Stream) PeekTime() (nano int64, ok bool, err error) {
	if s.Err != nil {
		return 0, false, s.Err
	}
	if s.EOF || s.Cur == nil {
		return 0, false, nil
	}
	ti, ok := s.Idx["datetime_utc"]
	if !ok || ti >= len(s.Cur) {
		return 0, false, nil
	}
	at, err := ParseDateTimeUTC(s.Cur[ti])
	if err != nil {
		return 0, false, err
	}
	return at.UnixNano(), true, nil
}
