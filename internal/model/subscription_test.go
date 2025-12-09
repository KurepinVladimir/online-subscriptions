package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseYearMonth_OK(t *testing.T) {
	tests := []struct {
		name string
		in   string
		year int
		mon  time.Month
	}{
		{"jan", "01-2025", 2025, time.January},
		{"dec", "12-2030", 2030, time.December},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ym, err := ParseYearMonth(tt.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if ym.Year() != tt.year || ym.Month() != tt.mon || ym.Day() != 1 {
				t.Fatalf("unexpected result: got %v, want %d-%02d-01", ym.Time, tt.year, tt.mon)
			}
		})
	}
}

func TestParseYearMonth_Invalid(t *testing.T) {
	_, err := ParseYearMonth("2025-01")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestYearMonth_JSON(t *testing.T) {
	ym := YearMonth{Time: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)}

	b, err := json.Marshal(ym)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if string(b) != `"07-2025"` {
		t.Fatalf("unexpected json: %s", string(b))
	}

	var got YearMonth
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if !got.Equal(ym.Time) {
		t.Fatalf("expected %v, got %v", ym.Time, got.Time)
	}
}
