package main

import (
	"fmt"
	"time"
)

// TrayState — снимок состояния таймера для отображения в трее.
type TrayState struct {
	Remaining time.Duration
	EndTime   string
	Ok        bool // есть ли рассчитанное время окончания
	DayEnded  bool
}

// TrayLabel возвращает короткий текст для иконки/заголовка трея.
// Пустая строка означает, что отображать нечего.
func (s TrayState) TrayLabel() string {
	if s.DayEnded {
		return "0:00"
	}
	if !s.Ok {
		return ""
	}
	return formatTrayRemaining(s.Remaining)
}

// TrayTooltip собирает текст подсказки из локализованных частей.
func (s TrayState) TrayTooltip(loc Locale) string {
	if s.DayEnded {
		return fmt.Sprintf(loc.TrayRemaining, "0:00") + " — " + loc.TrayDone
	}
	if !s.Ok {
		return loc.TrayNoData
	}
	remaining := fmt.Sprintf(loc.TrayRemaining, formatTrayRemaining(s.Remaining))
	if s.EndTime == "" {
		return remaining
	}
	return fmt.Sprintf(loc.TrayTooltip, remaining, fmt.Sprintf(loc.TrayEndTime, s.EndTime))
}

// formatTrayRemaining форматирует длительность как H:MM (или HH:MM при >9ч).
// Отрицательные значения и переработка обрезаются до нуля.
func formatTrayRemaining(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	d = d.Round(time.Minute)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%d:%02d", h, m)
}

// trayBackend — платформенная реализация трея. Методы должны быть
// безопасны для вызова из любой горутины.
type trayBackend interface {
	// Apply обновляет иконку, заголовок и тултип. label может быть пустым.
	Apply(label, tooltip string, dayEnded bool)
	// Quit убирает иконку и завершает событийный цикл трея.
	Quit()
}

// TrayManager связывает модель с платформенным бэкендом трея.
type TrayManager struct {
	backend  trayBackend
	loc      Locale
	lastKey  string
	quitFunc func()
}

// NewTrayManager создаёт менеджер трея с заданным бэкендом. Может вернуть nil,
// если трей недоступен — вызывающий код должен проверить результат.
func NewTrayManager(backend trayBackend, loc Locale) *TrayManager {
	if backend == nil {
		return nil
	}
	return &TrayManager{backend: backend, loc: loc}
}

// SetOnQuit регистрирует функцию, вызываемую при выборе "Выход" в меню трея.
func (t *TrayManager) SetOnQuit(f func()) {
	if t == nil {
		return
	}
	t.quitFunc = f
}

// Update пересчитывает отображение и обновляет трей. Повторные вызовы с тем же
// состоянием игнорируются, чтобы не дёргать DBus зря.
func (t *TrayManager) Update(s TrayState) {
	if t == nil {
		return
	}
	label := s.TrayLabel()
	tooltip := s.TrayTooltip(t.loc)
	key := fmt.Sprintf("%s|%s|%t", label, tooltip, s.DayEnded)
	if key == t.lastKey {
		return
	}
	t.lastKey = key
	t.backend.Apply(label, tooltip, s.DayEnded)
}

// Quit останавливает трей.
func (t *TrayManager) Quit() {
	if t == nil {
		return
	}
	t.backend.Quit()
}

// handleQuit вызывается из меню трея.
func (t *TrayManager) handleQuit() {
	if t == nil || t.quitFunc == nil {
		return
	}
	t.quitFunc()
}