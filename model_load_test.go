package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadShowsCalculations(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "03_08_26.json")
	data := `{"start_time":"09:00","work_time":"08:00","worked":"","plan":"","add_tz":false,"breaks":[]}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	m := NewModel(path)
	m.currentTime, _ = time.Parse("2006-01-02 15:04", "2026-08-03 10:00")
	m.recalculate()

	if m.result == "" {
		t.Fatalf("expected result computed right after load, got empty. err=%q", m.err)
	}
	if m.endTimeRaw != "17:00" {
		t.Errorf("endTimeRaw = %q, want 17:00", m.endTimeRaw)
	}
}
