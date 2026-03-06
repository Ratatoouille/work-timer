package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.WorkDir != DefaultWorkDir {
		t.Errorf("WorkDir = %v, want %v", cfg.WorkDir, DefaultWorkDir)
	}
	if cfg.Language != "ru" {
		t.Errorf("Language = %v, want ru", cfg.Language)
	}
	if cfg.InputTimezone != "Europe/Moscow" {
		t.Errorf("InputTimezone = %v, want Europe/Moscow", cfg.InputTimezone)
	}
	if cfg.Timezone != "" {
		t.Errorf("Timezone = %v, want empty", cfg.Timezone)
	}
	if cfg.DefaultFile != "" {
		t.Errorf("DefaultFile = %v, want empty", cfg.DefaultFile)
	}
	if cfg.UI.LabelWidth != 22 {
		t.Errorf("LabelWidth = %v, want 22", cfg.UI.LabelWidth)
	}
	if cfg.UI.Colors.Accent != "12" {
		t.Errorf("Accent = %v, want 12", cfg.UI.Colors.Accent)
	}
	if cfg.UI.Timeouts.Clipboard != 2 {
		t.Errorf("Clipboard timeout = %v, want 2", cfg.UI.Timeouts.Clipboard)
	}
	if cfg.UI.Timeouts.Status != 3 {
		t.Errorf("Status timeout = %v, want 3", cfg.UI.Timeouts.Status)
	}
	if cfg.UI.Timeouts.Warning != 4 {
		t.Errorf("Warning timeout = %v, want 4", cfg.UI.Timeouts.Warning)
	}
}

func TestLoadConfigCreatesDefault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "work-timer-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	t.Setenv("HOME", tmpDir)

	cfg := LoadConfig()

	if cfg.WorkDir != DefaultWorkDir {
		t.Errorf("WorkDir = %v, want %v", cfg.WorkDir, DefaultWorkDir)
	}
	if cfg.InputTimezone != "Europe/Moscow" {
		t.Errorf("InputTimezone = %v, want Europe/Moscow", cfg.InputTimezone)
	}

	configFile := filepath.Join(tmpDir, ".config", "work_timer", "config.toml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("LoadConfig() should create default config file")
	}
}

func TestLoadConfigParsesValues(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "work-timer-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "work_timer")
	os.MkdirAll(configDir, 0o755)

	content := `
work_dir        = "~/my_timers"
default_file    = "~/my_timers/today.json"
language        = "en"
input_timezone  = "Europe/Moscow"
timezone        = "Asia/Krasnoyarsk"

[ui]
label_width = 30

[ui.colors]
accent = "9"
result = "10"
break  = "11"
warn   = "12"

[ui.timeouts]
clipboard = 5
status    = 6
warning   = 7
`
	os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(content), 0o644)

	cfg := LoadConfig()

	if cfg.WorkDir != "~/my_timers" {
		t.Errorf("WorkDir = %v, want ~/my_timers", cfg.WorkDir)
	}
	if cfg.DefaultFile != "~/my_timers/today.json" {
		t.Errorf("DefaultFile = %v, want ~/my_timers/today.json", cfg.DefaultFile)
	}
	if cfg.Language != "en" {
		t.Errorf("Language = %v, want en", cfg.Language)
	}
	if cfg.InputTimezone != "Europe/Moscow" {
		t.Errorf("InputTimezone = %v, want Europe/Moscow", cfg.InputTimezone)
	}
	if cfg.Timezone != "Asia/Krasnoyarsk" {
		t.Errorf("Timezone = %v, want Asia/Krasnoyarsk", cfg.Timezone)
	}
	if cfg.UI.LabelWidth != 30 {
		t.Errorf("LabelWidth = %v, want 30", cfg.UI.LabelWidth)
	}
	if cfg.UI.Colors.Accent != "9" {
		t.Errorf("Accent = %v, want 9", cfg.UI.Colors.Accent)
	}
	if cfg.UI.Timeouts.Clipboard != 5 {
		t.Errorf("Clipboard = %v, want 5", cfg.UI.Timeouts.Clipboard)
	}
}

func TestLoadConfigFillsZeroValues(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "work-timer-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "work_timer")
	os.MkdirAll(configDir, 0o755)

	content := `timezone = "Asia/Krasnoyarsk"`
	os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(content), 0o644)

	cfg := LoadConfig()

	if cfg.UI.LabelWidth != 22 {
		t.Errorf("LabelWidth should fallback to 22, got %v", cfg.UI.LabelWidth)
	}
	if cfg.UI.Colors.Accent != "12" {
		t.Errorf("Accent should fallback to 12, got %v", cfg.UI.Colors.Accent)
	}
	if cfg.UI.Timeouts.Status != 3 {
		t.Errorf("Status timeout should fallback to 3, got %v", cfg.UI.Timeouts.Status)
	}
	if cfg.WorkDir != DefaultWorkDir {
		t.Errorf("WorkDir should fallback to default, got %v", cfg.WorkDir)
	}
	if cfg.Language != "ru" {
		t.Errorf("Language should fallback to ru, got %v", cfg.Language)
	}
	if cfg.Timezone != "Asia/Krasnoyarsk" {
		t.Errorf("Timezone = %v, want Asia/Krasnoyarsk", cfg.Timezone)
	}
}

func TestTimezoneLabel(t *testing.T) {
	tests := []struct {
		timezone string
		wantErr  bool
	}{
		{timezone: "", wantErr: true},
		{timezone: "Invalid/Zone", wantErr: true},
		{timezone: "Europe/Moscow", wantErr: false},
		{timezone: "Asia/Krasnoyarsk", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.timezone, func(t *testing.T) {
			got := TimezoneLabel(tt.timezone)
			if (got == "") != tt.wantErr {
				t.Errorf("TimezoneLabel(%q) = %q, wantEmpty=%v", tt.timezone, got, tt.wantErr)
			}
		})
	}
}

func TestTimezoneLocation(t *testing.T) {
	if TimezoneLocation("") != nil {
		t.Error("TimezoneLocation(\"\") should return nil")
	}
	if TimezoneLocation("Invalid/Zone") != nil {
		t.Error("TimezoneLocation(invalid) should return nil")
	}
	if TimezoneLocation("Europe/Moscow") == nil {
		t.Error("TimezoneLocation(Europe/Moscow) should not return nil")
	}
}

func TestCurrentTimeInZone(t *testing.T) {
	// Проверяем формат "15:04"
	got := CurrentTimeInZone("Europe/Moscow")
	if len(got) != 5 || got[2] != ':' {
		t.Errorf("CurrentTimeInZone() = %q, want HH:MM format", got)
	}

	// Невалидная зона — падает на time.Local без паники
	got2 := CurrentTimeInZone("Invalid/Zone")
	if len(got2) != 5 || got2[2] != ':' {
		t.Errorf("CurrentTimeInZone(invalid) = %q, want HH:MM format", got2)
	}
}
