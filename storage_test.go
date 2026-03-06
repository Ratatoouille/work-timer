package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStorageSaveAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "work-timer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.json")
	storage := NewStorage(testFile)

	// Test data
	originalData := SaveData{
		StartTime: "09:00",
		WorkTime:  "08:00",
		Worked:    "05:00",
		Plan:      "08:00",
		AddTZ:     true,
		Breaks: []BreakData{
			{From: "12:00", To: "13:00"},
			{From: "15:00", To: "15:30"},
		},
	}

	if err := storage.Save(originalData); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Check file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("File was not created")
	}

	// Test Load
	loadedData, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify data
	if loadedData.StartTime != originalData.StartTime {
		t.Errorf("StartTime = %v, want %v", loadedData.StartTime, originalData.StartTime)
	}
	if loadedData.WorkTime != originalData.WorkTime {
		t.Errorf("WorkTime = %v, want %v", loadedData.WorkTime, originalData.WorkTime)
	}
	if loadedData.AddTZ != originalData.AddTZ {
		t.Errorf("AddTZ = %v, want %v", loadedData.AddTZ, originalData.AddTZ)
	}
	if len(loadedData.Breaks) != len(originalData.Breaks) {
		t.Fatalf("Breaks length = %v, want %v", len(loadedData.Breaks), len(originalData.Breaks))
	}
	if loadedData.Breaks[0].From != "12:00" || loadedData.Breaks[0].To != "13:00" {
		t.Errorf("Break[0] = {%v, %v}, want {12:00, 13:00}",
			loadedData.Breaks[0].From, loadedData.Breaks[0].To)
	}
	if loadedData.Breaks[1].From != "15:00" || loadedData.Breaks[1].To != "15:30" {
		t.Errorf("Break[1] = {%v, %v}, want {15:00, 15:30}",
			loadedData.Breaks[1].From, loadedData.Breaks[1].To)
	}
}

func TestStorageBackwardCompatAddFour(t *testing.T) {
	// Старые файлы используют "add_four": true — должны читаться корректно
	tmpDir, err := os.MkdirTemp("", "work-timer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "legacy.json")
	legacy := `{"start_time":"09:00","work_time":"08:00","add_four":true,"breaks":[]}`
	os.WriteFile(testFile, []byte(legacy), 0o644)

	storage := NewStorage(testFile)
	data, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// LegacyAddFour должен быть true
	if !data.LegacyAddFour {
		t.Error("LegacyAddFour should be true for old files with add_four=true")
	}
	// новое поле должно быть false (не задано в старом файле)
	if data.AddTZ {
		t.Error("AddTZ should be false for old files without add_tz field")
	}
}

func TestStorageNewFileUsesAddTZ(t *testing.T) {
	// Новые файлы должны писать "add_tz", а не "add_four"
	tmpDir, err := os.MkdirTemp("", "work-timer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "new.json")
	storage := NewStorage(testFile)

	storage.Save(SaveData{StartTime: "09:00", AddTZ: true})

	content, _ := os.ReadFile(testFile)
	if !contains(string(content), `"add_tz": true`) {
		t.Error("New file should contain add_tz field")
	}
}

func TestStorageSaveCreatesDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "work-timer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "subdir", "test.json")
	storage := NewStorage(testFile)

	if err := storage.Save(SaveData{StartTime: "09:00"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Check file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("File was not created in subdirectory")
	}
}

func TestStorageLoadNonExistentFile(t *testing.T) {
	storage := NewStorage("/nonexistent/path/file.json")

	_, err := storage.Load()
	if err == nil {
		t.Error("Load() should return error for non-existent file")
	}
}

func TestStorageSaveEmptyPath(t *testing.T) {
	storage := NewStorage("")
	err := storage.Save(SaveData{StartTime: "09:00"})
	if err == nil {
		t.Error("Save() should return error for empty path")
	}
}

func TestStorageLoadInvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "work-timer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "invalid.json")
	os.WriteFile(testFile, []byte("not valid json"), 0o644)

	storage := NewStorage(testFile)
	_, err = storage.Load()
	if err == nil {
		t.Error("Load() should return error for invalid JSON")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
