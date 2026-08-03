package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// HistoryEntry — одна запись в журнале рабочих дней.
type HistoryEntry struct {
	Date     string `json:"date"`      // YYYY-MM-DD
	FileName string `json:"file"`      // имя файла сохранения
	StartTime string `json:"start"`
	EndTime   string `json:"end"`
	Breaks    int    `json:"breaks"`
	BreaksDur string `json:"breaks_dur"` // "1h30m"
	Worked    string `json:"worked"`     // итоговая длительность работы
	SavedAt   string `json:"saved_at"`
}

// HistoryStorage хранит массив записей в history.json.
type HistoryStorage struct {
	filePath string
}

func NewHistoryStorage(workDir string) *HistoryStorage {
	return &HistoryStorage{filePath: filepath.Join(workDir, "history.json")}
}

// Append добавляет или обновляет запись за текущую дату.
func (h *HistoryStorage) Append(entry HistoryEntry) error {
	entries, err := h.Load()
	if err != nil && !os.IsNotExist(err) {
		// если файл повреждён — перезаписываем
		entries = []HistoryEntry{}
	}

	updated := false
	for i, e := range entries {
		if e.Date == entry.Date {
			entries[i] = entry
			updated = true
			break
		}
	}
	if !updated {
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date > entries[j].Date
	})

	return h.write(entries)
}

func (h *HistoryStorage) Load() ([]HistoryEntry, error) {
	data, err := os.ReadFile(h.filePath)
	if err != nil {
		return nil, err
	}

	var entries []HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse history: %w", err)
	}

	return entries, nil
}

func (h *HistoryStorage) write(entries []HistoryEntry) error {
	dir := filepath.Dir(h.filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir: %w", err)
		}
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	return os.WriteFile(h.filePath, data, 0o644)
}

// recordHistory создаёт запись из текущего состояния модели и сохраняет её.
func (m *Model) recordHistory() {
	if m.result == "" || m.err != "" {
		return
	}

	breaksDur := m.calculator.BreaksDuration(m.getBreaksData())

	entry := HistoryEntry{
		Date:      time.Now().Format("2006-01-02"),
		FileName:  filepath.Base(m.saveFile),
		StartTime: m.startTime.Value(),
		EndTime:   m.result,
		Breaks:    len(m.breaks),
		BreaksDur: formatDurationGo(breaksDur),
		Worked:    m.workTime.Value(),
		SavedAt:   time.Now().Format("15:04"),
	}

	histStorage := NewHistoryStorage(m.workDir)
	if err := histStorage.Append(entry); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not save history: %v\n", err)
	}
}

// formatDurationGo форматирует Duration в компактный вид "1h30m".
func formatDurationGo(d time.Duration) string {
	if d == 0 {
		return "0m"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%dh%dm", h, m)
	case h > 0:
		return fmt.Sprintf("%dh", h)
	default:
		return fmt.Sprintf("%dm", m)
	}
}
