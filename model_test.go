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
			check: func(s string) bool { return s == "/tmp/test.json" },
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
	m := NewModelForTesting("")

	// Initially 5 fields (start, workTime, worked, plan, addTZ)
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
	m := NewModelForTesting("")

	if len(m.breaks) != 0 {
		t.Errorf("Initial breaks = %v, want 0", len(m.breaks))
	}

	m.addBreak()
	if len(m.breaks) != 1 {
		t.Errorf("After addBreak() = %v, want 1", len(m.breaks))
	}

	if m.breaks[0].from.Placeholder != m.locale.PlaceholderTime {
		t.Errorf("Break from placeholder = %q, want %q",
			m.breaks[0].from.Placeholder, m.locale.PlaceholderTime)
	}

	m.cursor = 5
	m.deleteCurrentBreak()
	if len(m.breaks) != 0 {
		t.Errorf("After deleteCurrentBreak() = %v, want 0", len(m.breaks))
	}
}

func TestModelMoveCursor(t *testing.T) {
	m := NewModelForTesting("")

	if m.cursor != 0 {
		t.Errorf("Initial cursor = %v, want 0", m.cursor)
	}

	m.moveCursor(1)
	if m.cursor != 1 {
		t.Errorf("After moveCursor(1) = %v, want 1", m.cursor)
	}

	m.moveCursor(-1)
	if m.cursor != 0 {
		t.Errorf("After moveCursor(-1) = %v, want 0", m.cursor)
	}

	// Stay at boundary
	m.moveCursor(-1)
	if m.cursor != 0 {
		t.Errorf("After moveCursor(-1) at boundary = %v, want 0", m.cursor)
	}
}

func TestModelCurrentBreakIndex(t *testing.T) {
	m := NewModelForTesting("")
	m.addBreak()
	m.addBreak()

	tests := []struct {
		cursor    int
		wantIndex int
		wantOk    bool
	}{
		{cursor: 0, wantIndex: 0, wantOk: false},
		{cursor: 4, wantIndex: 0, wantOk: false},
		{cursor: 5, wantIndex: 0, wantOk: true},
		{cursor: 6, wantIndex: 0, wantOk: true},
		{cursor: 7, wantIndex: 1, wantOk: true},
		{cursor: 8, wantIndex: 1, wantOk: true},
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
	m := NewModelForTesting("")
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
	m := NewModelForTesting("")
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

func TestModelRecalculateWorkedExceedsPlan(t *testing.T) {
	m := NewModelForTesting("")
	m.startTime.SetValue("09:00")
	m.worked.SetValue("10:00")
	m.plan.SetValue("08:00")

	m.recalculate()

	if m.result != "" {
		t.Errorf("recalculate() result should be empty when worked > plan, got %v", m.result)
	}
	if m.err == "" {
		t.Error("recalculate() should set error when worked > plan")
	}
}

func TestModelIsDirtyOnEdit(t *testing.T) {
	m := NewModelForTesting("")

	if m.isDirty {
		t.Error("New model should not be dirty")
	}

	m.addBreak()
	if !m.isDirty {
		t.Error("Model should be dirty after addBreak()")
	}
}

func TestModelSetStatus(t *testing.T) {
	m := NewModelForTesting("")

	m.setStatus("test success", StatusSuccess)
	if m.statusMessage != "test success" {
		t.Errorf("statusMessage = %q, want %q", m.statusMessage, "test success")
	}
	if m.statusType != StatusSuccess {
		t.Errorf("statusType = %v, want StatusSuccess", m.statusType)
	}

	m.setStatus("test error", StatusError)
	if m.statusType != StatusError {
		t.Errorf("statusType = %v, want StatusError", m.statusType)
	}

	m.setStatus("test warn", StatusWarn)
	if m.statusType != StatusWarn {
		t.Errorf("statusType = %v, want StatusWarn", m.statusType)
	}
}

func TestModelDefaultFileFromConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "work-timer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "default.json")
	os.WriteFile(testFile, []byte(`{"start_time":"10:00","work_time":"08:00"}`), 0o644)

	cfg := defaultConfig()
	cfg.WorkDir = tmpDir
	cfg.DefaultFile = testFile

	m := newModelWithConfig("", cfg, localeEN)

	if m.saveFile != testFile {
		t.Errorf("saveFile = %q, want %q", m.saveFile, testFile)
	}
	if m.startTime.Value() != "10:00" {
		t.Errorf("startTime = %q, want 10:00 (loaded from default_file)", m.startTime.Value())
	}
}

func TestModelArgOverridesDefaultFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "work-timer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	argFile := filepath.Join(tmpDir, "arg.json")
	defaultFile := filepath.Join(tmpDir, "default.json")
	os.WriteFile(argFile, []byte(`{"start_time":"09:00"}`), 0o644)
	os.WriteFile(defaultFile, []byte(`{"start_time":"11:00"}`), 0o644)

	cfg := defaultConfig()
	cfg.WorkDir = tmpDir
	cfg.DefaultFile = defaultFile

	m := newModelWithConfig(argFile, cfg, localeEN)

	if m.saveFile != argFile {
		t.Errorf("saveFile = %q, want argFile %q", m.saveFile, argFile)
	}
	if m.startTime.Value() != "09:00" {
		t.Errorf("startTime = %q, want 09:00 (from arg, not default_file)", m.startTime.Value())
	}
}

func TestModelLoadAvailableFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "work-timer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "test1.json"), []byte("{}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "test2.json"), []byte("{}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte(""), 0o644)

	m := NewModelForTesting("")
	m.workDir = tmpDir
	m.loadAvailableFiles()

	if len(m.availableFiles) != 2 {
		t.Errorf("loadAvailableFiles() found %v files, want 2", len(m.availableFiles))
	}

	hasTest1, hasTest2 := false, false
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
