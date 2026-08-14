package main

import (
	"testing"
	"time"
)

func TestFormatDisplayDate(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"2026-08-06", "06-08-2026"}, // из истории (YYYY-MM-DD)
		{"06-08-2026", "06-08-2026"}, // уже ДД-ММ-ГГГГ
		{"06/08/2026", "06-08-2026"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := formatDisplayDate(tt.in); got != tt.want {
			t.Errorf("formatDisplayDate(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatFileDate(t *testing.T) {
	got := formatFileDate(time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC))
	if got != "03_08_26.json" {
		t.Errorf("got %q, want 03_08_26.json (ведущие нули + год обязательны)", got)
	}
}

func TestFileSortDateChronologicalAcrossYears(t *testing.T) {
	// При смене года строковая сортировка имён ДД_ММ уже не даёт верный
	// хронологический порядок — сортируем по распарсенной дате (год из имени).
	jan := fileSortDate("04_01_26.json", time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC))
	aug := fileSortDate("10_08_26.json", time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC))
	if !aug.After(jan) {
		t.Errorf("august date should sort after january")
	}

	// Год из имени: 2025 должна идти раньше 2026, несмотря на день/месяц.
	lastYear := fileSortDate("04_01_25.json", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	if !aug.After(lastYear) {
		t.Errorf("2026 should sort after 2025")
	}
}

func TestHistoryDateSortKey(t *testing.T) {
	a := historyDateSortKey("15-01-2025")
	b := historyDateSortKey("2025-02-01")
	if !b.After(a) {
		t.Errorf("feb 2025 should be after jan 2025")
	}
	if !historyDateSortKey("").IsZero() {
		t.Errorf("empty date should be zero")
	}
}

func TestParseBreakDur(t *testing.T) {
	if parseBreakDur("1h30m") != 90*time.Minute {
		t.Error("1h30m parse failed")
	}
	if parseBreakDur("45m") != 45*time.Minute {
		t.Error("45m parse failed")
	}
	if parseBreakDur("0m") != 0 {
		t.Error("0m parse failed")
	}
}
