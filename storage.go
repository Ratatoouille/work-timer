package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Storage struct {
	filePath string
}

type SaveData struct {
	StartTime string      `json:"start_time"`
	WorkTime  string      `json:"work_time"`
	Worked    string      `json:"worked"`
	Plan      string      `json:"plan"`
	AddTZ     bool        `json:"add_tz"`
	Breaks    []BreakData `json:"breaks"`
}

type BreakData struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func NewStorage(filePath string) *Storage {
	return &Storage{filePath: filePath}
}

func (s *Storage) Save(data SaveData) error {
	if s.filePath == "" {
		return fmt.Errorf("no save file specified")
	}

	dir := filepath.Dir(s.filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir: %w", err)
		}
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	if err := os.WriteFile(s.filePath, jsonData, 0o644); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

func (s *Storage) Load() (SaveData, error) {
	var data SaveData

	if s.filePath == "" {
		return data, fmt.Errorf("no load file specified")
	}

	fileData, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return data, fmt.Errorf("file not found: %s", s.filePath)
		}
		return data, fmt.Errorf("read: %w", err)
	}

	if err := json.Unmarshal(fileData, &data); err != nil {
		return data, fmt.Errorf("parse: %w", err)
	}

	return data, nil
}
