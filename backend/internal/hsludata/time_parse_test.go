package hsludata

import (
	"testing"
	"time"
)

func TestParseDateTimeUTC_sample(t *testing.T) {
	at, err := ParseDateTimeUTC("2023-05-30T13:30:02.403000000")
	if err != nil {
		t.Fatal(err)
	}
	if at.Year() != 2023 || at.Month() != time.May || at.Day() != 30 || at.Hour() != 13 || at.Minute() != 30 {
		t.Fatalf("got %v", at)
	}
}
