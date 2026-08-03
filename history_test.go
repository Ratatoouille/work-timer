package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHistoryAppendAndLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "work-timer-hist-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHistoryStorage(tmpDir)

	e1 := HistoryEntry{
		Date:      "2025-01-15",
		StartTime: "09:00",
		EndTime:   "18:00",
		Breaks:    1,
		BreaksDur: "1h",
		Worked:    "08:00",
		SavedAt:   "18:01",
	}
	if err := hs.Append(e1); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	entries, err := hs.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Date != e1.Date {
		t.Errorf("Date = %v, want %v", entries[0].Date, e1.Date)
	}
}

func TestHistoryAppendUpdatesExistingDate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "work-timer-hist-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHistoryStorage(tmpDir)

	e1 := HistoryEntry{Date: "2025-01-15", StartTime: "09:00", EndTime: "18:00", SavedAt: "18:01"}
	e2 := HistoryEntry{Date: "2025-01-15", StartTime: "08:00", EndTime: "17:00", SavedAt: "17:01"}

	hs.Append(e1)
	hs.Append(e2)

	entries, _ := hs.Load()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (updated), got %d", len(entries))
	}
	if entries[0].StartTime != "08:00" {
		t.Errorf("StartTime = %v, want 08:00 (updated)", entries[0].StartTime)
	}
}

func TestHistorySortedNewestFirst(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "work-timer-hist-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHistoryStorage(tmpDir)

	hs.Append(HistoryEntry{Date: "2025-01-10"})
	hs.Append(HistoryEntry{Date: "2025-01-15"})
	hs.Append(HistoryEntry{Date: "2025-01-12"})

	entries, _ := hs.Load()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Date != "2025-01-15" {
		t.Errorf("first entry Date = %v, want 2025-01-15", entries[0].Date)
	}
	if entries[2].Date != "2025-01-10" {
		t.Errorf("last entry Date = %v, want 2025-01-10", entries[2].Date)
	}
}

func TestHistoryLoadNonExistent(t *testing.T) {
	hs := NewHistoryStorage("/nonexistent/path")
	_, err := hs.Load()
	if err == nil {
		t.Error("Load() should return error for non-existent file")
	}
}

func TestHistoryFilePath(t *testing.T) {
	hs := NewHistoryStorage("/tmp/testdir")
	expected := filepath.Join("/tmp/testdir", "history.json")
	if hs.filePath != expected {
		t.Errorf("filePath = %v, want %v", hs.filePath, expected)
	}
}
