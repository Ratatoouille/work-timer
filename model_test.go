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
	m := NewModel("")

	if len(m.breaks) != 0 {
		t.Errorf("Initial breaks = %v, want 0", len(m.breaks))
	}

	m.addBreak()
	if len(m.breaks) != 1 {
		t.Errorf("After addBreak() = %v, want 1", len(m.breaks))
	}

	m.cursor = 5 // First break field (FieldBreaksStart)
	m.deleteCurrentBreak()
	if len(m.breaks) != 0 {
		t.Errorf("After deleteCurrentBreak() = %v, want 0", len(m.breaks))
	}
}

func TestModelMoveCursor(t *testing.T) {
	m := NewModel("")

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
	m := NewModel("")
	m.addBreak()
	m.addBreak()

	tests := []struct {
		cursor    int
		wantIndex int
		wantOk    bool
	}{
		{cursor: 0, wantIndex: 0, wantOk: false}, // FieldStartTime
		{cursor: 4, wantIndex: 0, wantOk: false}, // FieldAddTZ
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
	m := NewModel("")
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
	m := NewModel("")

	if m.isDirty {
		t.Error("New model should not be dirty")
	}

	m.addBreak()
	if !m.isDirty {
		t.Error("Model should be dirty after addBreak()")
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

	m := NewModel("")
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

func TestCreateTimeInputCharLimit(t *testing.T) {
	ti := createTimeInput()
	if ti.CharLimit != 5 {
		t.Errorf("createTimeInput() CharLimit = %v, want 5", ti.CharLimit)
	}
}

func TestCreateDurationInputCharLimit(t *testing.T) {
	ti := createDurationInput()
	if ti.CharLimit != 6 {
		t.Errorf("createDurationInput() CharLimit = %v, want 6", ti.CharLimit)
	}
}

func TestWorkedAndPlanHaveLargerCharLimit(t *testing.T) {
	m := NewModel("")
	if m.worked.CharLimit < 6 {
		t.Errorf("worked.CharLimit = %v, want >= 6", m.worked.CharLimit)
	}
	if m.plan.CharLimit < 6 {
		t.Errorf("plan.CharLimit = %v, want >= 6", m.plan.CharLimit)
	}
}

func TestIsInvalidTimeValue(t *testing.T) {
	valid := []string{
		"", "0", "6", "06", "23", "6:", "06:", "6:0", "6:5", "6:30", "06:00", "23:59",
	}
	for _, s := range valid {
		if isInvalidTimeValue(s) {
			t.Errorf("isInvalidTimeValue(%q) = true, want false", s)
		}
	}

	invalid := []string{
		"24:00", "99:00", "25:00", "6:60", "6:99", "ab:cd", "1:60", "2:60",
		"123", "1:234",
	}
	for _, s := range invalid {
		if !isInvalidTimeValue(s) {
			t.Errorf("isInvalidTimeValue(%q) = false, want true", s)
		}
	}
}

func TestIsInvalidDurationValue(t *testing.T) {
	valid := []string{
		"", "1", "10", "100", "999", "1:", "10:", "100:", "1:00", "100:30", "999:59",
	}
	for _, s := range valid {
		if isInvalidDurationValue(s) {
			t.Errorf("isInvalidDurationValue(%q) = true, want false", s)
		}
	}

	invalid := []string{
		":00", "1:60", "100:99", "abc", "1:6x",
	}
	for _, s := range invalid {
		if !isInvalidDurationValue(s) {
			t.Errorf("isInvalidDurationValue(%q) = false, want true", s)
		}
	}
}

func TestFillCurrentWithNow(t *testing.T) {
	m := NewModel("")

	// На FieldAddTZ — ничего не делает
	m.cursor = FieldAddTZ
	m.fillCurrentWithNow()
	if m.startTime.Value() != "" {
		t.Error("fillCurrentWithNow() on FieldAddTZ should not touch startTime")
	}

	// На FieldStartTime — вставляет время
	m.cursor = FieldStartTime
	m.fillCurrentWithNow()
	val := m.startTime.Value()
	if val == "" {
		t.Error("fillCurrentWithNow() on FieldStartTime should set a value")
	}
	// Должно быть формат HH:MM (5 символов)
	if len(val) != 5 || val[2] != ':' {
		t.Errorf("fillCurrentWithNow() value %q is not HH:MM format", val)
	}
	if !m.isDirty {
		t.Error("fillCurrentWithNow() should mark model dirty")
	}
}

func TestFillCurrentWithNowOnBreakField(t *testing.T) {
	m := NewModel("")
	m.addBreak()

	// FieldBreaksStart = 5 (первый from перерыва)
	m.cursor = FieldBreaksStart
	m.fillCurrentWithNow()

	if m.breaks[0].from.Value() == "" {
		t.Error("fillCurrentWithNow() on break field should set value")
	}
}

func TestRecalculateSnapshotSkips(t *testing.T) {
	m := NewModel("")
	m.startTime.SetValue("09:00")
	m.workTime.SetValue("08:00")

	m.recalculate()
	first := m.result

	// Сбросим результат вручную — повторный вызов с теми же значениями должен пропустить вычисление
	m.result = "CHANGED"
	m.recalculate()

	if m.result != "CHANGED" {
		t.Error("recalculate() should skip when snapshot unchanged, result should stay CHANGED")
	}

	// После изменения значения — должен пересчитать
	m.lastSnapshot = "" // сбрасываем snapshot принудительно
	m.recalculate()
	if m.result != first {
		t.Errorf("recalculate() after snapshot reset = %q, want %q", m.result, first)
	}
}

func TestSetStatus(t *testing.T) {
	m := NewModel("")

	m.setStatus("сохранено", StatusSuccess)
	if m.statusMessage != "сохранено" {
		t.Errorf("setStatus message = %q, want %q", m.statusMessage, "сохранено")
	}
	if m.statusType != StatusSuccess {
		t.Errorf("setStatus type = %v, want StatusSuccess", m.statusType)
	}

	m.setStatus("ошибка", StatusError)
	if m.statusType != StatusError {
		t.Errorf("setStatus type = %v, want StatusError", m.statusType)
	}
}

func TestClearCurrentField(t *testing.T) {
	m := NewModel("")
	m.startTime.SetValue("09:00")

	// Основное поле очищается
	m.cursor = FieldStartTime
	m.clearCurrentField()
	if m.startTime.Value() != "" {
		t.Errorf("clearCurrentField() startTime = %q, want empty", m.startTime.Value())
	}
	if !m.isDirty {
		t.Error("clearCurrentField() should mark model dirty")
	}

	// Повторный вызов на пустом поле — isDirty не меняется повторно
	m.isDirty = false
	m.clearCurrentField()
	if m.isDirty {
		t.Error("clearCurrentField() on empty field should not mark dirty")
	}

	// FieldAddTZ — не трогает
	m.cursor = FieldAddTZ
	m.startTime.SetValue("09:00")
	m.clearCurrentField()
	if m.startTime.Value() != "09:00" {
		t.Error("clearCurrentField() should not affect fields when cursor on FieldAddTZ")
	}

	// Поле перерыва — очищается
	m.addBreak()
	m.breaks[0].from.SetValue("12:00")
	m.cursor = FieldBreaksStart
	m.isDirty = false
	m.clearCurrentField()
	if m.breaks[0].from.Value() != "" {
		t.Error("clearCurrentField() should clear break fields")
	}
	if !m.isDirty {
		t.Error("clearCurrentField() on break field should mark dirty")
	}
}
