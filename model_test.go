package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToAbsolutePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(string) bool
	}{
		{
			name:  "tilde expansion",
			input: "~/test.json",
			check: func(s string) bool {
				return strings.Contains(s, "test.json") && !strings.Contains(s, "~")
			},
		},
		{
			name:  "absolute path unchanged",
			input: "/tmp/test.json",
			check: func(s string) bool {
				return s == "/tmp/test.json"
			},
		},
		{
			name:  "relative path converted",
			input: "test.json",
			check: func(s string) bool {
				return filepath.IsAbs(s) && strings.HasSuffix(s, "test.json")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toAbsolutePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("toAbsolutePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.check(got) {
				t.Errorf("toAbsolutePath() = %v, doesn't match expectations", got)
			}
		})
	}
}

func TestModelTotalFields(t *testing.T) {
	m := NewModel("")

	// Initially 5 fields (start, work, worked, plan, addFour)
	if m.totalFields() != 5 {
		t.Errorf("totalFields() = %v, want 5", m.totalFields())
	}

	// Add one break (2 fields)
	m.addBreak()
	if m.totalFields() != 7 {
		t.Errorf("totalFields() after one break = %v, want 7", m.totalFields())
	}

	// Add another break
	m.addBreak()
	if m.totalFields() != 9 {
		t.Errorf("totalFields() after two breaks = %v, want 9", m.totalFields())
	}
}

func TestModelAddAndDeleteBreak(t *testing.T) {
	m := NewModel("")

	// Initially no breaks
	if len(m.breaks) != 0 {
		t.Errorf("Initial breaks = %v, want 0", len(m.breaks))
	}

	// Add break
	m.addBreak()
	if len(m.breaks) != 1 {
		t.Errorf("After addBreak() = %v, want 1", len(m.breaks))
	}

	// Delete break
	m.cursor = 5 // First break field
	m.deleteCurrentBreak()
	if len(m.breaks) != 0 {
		t.Errorf("After deleteCurrentBreak() = %v, want 0", len(m.breaks))
	}
}

func TestModelMoveCursor(t *testing.T) {
	m := NewModel("")

	// Start at 0
	if m.cursor != 0 {
		t.Errorf("Initial cursor = %v, want 0", m.cursor)
	}

	// Move down
	m.moveCursor(1)
	if m.cursor != 1 {
		t.Errorf("After moveCursor(1) = %v, want 1", m.cursor)
	}

	// Move up
	m.moveCursor(-1)
	if m.cursor != 0 {
		t.Errorf("After moveCursor(-1) = %v, want 0", m.cursor)
	}

	// Try to move beyond bounds (should stay at 0)
	m.moveCursor(-1)
	if m.cursor != 0 {
		t.Errorf("After moveCursor(-1) at boundary = %v, want 0", m.cursor)
	}
}

func TestModelCurrentBreakIndex(t *testing.T) {
	m := NewModel("")
	m.addBreak()
	m.addBreak()

	tests := []struct {
		cursor    int
		wantIndex int
		wantOk    bool
	}{
		{cursor: 0, wantIndex: 0, wantOk: false}, // StartTime field
		{cursor: 4, wantIndex: 0, wantOk: false}, // AddFour field
		{cursor: 5, wantIndex: 0, wantOk: true},  // First break from
		{cursor: 6, wantIndex: 0, wantOk: true},  // First break to
		{cursor: 7, wantIndex: 1, wantOk: true},  // Second break from
		{cursor: 8, wantIndex: 1, wantOk: true},  // Second break to
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			m.cursor = tt.cursor
			gotIndex, gotOk := m.currentBreakIndex()
			if gotIndex != tt.wantIndex || gotOk != tt.wantOk {
				t.Errorf("currentBreakIndex() at cursor %v = (%v, %v), want (%v, %v)",
					tt.cursor, gotIndex, gotOk, tt.wantIndex, tt.wantOk)
			}
		})
	}
}

func TestModelRecalculate(t *testing.T) {
	m := NewModel("")

	// Set valid data
	m.startTime.SetValue("09:00")
	m.workTime.SetValue("08:00")

	m.recalculate()

	if m.result != "17:00" {
		t.Errorf("recalculate() result = %v, want 17:00", m.result)
	}
	if m.err != "" {
		t.Errorf("recalculate() err = %v, want empty", m.err)
	}
}

func TestModelRecalculateWithError(t *testing.T) {
	m := NewModel("")

	// Set invalid data
	m.startTime.SetValue("25:00")
	m.workTime.SetValue("08:00")

	m.recalculate()

	if m.result != "" {
		t.Errorf("recalculate() result should be empty, got %v", m.result)
	}
	if m.err == "" {
		t.Error("recalculate() should set error for invalid time")
	}
}

func TestModelLoadAvailableFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "work-timer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "test1.json"), []byte("{}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "test2.json"), []byte("{}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte(""), 0o644)

	m := NewModel("")
	m.workDir = tmpDir
	m.loadAvailableFiles()

	// Should only find .json files
	if len(m.availableFiles) != 2 {
		t.Errorf("loadAvailableFiles() found %v files, want 2", len(m.availableFiles))
	}

	// Check that both json files are in the list
	hasTest1 := false
	hasTest2 := false
	for _, f := range m.availableFiles {
		if f == "test1.json" {
			hasTest1 = true
		}
		if f == "test2.json" {
			hasTest2 = true
		}
	}

	if !hasTest1 || !hasTest2 {
		t.Error("loadAvailableFiles() didn't find all json files")
	}
}
