package main

import (
	"strings"
	"testing"
	"time"
)

func TestFormatTrayRemaining(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"zero", 0, "0:00"},
		{"minutes only", 45 * time.Minute, "0:45"},
		{"one hour", time.Hour, "1:00"},
		{"hour and minutes", time.Hour + 30*time.Minute, "1:30"},
		{"ten hours", 10*time.Hour + 15*time.Minute, "10:15"},
		{"seconds rounded down", 2*time.Hour + 30*time.Minute + 20*time.Second, "2:30"},
		{"seconds rounded up", 2*time.Hour + 30*time.Minute + 40*time.Second, "2:31"},
		{"negative clamps to zero", -5 * time.Minute, "0:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatTrayRemaining(tt.d); got != tt.want {
				t.Errorf("formatTrayRemaining(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestTrayStateLabel(t *testing.T) {
	tests := []struct {
		name  string
		state TrayState
		want  string
	}{
		{"no data", TrayState{Ok: false}, ""},
		{"remaining", TrayState{Ok: true, Remaining: 90 * time.Minute}, "1:30"},
		{"day ended overrides", TrayState{Ok: true, Remaining: 5 * time.Minute, DayEnded: true}, "0:00"},
		{"day ended without data", TrayState{Ok: false, DayEnded: true}, "0:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.TrayLabel(); got != tt.want {
				t.Errorf("TrayLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTrayStateTooltip(t *testing.T) {
	loc := localeEN

	if got := (TrayState{}).TrayTooltip(loc); got != loc.TrayNoData {
		t.Errorf("no data tooltip = %q, want %q", got, loc.TrayNoData)
	}

	state := TrayState{Ok: true, Remaining: 90 * time.Minute, EndTime: "17:30"}
	got := state.TrayTooltip(loc)
	if !strings.Contains(got, "1:30") || !strings.Contains(got, "17:30") {
		t.Errorf("tooltip = %q, want remaining and end time", got)
	}

	done := TrayState{Ok: true, DayEnded: true}
	if got := done.TrayTooltip(loc); !strings.Contains(got, loc.TrayDone) {
		t.Errorf("done tooltip = %q, want to contain %q", got, loc.TrayDone)
	}
}

func TestTrayManagerSkipsDuplicateUpdates(t *testing.T) {
	backend := &fakeTrayBackend{}
	manager := NewTrayManager(backend, localeEN)
	if manager == nil {
		t.Fatal("NewTrayManager() = nil")
	}

	state := TrayState{Ok: true, Remaining: 10 * time.Minute, EndTime: "18:00"}
	manager.Update(state)
	manager.Update(state)
	manager.Update(state)

	if backend.applies != 1 {
		t.Errorf("Apply called %d times, want 1", backend.applies)
	}

	manager.Update(TrayState{Ok: true, Remaining: 9 * time.Minute, EndTime: "18:00"})
	if backend.applies != 2 {
		t.Errorf("Apply called %d times after change, want 2", backend.applies)
	}
}

func TestTrayManagerQuitCallback(t *testing.T) {
	backend := &fakeTrayBackend{}
	manager := NewTrayManager(backend, localeEN)

	called := false
	manager.SetOnQuit(func() { called = true })
	manager.handleQuit()

	if !called {
		t.Error("quit callback was not called")
	}

	manager.Quit()
	if !backend.quit {
		t.Error("backend.Quit() was not called")
	}
}

func TestNilTrayManagerIsSafe(t *testing.T) {
	var manager *TrayManager
	manager.Update(TrayState{Ok: true, Remaining: time.Minute})
	manager.Quit()
	manager.SetOnQuit(func() {})
	manager.handleQuit()
}

// fakeTrayBackend фиксирует вызовы для тестов.
type fakeTrayBackend struct {
	applies int
	quit    bool
}

func (f *fakeTrayBackend) Apply(label, tooltip string, dayEnded bool) {
	f.applies++
}

func (f *fakeTrayBackend) Quit() {
	f.quit = true
}
