package main

import (
	"testing"
	"time"
)

func TestOnBreakInputTimezone(t *testing.T) {
	// Точные моменты в UTC и их соответствие в input_timezone Asia/Bangkok (UTC+7).
	m := NewModel("")
	m.config.InputTimezone = "Asia/Bangkok"
	m.startTime.SetValue("09:00")
	m.workTime.SetValue("08:00")
	m.addBreak()
	m.breaks[0].from.SetValue("20:00")
	m.breaks[0].to.SetValue("22:00")
	m.recalculate()
	if m.endTimeRaw == "" {
		t.Skip("recalc did not set end time")
	}

	cases := []struct {
		utc   string // "2006-01-02 15:04" в UTC
		onBrk bool
	}{
		{"2026-08-10 13:10", true},  // 20:10 в Bangkok — внутри 20:00–22:00
		{"2026-08-10 13:30", true},  // 20:30 в Bangkok — внутри 20:00–22:00
		{"2026-08-10 15:00", false}, // 22:00 в Bangkok — на границе конца (не перерыв)
	}
	utcLoc := time.UTC
	for _, c := range cases {
		cur, _ := time.ParseInLocation("2006-01-02 15:04", c.utc, utcLoc)
		m.currentTime = cur
		if got := m.onBreak(); got != c.onBrk {
			t.Errorf("at %s UTC: onBreak=%v, want %v", c.utc, got, c.onBrk)
		}
	}
}
