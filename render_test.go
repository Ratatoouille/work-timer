package main

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
)

// buildModel создаёт модель и форсирует currentTime для детерминированных состояний.
func buildModel(t *testing.T, now time.Time) Model {
	t.Helper()
	m := NewModel("")
	m.currentTime = now
	m.width = 80
	return m
}

func locAt(h, min int) time.Time {
	loc, _ := time.LoadLocation("Europe/Moscow")
	return time.Date(2026, 8, 10, h, min, 0, 0, loc)
}

func TestRenderDoneNoStaleRemaining(t *testing.T) {
	m := buildModel(t, locAt(23, 0)) // после конца смены 09:00+8h=17:00
	m.startTime.SetValue("09:00")
	m.workTime.SetValue("08:00")
	m.recalculate()

	if m.runtimeState() != stateDone {
		t.Fatalf("expected stateDone, got %v", m.runtimeState())
	}

	out := m.renderHero(true)
	if !strings.Contains(out, "Рабочий день завершён") {
		t.Errorf("DONE hero should show completion label, got:\n%s", out)
	}
	if strings.Contains(out, "Осталось") {
		t.Errorf("DONE hero should not show REMAINING label, got:\n%s", out)
	}

	params := m.renderParams()
	// Параметры — конфигурация и поля ввода. Никакого runtime-значения
	// "осталось" быть не должно (оно живёт только в hero).
	if strings.Contains(params, "Осталось") {
		t.Errorf("params must not show runtime remaining, got:\n%s", params)
	}
	tzLine := params[strings.Index(params, "\n")+1:]
	if strings.Contains(tzLine, "Нет") == false {
		t.Errorf("params should contain tz config row, got:\n%s", params)
	}

	// Значение "оставшееся время" (08:00) не должно появляться в DONE-картинке
	// как остаток — только как часть режима/таймлайна начала.
	full := m.renderMain()
	if strings.Contains(full, "Осталось") && strings.Contains(full, "08:00") {
		t.Errorf("DONE screen must not present 08:00 as remaining:\n%s", full)
	}
}

func TestRenderRunning(t *testing.T) {
	m := buildModel(t, locAt(10, 0)) // 09:00+1h, 12% done
	m.startTime.SetValue("09:00")
	m.workTime.SetValue("08:00")
	m.recalculate()

	if m.runtimeState() != stateRunning {
		t.Fatalf("expected stateRunning, got %v", m.runtimeState())
	}

	out := m.renderHero(true)
	if !strings.Contains(out, "Осталось") {
		t.Errorf("running hero should show REMAINING label:\n%s", out)
	}
	if !strings.Contains(out, "Работаем") {
		t.Errorf("running hero should show RUNNING badge:\n%s", out)
	}
}

func TestRenderPaused(t *testing.T) {
	m := buildModel(t, locAt(12, 30)) // в перерыве 12:00-13:00
	m.startTime.SetValue("09:00")
	m.workTime.SetValue("08:00")
	m.addBreak()
	m.breaks[0].from.SetValue("12:00")
	m.breaks[0].to.SetValue("13:00")
	m.recalculate()

	if m.runtimeState() != statePaused {
		t.Fatalf("expected statePaused, got %v", m.runtimeState())
	}

	out := m.renderHero(true)
	if !strings.Contains(out, "Перерыв") {
		t.Errorf("paused hero should show PAUSED badge:\n%s", out)
	}
}

func TestRenderMode2(t *testing.T) {
	m := buildModel(t, locAt(11, 0))
	m.startTime.SetValue("09:00")
	m.worked.SetValue("01:00")
	m.plan.SetValue("08:00")
	m.recalculate()

	if !m.mode2Active() {
		t.Fatalf("expected mode2 active")
	}

	params := m.renderParams()
	if !strings.Contains(params, "2 · Отработано / План") {
		t.Errorf("mode2 should show mode2 descriptor, got:\n%s", params)
	}
}

func TestRenderNoStatusInMiddle(t *testing.T) {
	m := buildModel(t, locAt(10, 0))
	m.startTime.SetValue("09:00")
	m.workTime.SetValue("08:00")
	m.recalculate()
	m.setStatus("Загрузка отменена", StatusError)

	rendered := m.renderMain()
	// Статус должен быть прямо над футером (под разделителем), отделён от
	// статичных подсказок клавиш, а не в середине контента.
	controlsIdx := strings.Index(rendered, "сохранить")
	statusIdx := strings.Index(rendered, "Загрузка отменена")
	// Разделитель, под которым идёт статус, должен идти до строки статуса.
	divIdx := strings.Index(rendered, "──")
	if statusIdx == -1 {
		t.Fatalf("status should appear somewhere: \n%s", rendered)
	}
	if controlsIdx == -1 || statusIdx > controlsIdx {
		t.Errorf("status must render above footer (before 'сохранить')\nstatus=%d controls=%d\n%s", statusIdx, controlsIdx, rendered)
	}
	if divIdx == -1 || statusIdx < divIdx {
		t.Errorf("status must be below the horizontal divider\ndiv=%d status=%d", divIdx, statusIdx)
	}
}

func TestRenderNarrowNoOverflow(t *testing.T) {
	widths := []int{40, 48, 56, 80}
	for _, w := range widths {
		for _, h := range []int{20, 24, 30, 50} {
			m := buildModel(t, locAt(10, 0))
			m.startTime.SetValue("09:00")
			m.workTime.SetValue("08:00")
			m.addBreak()
			m.breaks[0].from.SetValue("12:00")
			m.breaks[0].to.SetValue("13:00")
			m.recalculate()
			m.width = w
			m.height = h

			content := m.View().Content
			for _, line := range strings.Split(content, "\n") {
				if lw := lipgloss.Width(line); lw > w {
					t.Errorf("w=%d h=%d line overflow width=%d: %q", w, h, lw, line)
				}
			}
			if ch := lipgloss.Height(content); ch > h {
				t.Errorf("w=%d h=%d content height %d overflows", w, h, ch)
			}
		}
	}
}

// TestRenderListScreensFitHeight проверяет, что экраны списков не выходят за
// пределы высоты терминала даже при большом числе элементов и курсоре у края.
func TestRenderListScreensFitHeight(t *testing.T) {
	for _, h := range []int{20, 24, 30, 40} {
		m := buildModel(t, locAt(10, 0))
		m.height = h

		m.availableFiles = nil
		for i := 0; i < 30; i++ {
			m.availableFiles = append(m.availableFiles, "file_"+string(rune('A'+i%26))+".json")
		}
		m.mode = ModeFileList
		m.fileListCursor = len(m.availableFiles) - 1
		if ch := lipgloss.Height(m.renderFileList()); ch > h {
			t.Errorf("h=%d fileList overflow: %d", h, ch)
		}

		m.historyEntries = nil
		for i := 0; i < 40; i++ {
			m.historyEntries = append(m.historyEntries, HistoryEntry{Date: "2026-08-10", StartTime: "09:00", EndTime: "17:00"})
		}
		m.mode = ModeHistory
		m.historyCursor = len(m.historyEntries) - 1
		if ch := lipgloss.Height(m.renderHistory()); ch > h {
			t.Errorf("h=%d history overflow: %d", h, ch)
		}

		m.mode = ModePresetList
		m.config.Breaks = nil
		for i := 0; i < 30; i++ {
			m.config.Breaks = append(m.config.Breaks, BreakPreset{Name: "p", From: "12:00", To: "13:00"})
		}
		m.presetCursor = len(m.config.Breaks) - 1
		if ch := lipgloss.Height(m.renderPresetList()); ch > h {
			t.Errorf("h=%d presetList overflow: %d", h, ch)
		}

		m.mode = ModeNormal
		m.helpState = HelpVisible
		if ch := lipgloss.Height(m.renderHelp()); ch > h {
			t.Errorf("h=%d help overflow: %d", h, ch)
		}
	}
}

// TestListWindow проверяет расчёт видимого окна списка.
func TestListWindow(t *testing.T) {
	tests := []struct {
		name                  string
		total, cursor, offset int
		avail                 int
		wantStart, wantEnd    int
	}{
		{"all fit", 5, 2, 0, 10, 0, 5},
		{"cursor at top", 30, 0, 0, 5, 0, 5},
		{"cursor below window", 30, 10, 0, 5, 6, 11},
		{"cursor above offset", 30, 2, 10, 5, 2, 7},
		{"cursor at end", 30, 29, 0, 5, 25, 30},
		{"empty", 0, 0, 0, 5, 0, 0},
		{"avail zero clamps to one", 30, 4, 0, 0, 4, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := listWindow(tt.total, tt.cursor, tt.offset, tt.avail)
			if start != tt.wantStart || end != tt.wantEnd {
				t.Errorf("listWindow(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
					tt.total, tt.cursor, tt.offset, tt.avail, start, end, tt.wantStart, tt.wantEnd)
			}
			if tt.total > 0 && tt.cursor >= start && tt.cursor < end {
				return
			}
			if tt.total == 0 {
				return
			}
			t.Errorf("cursor %d not in window [%d,%d)", tt.cursor, start, end)
		})
	}
}
