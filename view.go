package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"image/color"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Стили инициализируются из конфига через initStyles().
var (
	colorAccent  color.Color
	colorMuted   = lipgloss.Color("8")
	colorSuccess = lipgloss.Color("10")
	colorError   = lipgloss.Color("9")
	colorResult  color.Color
	colorWarn    color.Color

	fieldBoxStyle      lipgloss.Style
	fieldActiveStyle   lipgloss.Style
	containerStyle     lipgloss.Style
	headerStyle        lipgloss.Style
	modeNormalStyle    lipgloss.Style
	modeInsertStyle    lipgloss.Style
	headerDividerStyle lipgloss.Style
	statusBarStyle     lipgloss.Style
	fileNameStyle      lipgloss.Style
	dirtyDotStyle      lipgloss.Style
	statusDotStyle     lipgloss.Style
	clockStyle         lipgloss.Style

	sectionBreakHeaderStyle lipgloss.Style
	sectionDividerStyle     lipgloss.Style

	heroLabelStyle  lipgloss.Style
	heroValueStyle  lipgloss.Style
	timerStateStyle lipgloss.Style

	paramLabelStyle lipgloss.Style
	paramValueStyle lipgloss.Style
	breakTimeStyle  lipgloss.Style
	breakLineStyle  lipgloss.Style
	breakDurStyle   lipgloss.Style
	offStyle        lipgloss.Style

	resultValueStyle lipgloss.Style
	errorStyle       lipgloss.Style

	statusSuccessStyle lipgloss.Style
	statusErrorStyle   lipgloss.Style
	statusWarnStyle    lipgloss.Style

	controlsBarStyle lipgloss.Style
	controlKeyStyle  lipgloss.Style

	promptStyle             lipgloss.Style
	fileListItemStyle       lipgloss.Style
	fileListItemActiveStyle lipgloss.Style
	helpBoxStyle            lipgloss.Style
)

// initStyles вызывается из NewModel после загрузки конфига.
func initStyles(cfg Config) {
	colorAccent = lipgloss.Color(cfg.UI.Colors.Accent)
	colorResult = lipgloss.Color(cfg.UI.Colors.Result)
	colorWarn = lipgloss.Color(cfg.UI.Colors.Warn)

	fieldBoxStyle = lipgloss.NewStyle().Padding(0, 1)
	fieldActiveStyle = fieldBoxStyle.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(0, 1)

	containerStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorMuted).
		Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorAccent)

	modeNormalStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorSuccess)

	modeInsertStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("0")).
		Background(colorWarn).
		Padding(0, 1)

	headerDividerStyle = lipgloss.NewStyle().Foreground(colorMuted)
	statusBarStyle = lipgloss.NewStyle().Foreground(colorMuted)
	clockStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))

	fileNameStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Bold(true)

	dirtyDotStyle = lipgloss.NewStyle().Foreground(colorWarn)
	statusDotStyle = lipgloss.NewStyle().Foreground(colorSuccess)

	sectionBreakHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15"))

	sectionDividerStyle = lipgloss.NewStyle().Foreground(colorMuted)

	// Крупный акцентный блок "оставшееся время"
	heroLabelStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7"))
	heroValueStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorAccent)
	timerStateStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorAccent)

	// Вторичные label/value пары
	paramLabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("7"))
	paramValueStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15"))

	breakTimeStyle = lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true)
	breakLineStyle = lipgloss.NewStyle().Foreground(colorMuted)
	breakDurStyle = lipgloss.NewStyle().Foreground(colorMuted)
	offStyle = lipgloss.NewStyle().Foreground(colorMuted)

	resultValueStyle = lipgloss.NewStyle().Bold(true).Foreground(colorResult)

	errorStyle = lipgloss.NewStyle().Bold(true).Foreground(colorError)

	statusSuccessStyle = lipgloss.NewStyle().Foreground(colorSuccess)
	statusErrorStyle = lipgloss.NewStyle().Foreground(colorError)
	statusWarnStyle = lipgloss.NewStyle().Foreground(colorWarn)

	controlsBarStyle = lipgloss.NewStyle().
		Foreground(colorMuted).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorMuted).
		MarginTop(1).
		PaddingTop(1)

	controlKeyStyle = lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true)

	promptStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		BorderForeground(colorAccent)

	fileListItemStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("7"))

	fileListItemActiveStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Background(lipgloss.Color("4")).
		Foreground(lipgloss.Color("15")).
		Bold(true)

	helpBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorMuted).
		Padding(1, 2)
}

// containerWidth возвращает ширину основного блока в зависимости от терминала.
func (m Model) containerWidth() int {
	w := m.width - 4
	if w < 52 {
		return 52
	}
	if w > 92 {
		return 92
	}
	return w
}

func (m Model) dividerWidth() int {
	w := m.containerWidth() - 8
	if w < 40 {
		return 40
	}
	return w
}

// labelCol — ширина колонки label/value пар. Равна ширине самого длинного
// видимого label (чтобы пары выравнивались и длинные подписи не переносились),
// но ограничена шириной контейнера, оставляя место для значения.
func (m Model) labelCol(labels ...string) int {
	widest := 0
	maxAvail := max(m.containerWidth()-14, 8)
	for _, l := range labels {
		if w := lipgloss.Width(l); w > widest {
			widest = w
		}
	}
	if widest > maxAvail {
		widest = maxAvail
	}
	return widest
}

func (m Model) View() tea.View {
	var content string
	switch {
	case m.helpState == HelpVisible:
		content = m.renderHelp()
	case m.mode == ModeSavePrompt:
		content = m.renderSavePrompt()
	case m.mode == ModeLoadPrompt:
		content = m.renderLoadPrompt()
	case m.mode == ModeFileList:
		content = m.renderFileList()
	case m.mode == ModePresetList:
		content = m.renderPresetList()
	case m.mode == ModeHistory:
		content = m.renderHistory()
	default:
		content = containerStyle.Width(m.containerWidth()).Render(m.renderMain())
	}

	if m.width == 0 || m.height == 0 {
		v := tea.NewView(content)
		v.AltScreen = true
		return v
	}
	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}

func (m Model) renderMain() string {
	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteString(m.renderHero())
	b.WriteString(m.renderTimerRow())
	b.WriteString(m.renderBreaks())
	b.WriteString(m.renderParams())
	b.WriteString(m.renderStatusMessage())
	b.WriteString(m.renderResult())
	b.WriteString(m.renderControls())
	return b.String()
}

func (m Model) renderHeader() string {
	modeStyle := modeNormalStyle
	modeStr := m.locale.ModeNormal
	if m.mode == ModeInsert {
		modeStyle = modeInsertStyle
		modeStr = m.locale.ModeInsert
	}

	div := headerDividerStyle.Render("  │  ")

	var filePart string
	if m.saveFile != "" {
		dot := statusDotStyle.Render("●")
		if m.isDirty {
			dot = dirtyDotStyle.Render("●")
		}
		filePart = dot + "  " + fileNameStyle.Render(filepath.Base(m.saveFile))
	} else {
		filePart = statusBarStyle.Render(m.locale.NoFileSelected)
	}

	clock := ""
	if !m.currentTime.IsZero() {
		clock = clockStyle.Render(m.currentTime.Format("15:04"))
	}

	parts := []string{headerStyle.Render("◉ WORK TIMER"), " ", modeStyle.Render(modeStr), div, filePart}
	if clock != "" {
		parts = append(parts, div, clock)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, parts...) + "\n"
}

// renderSectionHeader рисует разделитель секции с заголовком.
// "▾" обозначает раскрытую секцию (в отличие от "▸" — маркера выбора).
func (m Model) renderSectionHeader(title string, style lipgloss.Style) string {
	text := style.Render("▾ " + title + " ")
	lineLen := max(m.dividerWidth()-lipgloss.Width(text), 2)
	line := sectionDividerStyle.Render(strings.Repeat("─", lineLen))
	return "\n" + lipgloss.JoinHorizontal(lipgloss.Left, text, line) + "\n"
}

// runtimeState возвращает состояние таймера на основе фактического времени
// относительно рабочего окна. Определяется из существующих данных (start,
// endTimeRaw, breaks), без новой бизнес-логики.
func (m Model) runtimeState() timerState {
	percent, _, remaining, ok := m.progressInfo()

	switch {
	case !ok:
		if m.hasTimerInput() {
			return stateReady
		}
		return stateIdle
	case percent >= 1.0 && remaining <= 0:
		return stateDone
	case percent >= 1.0:
		return stateOvertime
	case m.onBreak():
		return statePaused
	default:
		return stateRunning
	}
}

// hasTimerInput — есть ли какие-то введённые данные таймера (не пустой экран).
func (m Model) hasTimerInput() bool {
	return m.startTime.Value() != "" || m.workTime.Value() != "" ||
		m.worked.Value() != "" || m.plan.Value() != "" || len(m.breaks) > 0
}

// renderHero — крупный блок "оставшееся время" + прогресс.
// value = оставшееся время, bar = доля прошедшего времени рабочего окна.
// Статус выводится отдельным бейджем, чтобы не возникало противоречия
// "ОСТАЛОСЬ 09:00" рядом с "ГОТОВО".
func (m Model) renderHero() string {
	percent, _, _, ok := m.progressInfo()

	value := m.heroRemaining()
	badge := m.statusBadge()

	line := value
	if badge != "" {
		line = lipgloss.JoinHorizontal(lipgloss.Center, value, badge)
	}

	bar := m.renderEmptyProgressBar()
	if ok {
		// Подпись шкалы: Начало ... Окончание по краям прогресс-бара.
		bar = m.renderProgressTimeline(percent)
	}

	return "\n" + heroLabelStyle.Render(m.locale.RemainingLabel+":") + "\n\n" +
		line + "\n\n" +
		bar + "\n\n"
}

// heroRemaining возвращает главное значение оставшегося времени.
// Когда рабочий день завершён — показывает "00:00", иначе вычисленное осталось.
func (m Model) heroRemaining() string {
	if m.runtimeState() == stateDone {
		return heroValueStyle.Render("00:00")
	}
	if secs := m.resultSeconds(); secs > 0 {
		t := time.Duration(secs) * time.Second
		return heroValueStyle.Render(fmt.Sprintf("%02d:%02d", int(t.Hours()), int(t.Minutes())%60))
	}
	return heroValueStyle.Render("—:—")
}

// statusBadge возвращает бейдж состояния таймера (цветной текст) или пустую строку.
func (m Model) statusBadge() string {
	st := m.runtimeState()
	var s string
	switch st {
	case stateRunning:
		s = timerStateStyle.Render(m.locale.StateRunning)
	case statePaused:
		s = statusWarnStyle.Render(m.locale.StatePaused)
	case stateDone:
		s = statusSuccessStyle.Render(m.locale.StateDone)
	case stateOvertime:
		s = statusErrorStyle.Render(m.locale.StateOvertime)
	case stateReady:
		s = timerStateStyle.Render(m.locale.StateRunning)
	default:
		return ""
	}
	return "  " + s
}

type timerState int

const (
	stateIdle timerState = iota
	stateReady
	stateRunning
	statePaused
	stateDone
	stateOvertime
)

// onBreak возвращает true, если текущее время попадает в один из перерывов.
func (m Model) onBreak() bool {
	if m.endTimeRaw == "" || m.currentTime.IsZero() {
		return false
	}
	if _, _, _, ok := m.progressInfo(); !ok {
		return false
	}
	nowStr := m.currentTime.Format("15:04")
	for _, br := range m.getBreaksData() {
		if br.From == "" || br.To == "" {
			continue
		}
		if nowStr >= br.From && nowStr < br.To {
			return true
		}
	}
	return false
}

// resultSeconds возвращает оставшееся время (в секундах) из активного режима.
func (m Model) resultSeconds() float64 {
	switch {
	case m.worked.Value() != "" && m.plan.Value() != "":
		w, err1 := m.calculator.ParseDuration(m.worked.Value())
		pl, err2 := m.calculator.ParseDuration(m.plan.Value())
		if err1 != nil || err2 != nil {
			return 0
		}
		return (pl - w).Seconds()
	case m.workTime.Value() != "":
		if d, err := m.calculator.ParseDuration(m.workTime.Value()); err == nil {
			return d.Seconds()
		}
	}
	return 0
}

// renderTimerRow — время начала и окончания рабочего дня.
func (m Model) renderTimerRow() string {
	end := m.formatResultEnd()
	col := m.labelCol(m.locale.FieldStart, m.locale.FieldEnd)
	sep := "   "
	startLabel := m.valueLabelStyle(m.cursor == FieldStartTime)

	// В Insert-режиме на активном поле показываем настоящий textinput,
	// иначе — значение с индикацией фокуса (подчёркивание/placeholder).
	var start string
	if m.mode == ModeInsert && m.cursor == FieldStartTime {
		start = fieldActiveStyle.Render(m.startTime.View())
	} else if m.startTime.Value() != "" {
		start = paramValueStyle.Render(m.startTime.Value())
		if m.cursor == FieldStartTime {
			start = paramValueStyle.Underline(true).Foreground(colorAccent).Render(m.startTime.Value())
		}
	} else {
		start = statusBarStyle.Render("—:—")
		if m.cursor == FieldStartTime {
			start = offStyle.Render("▍" + m.locale.PlaceholderTime)
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Left,
		startLabel.Width(col).Render(m.locale.FieldStart),
		"  ",
		start,
		sep,
		paramLabelStyle.Width(col).Render(m.locale.FieldEnd),
		"  ",
		end,
	)

	return row + "\n\n"
}

// renderBreaks — компактные timeline-строки перерывов.
func (m Model) renderBreaks() string {
	if len(m.breaks) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(m.renderSectionHeader(m.locale.SectionBreaks, sectionBreakHeaderStyle))
	for i, br := range m.breaks {
		baseIndex := FieldBreaksStart + i*2
		b.WriteString(m.renderBreakRow(i, baseIndex, br) + "\n")
	}
	return b.String()
}

func (m Model) renderBreakRow(idx, baseIndex int, br Break) string {
	from := br.from.Value()
	to := br.to.Value()

	focusedFrom := m.cursor == baseIndex
	focusedTo := m.cursor == baseIndex+1

	// В Insert-режиме фокусируемое поле перерыва показываем как textinput.
	if m.mode == ModeInsert && (focusedFrom || focusedTo) {
		if focusedFrom {
			return "  " + fieldActiveStyle.Render(br.from.View()) + "  " + offStyle.Render("—:—")
		}
		return "  " + offStyle.Render("—:—") + "  " + fieldActiveStyle.Render(br.to.View())
	}

	dispFrom := offStyle.Render("—:—")
	if from != "" {
		st := breakTimeStyle
		if focusedFrom {
			st = st.Bold(true).Underline(true)
		}
		dispFrom = st.Render(from)
	} else if focusedFrom {
		dispFrom = offStyle.Render("▍" + m.locale.PlaceholderTime)
	}
	dispTo := offStyle.Render("—:—")
	if to != "" {
		st := breakTimeStyle
		if focusedTo {
			st = st.Bold(true).Underline(true)
		}
		dispTo = st.Render(to)
	} else if focusedTo {
		dispTo = offStyle.Render("▍" + m.locale.PlaceholderTime)
	}

	sep := 6
	line := breakLineStyle.Render(strings.Repeat("─", sep))

	row := lipgloss.JoinHorizontal(lipgloss.Left,
		dispFrom,
		"  "+line+"  ",
		dispTo,
	)

	if from != "" && to != "" {
		if d := m.calculator.BreaksDuration([]BreakTime{{From: from, To: to}}); d > 0 {
			row = lipgloss.JoinHorizontal(lipgloss.Left, row, "   ", breakDurStyle.Render(formatShortDuration(d, m.locale)))
		}
	}
	return "  " + row
}

func (m Model) renderParams() string {
	var b strings.Builder
	b.WriteString(m.renderSectionHeader(m.locale.SectionParams, sectionBreakHeaderStyle))

	col := m.labelCol(
		m.locale.StatMode,
		m.locale.FieldRemainingTime,
		m.locale.FieldWorked,
		m.locale.FieldPlan,
		checkboxLabel(m),
	)

	// Режим: показываем номер и описание как состояние, а не "кнопку".
	var modeStr string
	switch {
	case m.mode2Active():
		modeStr = paramValueStyle.Render("2 · " + m.locale.FieldWorked + " / " + m.locale.FieldPlan)
	case m.mode1Active():
		modeStr = paramValueStyle.Render("1 · " + m.locale.FieldRemainingTime)
	default:
		modeStr = offStyle.Render("—")
	}
	b.WriteString(m.renderPairAt(m.locale.StatMode, modeStr, col) + "\n")

	// Поля редактируются независимо от активного режима (навигация и ввод не
	// должны ломаться), но неактивная группа визуально приглушается.
	timeField := m.renderValueFieldAt(FieldWorkTime, m.locale.FieldRemainingTime, m.workTime, false, col)
	workedField := m.renderValueFieldAt(FieldWorked, m.locale.FieldWorked, m.worked, true, col)
	planField := m.renderValueFieldAt(FieldPlan, m.locale.FieldPlan, m.plan, true, col)

	if m.mode2Active() {
		timeField = offStyle.Render(timeField)
	}
	if m.mode1Active() {
		workedField = offStyle.Render(workedField)
		planField = offStyle.Render(planField)
	}

	b.WriteString(timeField + "\n")
	b.WriteString(workedField + "\n")
	b.WriteString(planField + "\n")

	tzVal := offStyle.Render("Нет")
	if m.addTZ {
		tzVal = statusSuccessStyle.Render("Да")
	}
	tzLabel := m.valueLabelStyle(m.cursor == FieldAddTZ)
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, tzLabel.Width(col).Render(checkboxLabel(m)), "  ", tzVal) + "\n")

	return b.String()
}

func checkboxLabel(m Model) string {
	if m.config.Timezone != "" {
		if label := TimezoneLabel(m.config.Timezone); label != "" {
			return fmt.Sprintf(m.locale.CheckboxShowIn, label)
		}
	}
	return m.locale.CheckboxAddTZ
}

func (m Model) mode1Active() bool { return m.workTime.Value() != "" }
func (m Model) mode2Active() bool {
	return m.workTime.Value() == "" && m.worked.Value() != "" && m.plan.Value() != ""
}

func (m Model) renderPairAt(label, value string, col int) string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		paramLabelStyle.Width(col).Render(label),
		"  ",
		value,
	)
}

// renderValueFieldAt рисует значение поля (без бокса и символа ">").
// В Insert-режиме на активном поле показывает настоящий textinput для ввода.
func (m Model) valueLabelStyle(focused bool) lipgloss.Style {
	s := paramLabelStyle
	if focused {
		s = s.Bold(true).Foreground(colorAccent)
	}
	return s
}

func (m Model) renderValueFieldAt(index int, label string, input textinput.Model, isDuration bool, col int) string {
	focused := m.cursor == index
	value := input.Value()
	labelStyle := m.valueLabelStyle(focused)

	if m.mode == ModeInsert && focused {
		return lipgloss.JoinHorizontal(lipgloss.Left,
			labelStyle.Width(col).Render(label),
			"  ",
			fieldActiveStyle.Render(input.View()),
		)
	}

	if value == "" {
		val := offStyle.Render("—")
		if focused {
			val = offStyle.Render("▍" + m.locale.PlaceholderTime)
		}
		return lipgloss.JoinHorizontal(lipgloss.Left,
			labelStyle.Width(col).Render(label),
			"  ",
			val,
		)
	}

	invalid := isInvalidTimeValue(value)
	if isDuration {
		invalid = isInvalidDurationValue(value)
	}
	valStyle := paramValueStyle
	if focused {
		valStyle = paramValueStyle.Underline(true).Foreground(colorAccent)
	}
	if invalid {
		valStyle = statusErrorStyle
	}

	return lipgloss.JoinHorizontal(lipgloss.Left,
		labelStyle.Width(col).Render(label),
		"  ",
		valStyle.Render(value),
	)
}

func (m Model) renderStatusMessage() string {
	if m.statusMessage == "" {
		return ""
	}

	var s string
	switch m.statusType {
	case StatusSuccess:
		s = statusSuccessStyle.Render(m.statusMessage)
	case StatusError:
		s = statusErrorStyle.Render(m.statusMessage)
	case StatusWarn:
		s = statusWarnStyle.Render(m.statusMessage)
	default:
		s = statusBarStyle.Render(m.statusMessage)
	}

	return "\n" + s + "\n"
}

func (m Model) renderResult() string {
	if m.result == "" && m.err == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString(m.renderSectionHeader(m.locale.SectionResult, resultValueStyle))

	if m.err != "" {
		b.WriteString("  " + errorStyle.Render("✗  "+m.err) + "\n")
		return b.String()
	}

	_, elapsed, remaining, ok := m.progressInfo()
	col := m.labelCol(m.locale.FieldWorked, m.locale.Remaining)

	// Итоговая сводка: отработано и осталось (или окончено). Прогресс-бар не
	// дублируем — он уже показан в верхнем блоке.
	startStr := "  " + m.renderSummaryRow(m.locale.FieldWorked, formatDuration(elapsed), col)
	b.WriteString(startStr + "\n")

	st := m.runtimeState()

	if ok && st != stateDone && st != stateOvertime {
		b.WriteString("  " + m.renderSummaryRow(m.locale.RemainingCap, formatDuration(remaining), col) + "\n")
	} else {
		var statusStr string
		switch st {
		case stateDone:
			statusStr = statusSuccessStyle.Render(m.locale.StateDone)
		case stateOvertime:
			statusStr = statusErrorStyle.Render(m.locale.StateOvertime)
		case statePaused:
			statusStr = statusWarnStyle.Render(m.locale.StatePaused)
		default:
			statusStr = timerStateStyle.Render(m.locale.StateRunning)
		}
		b.WriteString("  " + m.renderSummaryRow(m.locale.Status, statusStr, col) + "\n")
	}

	return b.String()
}

// renderSummaryRow — строка label/value с общим выравниванием колонки.
func (m Model) renderSummaryRow(label, value string, col int) string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		paramLabelStyle.Width(col).Render(label),
		"  ",
		value,
	)
}

// formatResultEnd возвращает время окончания с подписью часового пояса (UTC+07).
func (m Model) formatResultEnd() string {
	endStr := strings.TrimSpace(m.result)
	if endStr == "" {
		return statusBarStyle.Render("—:—")
	}
	parts := strings.SplitN(endStr, " ", 2)
	timePart := parts[0]
	out := paramValueStyle.Render(timePart)
	if len(parts) > 1 && m.config.Timezone != "" {
		if off := utcOffsetLabel(m.config.Timezone); off != "" {
			out = out + " " + statusBarStyle.Render("UTC"+off)
		}
	}
	return out
}

// utcOffsetLabel возвращает смещение зоны вида "+07" или "+05:30".
func utcOffsetLabel(tz string) string {
	loc := TimezoneLocation(tz)
	if loc == nil {
		return ""
	}
	_, offset := time.Now().In(loc).Zone()
	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}
	h := offset / 3600
	m := (offset % 3600) / 60
	if m == 0 {
		return fmt.Sprintf("%s%02d", sign, h)
	}
	return fmt.Sprintf("%s%02d:%02d", sign, h, m)
}

// renderProgressBar рисует текстовый прогресс-бар с процентом на той же строке.
// avail (если >0) ограничивает ширину полосы; иначе считается из контейнера.
func (m Model) renderProgressBar(percent float64, avail int) string {
	barWidth := min(max(m.containerWidth()-14, 14), 50)
	if avail > 0 {
		barWidth = min(avail, barWidth)
	}
	if barWidth > m.dividerWidth() {
		barWidth = m.dividerWidth()
	}
	barWidth = max(barWidth, 4)
	filled := int(percent * float64(barWidth))
	filled = min(filled, barWidth)
	filled = max(filled, 0)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	pct := fmt.Sprintf("%3.0f%%", percent*100)
	return lipgloss.NewStyle().Foreground(colorAccent).Render(bar) + " " + statusBarStyle.Render(pct)
}

// renderProgressTimeline — прогресс-бар как шкала рабочего дня: слева время
// начала, справа время окончания, под ним заполненная полоса и проценты.
func (m Model) renderProgressTimeline(percent float64) string {
	startStr := "—:—"
	if s := m.startTime.Value(); s != "" {
		startStr = s
	}
	endStr := m.formatResultEndPlain()

	// Реальная доступная ширина с учётом времени начала/окончания по краям.
	avail := max(m.dividerWidth()-lipgloss.Width(startStr)-lipgloss.Width(endStr)-4, 10)
	bar := m.renderProgressBar(percent, avail)

	return lipgloss.JoinHorizontal(lipgloss.Left,
		statusBarStyle.Render(startStr),
		"  ",
		bar,
		"  ",
		statusBarStyle.Render(endStr),
	)
}

// formatResultEndPlain возвращает только время окончания без timezone-суффикса.
func (m Model) formatResultEndPlain() string {
	endStr := strings.TrimSpace(m.result)
	if endStr == "" {
		return "—:—"
	}
	return strings.SplitN(endStr, " ", 2)[0]
}

func (m Model) renderEmptyProgressBar() string {
	barWidth := min(max(m.containerWidth()-14, 14), 50)
	if barWidth > m.dividerWidth() {
		barWidth = m.dividerWidth()
	}
	bar := strings.Repeat("░", barWidth)
	return lipgloss.NewStyle().Foreground(colorMuted).Render(bar)
}

func (m Model) progressInfo() (percent float64, elapsed, remaining time.Duration, ok bool) {
	if m.endTimeRaw == "" {
		return 0, 0, 0, false
	}

	now := m.currentTime
	if now.IsZero() {
		now = time.Now()
	}

	endTimeStr := strings.TrimSpace(m.endTimeRaw)

	parts := strings.Split(endTimeStr, " ")
	if len(parts) == 0 {
		return 0, 0, 0, false
	}

	timePart := parts[0]
	timeParts := strings.Split(timePart, ":")
	if len(timeParts) != 2 {
		return 0, 0, 0, false
	}

	endHour, err := strconv.Atoi(timeParts[0])
	if err != nil {
		return 0, 0, 0, false
	}
	endMin, err := strconv.Atoi(timeParts[1])
	if err != nil {
		return 0, 0, 0, false
	}

	// endTimeRaw задаётся в input timezone (без конвертации TZ),
	// поэтому окно прогресса считаем в той же зоне.
	endLoc := TimezoneLocation(m.config.InputTimezone)
	if endLoc == nil {
		endLoc = time.Local
	}

	startTimeStr := m.startTime.Value()
	if startTimeStr == "" {
		return 0, 0, 0, false
	}

	startParts := strings.Split(startTimeStr, ":")
	if len(startParts) != 2 {
		return 0, 0, 0, false
	}

	startHour, err := strconv.Atoi(startParts[0])
	if err != nil {
		return 0, 0, 0, false
	}
	startMin, err := strconv.Atoi(startParts[1])
	if err != nil {
		return 0, 0, 0, false
	}

	startLoc := TimezoneLocation(m.config.InputTimezone)
	if startLoc == nil {
		startLoc = time.Local
	}

	nowIn := now.In(endLoc)

	// Строим end на today, start на today.
	endToday := time.Date(nowIn.Year(), nowIn.Month(), nowIn.Day(), endHour, endMin, 0, 0, endLoc)
	startToday := time.Date(nowIn.Year(), nowIn.Month(), nowIn.Day(), startHour, startMin, 0, 0, startLoc).In(endLoc)

	// Если end <= start — ночная смена, end на завтра.
	if !endToday.After(startToday) {
		endToday = endToday.AddDate(0, 0, 1)
	}

	var bestStart, bestEnd time.Time

	switch {
	case !nowIn.Before(startToday) && !nowIn.After(endToday):
		// now внутри окна
		bestStart = startToday
		bestEnd = endToday
	case nowIn.After(endToday):
		// now после конца — рабочий день окончен
		bestStart = startToday
		bestEnd = endToday
	default:
		// now до начала — день ещё не начался.
		// Проверим вчерашнее окно (если работа началась вчера).
		startYesterday := startToday.AddDate(0, 0, -1)
		endYesterday := endToday.AddDate(0, 0, -1)
		if !nowIn.Before(startYesterday) && !nowIn.After(endYesterday) {
			bestStart = startYesterday
			bestEnd = endYesterday
		} else {
			return 0, 0, 0, false
		}
	}

	totalDuration := bestEnd.Sub(bestStart)
	elapsedDuration := nowIn.Sub(bestStart)

	breaksDur := m.getBreaksDuration()
	elapsedDuration = max(elapsedDuration-breaksDur, 0)

	if totalDuration <= 0 {
		return 0, 0, 0, false
	}

	p := float64(elapsedDuration) / float64(totalDuration)
	if p > 1.0 {
		p = 1.0
	}
	if p < 0 {
		p = 0
	}

	remainingDuration := max(totalDuration-elapsedDuration, 0)

	return p, elapsedDuration, remainingDuration, true
}

func (m Model) getBreaksDuration() time.Duration {
	return m.calculator.BreaksDuration(m.getBreaksData())
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dч %02dм", hours, minutes)
}

// formatShortDuration — компактное представление длительности (для перерывов).
func formatShortDuration(d time.Duration, loc Locale) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf(loc.DurFormatHours, hours, minutes)
	}
	return fmt.Sprintf(loc.DurFormatMins, minutes)
}

func (m Model) renderControls() string {
	k := func(s string) string { return controlKeyStyle.Render(s) }
	d := func(s string) string { return statusBarStyle.Render(s) }
	dot := headerDividerStyle.Render(" · ")
	divW := m.dividerWidth()

	if m.mode == ModeInsert {
		return controlsBarStyle.Width(divW).Render(k("esc") + "  " + d(m.locale.CtrlEdit))
	}

	line := strings.Join([]string{
		k("j/k") + " " + d(m.locale.CtrlNav),
		k("i") + " " + d(m.locale.CtrlEdit),
		k("s") + " " + d(m.locale.CtrlSave),
		k("o") + " " + d(m.locale.CtrlOpen),
		k("?") + " " + d(m.locale.CtrlHelp),
		k("q") + " " + d(m.locale.CtrlQuit),
	}, dot)

	return controlsBarStyle.Width(divW).Render(line)
}

func (m Model) renderSavePrompt() string {
	body := headerStyle.Render(m.locale.SaveTitle) +
		"\n\n" + statusBarStyle.Render(fmt.Sprintf(m.locale.SaveFolder, m.workDir)) +
		"\n\n" + m.filePathInput.View()

	if m.statusMessage != "" {
		if m.statusType == StatusSuccess {
			body += "\n\n" + statusSuccessStyle.Render(m.statusMessage)
		} else {
			body += "\n\n" + statusErrorStyle.Render(m.statusMessage)
		}
	}
	body += "\n\n" + statusBarStyle.Render(m.locale.SaveHint)
	return promptStyle.Render(body)
}

func (m Model) renderLoadPrompt() string {
	body := headerStyle.Render(m.locale.LoadTitle) +
		"\n\n" + m.filePathInput.View()

	if m.statusMessage != "" {
		body += "\n\n" + statusBarStyle.Render(m.statusMessage)
	}
	body += "\n\n" + statusBarStyle.Render(m.locale.LoadHint)
	return promptStyle.Render(body)
}

func (m Model) renderFileList() string {
	var b strings.Builder

	// Единый заголовок (в стиле заголовка приложения), без декоративного emoji.
	b.WriteString(headerStyle.Render(m.locale.FileListTitle) + "\n\n")

	// Путь — приглушённый, не доминирует.
	b.WriteString(statusBarStyle.Render(fmt.Sprintf(m.locale.SaveFolder, m.workDir)) + "\n\n")

	// Поиск — обычное TUI-поле (активируется клавишей "/").
	var searchInput string
	if m.fileSearchInput.Focused() {
		searchInput = fieldActiveStyle.Render(m.fileSearchInput.View())
	} else {
		searchInput = paramValueStyle.Render(m.fileSearchInput.View())
	}
	b.WriteString(paramLabelStyle.Render(m.locale.FileSearchLabel+":") + " " + searchInput + "\n\n")

	if len(m.availableFiles) == 0 {
		b.WriteString("\n  " + statusBarStyle.Render(m.locale.FileListEmpty) + "\n\n")
		b.WriteString(statusBarStyle.Render(m.locale.FileListHintNew))
		return promptStyle.Render(b.String())
	}

	// Список файлов. Выбранный — с маркером "▸" и highlight.
	for i, file := range m.availableFiles {
		if i == m.fileListCursor {
			b.WriteString(fileListItemActiveStyle.Render("▸ "+file) + "\n")
		} else {
			b.WriteString(fileListItemStyle.Render("  "+file) + "\n")
		}
	}

	b.WriteString("\n")

	// Состояние: переименование / подтверждение удаления.
	if m.renaming {
		b.WriteString(statusBarStyle.Render(m.locale.RenamePrompt+" ") + m.renameInput.View() + "\n")
		b.WriteString(statusBarStyle.Render("[Enter] rename  [Esc] cancel") + "\n")
	} else if m.confirmDelete && m.statusMessage != "" {
		b.WriteString(statusWarnStyle.Render(m.statusMessage) + "\n")
	}

	// Footer сгруппирован в две строки, клавиши выделены.
	fk := func(s string) string { return controlKeyStyle.Render(s) }
	fd := func(s string) string { return statusBarStyle.Render(s) }
	b.WriteString("\n" + fk("j/k") + " " + fd(m.locale.CtrlNav) + "   " + fk("Enter") + " " + fd(m.locale.CtrlOpen) +
		"   " + fk("/") + " " + fd(m.locale.CtrlSearch) + "\n")
	b.WriteString(fk("d") + " " + fd(m.locale.CtrlDelete) + "   " + fk("r") + " " + fd(m.locale.CtrlRename) +
		"   " + fk("n") + " " + fd(m.locale.CtrlNew) + "   " + fk("Esc") + " " + fd(m.locale.CtrlCancel) + "\n")

	return promptStyle.Render(b.String())
}

func (m Model) renderPresetList() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(m.locale.PresetTitle) + "\n\n")

	if len(m.config.Breaks) == 0 {
		b.WriteString("\n  " + statusBarStyle.Render(m.locale.PresetEmpty) + "\n\n")
	} else {
		for i, preset := range m.config.Breaks {
			label := fmt.Sprintf("%s  (%s – %s)", preset.Name, preset.From, preset.To)
			if i == m.presetCursor {
				b.WriteString(fileListItemActiveStyle.Render("▸ "+label) + "\n")
			} else {
				b.WriteString(fileListItemStyle.Render("  "+label) + "\n")
			}
		}
		b.WriteString("\n")
	}

	divider := sectionDividerStyle.Render(strings.Repeat("─", 40))
	b.WriteString(divider + "\n\n")
	b.WriteString(statusBarStyle.Render(m.locale.PresetHint))
	return promptStyle.Render(b.String())
}

func (m Model) renderHistory() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(m.locale.HistoryTitle) + "\n\n")

	if len(m.historyEntries) == 0 {
		b.WriteString("\n  " + statusBarStyle.Render(m.locale.HistoryEmpty) + "\n\n")
		b.WriteString(m.historyFooter())
		return promptStyle.Render(b.String())
	}

	// Список записей в том же стиле, что и список файлов: маркер "▸" у
	// выбранной записи, фиксированные колонки.
	for i, e := range m.historyEntries {
		startStr := e.StartTime
		if startStr == "" {
			startStr = "—:—"
		}
		endStr := e.EndTime
		if endStr == "" {
			endStr = "—:—"
		}
		breaksInfo := e.BreaksDur
		if e.Breaks > 0 {
			breaksInfo = fmt.Sprintf("%d × %s", e.Breaks, e.BreaksDur)
		}
		line := fmt.Sprintf("%-12s  %-6s  %-8s  %-12s",
			e.Date, startStr, endStr, breaksInfo)
		if i == m.historyCursor {
			b.WriteString(fileListItemActiveStyle.Render("▸ "+line) + "\n")
		} else {
			b.WriteString(fileListItemStyle.Render("  "+line) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.historyFooter())

	return promptStyle.Render(b.String())
}

// historyFooter — двухстрочный footer в стиле file picker.
func (m Model) historyFooter() string {
	fk := func(s string) string { return controlKeyStyle.Render(s) }
	fd := func(s string) string { return statusBarStyle.Render(s) }
	return "\n" +
		fk("j/k") + " " + fd(m.locale.CtrlNav) + "   " + fk("Enter") + " " + fd(m.locale.CtrlOpen) + "\n" +
		fk("Esc") + " " + fd(m.locale.CtrlCancel) + "\n"
}

func (m Model) renderHelp() string {
	return helpBoxStyle.Render(fmt.Sprintf(m.locale.HelpText, ConfigPath, DefaultWorkDir))
}

// isInvalidTimeValue возвращает true если строка непустая и не является
// корректным или частично набранным временем: [H]H:[M]M.
func isInvalidTimeValue(s string) bool {
	if s == "" {
		return false
	}
	colonIdx := strings.Index(s, ":")
	switch colonIdx {
	case -1:
		if len(s) > 2 {
			return true
		}
		for _, ch := range s {
			if ch < '0' || ch > '9' {
				return true
			}
		}
		h := 0
		for _, ch := range s {
			h = h*10 + int(ch-'0')
		}
		return h > 23
	case 1:
		if s[0] < '0' || s[0] > '9' {
			return true
		}
		return invalidMinutesSuffix(s[2:])
	case 2:
		if s[0] < '0' || s[0] > '9' || s[1] < '0' || s[1] > '9' {
			return true
		}
		if (int(s[0]-'0'))*10+int(s[1]-'0') > 23 {
			return true
		}
		return invalidMinutesSuffix(s[3:])
	default:
		return true
	}
}

// invalidMinutesSuffix проверяет суффикс после двоеточия ("", "M", "MM").
func invalidMinutesSuffix(s string) bool {
	switch len(s) {
	case 0:
		return false
	case 1:
		return s[0] < '0' || s[0] > '5'
	case 2:
		return s[0] < '0' || s[0] > '5' || s[1] < '0' || s[1] > '9'
	default:
		return true
	}
}

// isInvalidDurationValue — как isInvalidTimeValue, но часы могут быть любого
// количества цифр (для полей "Отработано" и "План").
func isInvalidDurationValue(s string) bool {
	if s == "" {
		return false
	}
	colonIdx := strings.Index(s, ":")
	if colonIdx == -1 {
		for _, ch := range s {
			if ch < '0' || ch > '9' {
				return true
			}
		}
		return false
	}
	if colonIdx == 0 {
		return true
	}
	for _, ch := range s[:colonIdx] {
		if ch < '0' || ch > '9' {
			return true
		}
	}
	return invalidMinutesSuffix(s[colonIdx+1:])
}
