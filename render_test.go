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

	out := m.renderHero()
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

	out := m.renderHero()
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

	out := m.renderHero()
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
	// Статус должен быть после controls (footer), не в середине.
	controlsIdx := strings.Index(rendered, "сохранить")
	statusIdx := strings.Index(rendered, "Загрузка отменена")
	if statusIdx == -1 {
		t.Fatalf("status should appear somewhere: \n%s", rendered)
	}
	if statusIdx < controlsIdx {
		t.Errorf("status must render after footer, not in middle:\n%s", rendered)
	}
}

func TestRenderNarrowNoOverflow(t *testing.T) {
	m := buildModel(t, locAt(10, 0))
	m.startTime.SetValue("09:00")
	m.workTime.SetValue("08:00")
	m.addBreak()
	m.breaks[0].from.SetValue("12:00")
	m.breaks[0].to.SetValue("13:00")
	m.recalculate()
	m.width = 56

	for _, line := range strings.Split(m.renderMain(), "\n") {
		if w := lipgloss.Width(line); w > 56 {
			t.Errorf("line overflow width=%d > 56: %q", w, line)
		}
	}
}
