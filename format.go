package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// displayDateLayouts — поддерживаемые входные форматы даты для приведения к
// единому отображению ДД-ММ-ГГГГ.
var displayDateLayouts = []string{
	"02-01-2006",
	"2006-01-02",
	"02/01/2006",
	"2006/01/02",
	"02_01_2006",
}

// formatDisplayDate приводит любые входные форматы даты (ДД-ММ-ГГГГ,
// ГГГГ-ММ-ДД, с разделителями -/_/) к единому формату отображения ДД-ММ-ГГГГ.
func formatDisplayDate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	for _, layout := range displayDateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("02-01-2006")
		}
	}
	return s
}

// formatFileDate возвращает имя файла сохранения для даты в формате ДД_ММ_ГГ.json
// с обязательным ведущим нулём дня, месяца и двухзначным годом (03_08_26.json).
func formatFileDate(t time.Time) string {
	return t.Format("02_01_06") + ".json"
}

// fileSortDate извлекает дату из имени файла (ДД_ММ_ГГ.json или старого
// ДД_ММ.json) и строит реальную дату для хронологической сортировки вместо
// строкового сравнения имени. Для имён без года год берётся из modTime.
func fileSortDate(name string, modTime time.Time) time.Time {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	parts := strings.Split(base, "_")
	if len(parts) == 2 {
		day, err1 := strconv.Atoi(parts[0])
		month, err2 := strconv.Atoi(parts[1])
		if err1 == nil && err2 == nil && day >= 1 && day <= 31 && month >= 1 && month <= 12 {
			return time.Date(modTime.Year(), time.Month(month), day, 0, 0, 0, 0, time.UTC)
		}
	}
	if len(parts) == 3 {
		day, err1 := strconv.Atoi(parts[0])
		month, err2 := strconv.Atoi(parts[1])
		year2, err3 := strconv.Atoi(parts[2])
		if err1 == nil && err2 == nil && err3 == nil &&
			day >= 1 && day <= 31 && month >= 1 && month <= 12 && year2 >= 0 && year2 <= 99 {
			year := 2000 + year2
			return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		}
	}
	return modTime
}

// historyDateSortKey парсит дату записи истории (ДД-ММ-ГГГГ или ГГГГ-ММ-ДД)
// для хронологической сортировки. Возвращает нулевую дату, если разобрать нельзя.
func historyDateSortKey(s string) time.Time {
	s = strings.TrimSpace(s)
	if len(s) >= 10 {
		for _, layout := range []string{"02-01-2006", "2006-01-02"} {
			if t, err := time.Parse(layout, s); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

// parseBreakDur парсит компактную строку длительности вида "1h30m", "45m",
// "2h" или "0" в time.Duration. Возвращает 0 при неудаче.
func parseBreakDur(s string) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	// значения из старых версий могут храниться как "0m"
	var h, m int
	rest := s
	if n := strings.IndexByte(rest, 'h'); n >= 0 {
		if v, err := strconv.Atoi(rest[:n]); err == nil {
			h = v
		}
		rest = rest[n+1:]
	}
	if n := strings.IndexByte(rest, 'm'); n >= 0 {
		if v, err := strconv.Atoi(rest[:n]); err == nil {
			m = v
		}
	}
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute
}

// tildePath сокращает абсолютный путь до "~", если он находится внутри
// домашней папки пользователя.
func tildePath(p string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	if p == home {
		return "~"
	}
	hp := home + string(filepath.Separator)
	if strings.HasPrefix(p, hp) {
		return "~" + string(filepath.Separator) + strings.TrimPrefix(p, hp)
	}
	return p
}
