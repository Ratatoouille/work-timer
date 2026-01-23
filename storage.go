package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Storage struct {
	filePath string
}

type SaveData struct {
	StartTime string      `json:"start_time"`
	WorkTime  string      `json:"work_time"`
	Worked    string      `json:"worked"`
	Plan      string      `json:"plan"`
	AddFour   bool        `json:"add_four"`
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
		return fmt.Errorf("не указан файл для сохранения")
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	if err := os.WriteFile(s.filePath, jsonData, 0o644); err != nil {
		return fmt.Errorf("ошибка записи файла: %w", err)
	}

	return nil
}

func (s *Storage) Load() (SaveData, error) {
	var data SaveData

	if s.filePath == "" {
		return data, fmt.Errorf("не указан файл для загрузки")
	}

	fileData, err := os.ReadFile(s.filePath)
	if err != nil {
		return data, fmt.Errorf("ошибка чтения файла: %w", err)
	}

	if err := json.Unmarshal(fileData, &data); err != nil {
		return data, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	return data, nil
}
