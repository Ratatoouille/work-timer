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
		AddFour:   true,
		Breaks: []BreakData{
			{From: "12:00", To: "13:00"},
			{From: "15:00", To: "15:30"},
		},
	}

	// Test Save
	err = storage.Save(originalData)
	if err != nil {
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
	if loadedData.AddFour != originalData.AddFour {
		t.Errorf("AddFour = %v, want %v", loadedData.AddFour, originalData.AddFour)
	}
	if len(loadedData.Breaks) != len(originalData.Breaks) {
		t.Errorf("Breaks length = %v, want %v", len(loadedData.Breaks), len(originalData.Breaks))
	}
}

func TestStorageSaveCreatesDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "work-timer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create path with non-existent subdirectory
	testFile := filepath.Join(tmpDir, "subdir", "test.json")
	storage := NewStorage(testFile)

	data := SaveData{
		StartTime: "09:00",
		WorkTime:  "08:00",
	}

	err = storage.Save(data)
	if err != nil {
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

	data := SaveData{StartTime: "09:00"}
	err := storage.Save(data)
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

	// Write invalid JSON
	err = os.WriteFile(testFile, []byte("not valid json"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	storage := NewStorage(testFile)
	_, err = storage.Load()
	if err == nil {
		t.Error("Load() should return error for invalid JSON")
	}
}
