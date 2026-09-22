package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// containerChrome — горизонтальные накладные расходы контейнера:
// рамка (2) + паддинг слева/справа (4).
const containerChrome = 6

// containerChromeV — вертикальные накладные расходы контейнера:
// рамка (2) + паддинг сверху/снизу (2).
const containerChromeV = 4

// minContainerWidth — минимальная ширина контейнера. Ниже неё поля и
// подписи перестают помещаться, но на очень узких терминалах контейнер
// всё равно не должен выходить за границы окна.
const minContainerWidth = 34

// Цвета берутся из базовой ANSI-палитры терминала, поэтому следуют теме
// пользователя, а не фиксированным индексам. Синий слот (color4) в разных
// темах может сливаться с фоном, поэтому как акцент он не используется:
// фокус показывается жирным/подчёркиванием, а не цветом.
var (
	colorMuted   = lipgloss.BrightBlack
	colorSuccess = lipgloss.Green
	colorError   = lipgloss.Red
	colorResult  = lipgloss.Cyan
	colorWarn    = lipgloss.Yellow

	fieldBoxStyle    lipgloss.Style
	fieldActiveStyle lipgloss.Style
	containerStyle   lipgloss.Style
	headerStyle      lipgloss.Style
	modeNormalStyle  lipgloss.Style
	modeInsertStyle  lipgloss.Style
	statusBarStyle   lipgloss.Style
	fileNameStyle    lipgloss.Style
	dirtyDotStyle    lipgloss.Style
	statusDotStyle   lipgloss.Style
	clockStyle       lipgloss.Style

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
	fieldBoxStyle = lipgloss.NewStyle().Padding(0, 1)
	fieldActiveStyle = fieldBoxStyle.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorWarn).
		Padding(0, 1)

	containerStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorMuted).
		Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.White)

	modeNormalStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorSuccess)

	modeInsertStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Black).
		Background(colorWarn).
		Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().Foreground(colorMuted)
	clockStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.BrightWhite)

	fileNameStyle = lipgloss.NewStyle().
		Foreground(lipgloss.BrightWhite).
		Bold(true)

	dirtyDotStyle = lipgloss.NewStyle().Foreground(colorWarn)
	statusDotStyle = lipgloss.NewStyle().Foreground(colorSuccess)

	sectionBreakHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.BrightWhite)

	sectionDividerStyle = lipgloss.NewStyle().Foreground(colorMuted)

	// Крупный акцентный блок "оставшееся время"
	heroLabelStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.White)
	heroValueStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorResult)
	timerStateStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorSuccess)

	// Вторичные label/value пары
	paramLabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.White)
	paramValueStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.BrightWhite)

	breakTimeStyle = lipgloss.NewStyle().
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
		Foreground(lipgloss.BrightWhite).
		Bold(true)

	promptStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		BorderForeground(colorMuted)

	fileListItemStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.White)

	fileListItemActiveStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Reverse(true).
		Bold(true)

	helpBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorMuted).
		Padding(1, 2)
}

// containerWidth возвращает ширину основного блока в зависимости от терминала.
// Ширина никогда не превышает ширину окна (иначе блок не влезет), имеет
// разумный максимум для широких терминалов и минимум для очень узких.
func (m Model) containerWidth() int {
	if m.width == 0 {
		return 52
	}
	w := m.width - 4
	if w > 92 {
		w = 92
	}
	if w < minContainerWidth {
		w = minContainerWidth
	}
	// Не выходить за пределы окна: обрезка рамки недопустима.
	if w > m.width {
		w = m.width
	}
	return w
}

// contentWidth — доступная ширина внутри рамки и паддинга контейнера.
func (m Model) contentWidth() int {
	w := m.containerWidth() - containerChrome
	if w < 8 {
		w = 8
	}
	return w
}

func (m Model) dividerWidth() int {
	w := m.contentWidth()
	if w < 16 {
		w = 16
	}
	return w
}

// truncate обрезает строку (включая ANSI-последовательности) до maxWidth
// видимых ячеек, добавляя многоточие, если обрезка произошла.
func truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	if maxWidth <= 1 {
		return ansi.Truncate(s, maxWidth, "")
	}
	return ansi.Truncate(s, maxWidth, "…")
}

// labelCol — ширина колонки label/value пар. Равна ширине самого длинного
// видимого label (чтобы пары выравнивались и длинные подписи не переносились),
// но не больше половины доступной ширины контейнера, оставляя место для
// значения. На узких терминалах подпись обрезается до этой ширины.
func (m Model) labelCol(labels ...string) int {
	widest := 0
	maxAvail := max(m.contentWidth()-12, 6)
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

// labelCell рендерит подпись в колонку фиксированной ширины col. Если подпись
// длиннее колонки, она обрезается (lipgloss.Width не усекает содержимое).
func labelCell(style lipgloss.Style, label string, col int) string {
	return style.Width(col).Render(truncate(label, col))
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

	// Страховка от переполнения по ширине: любая строка обрезается до ширины
	// окна, чтобы узкий терминал не давал переносов и «разъезжающегося» UI.
	content = m.clipWidth(content, m.width)

	// Если контент выше окна, прижимаем его к верху: иначе Center отрежет
	// и заголовок, и футер одновременно.
	vertAlign := lipgloss.Center
	if lipgloss.Height(content) > m.height {
		vertAlign = lipgloss.Top
	}

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, vertAlign, content))
	v.AltScreen = true
	return v
}

// clipWidth обрезает каждую строку блока до maxWidth видимых ячеек.
func (m Model) clipWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if lipgloss.Width(line) > maxWidth {
			lines[i] = ansi.Truncate(line, maxWidth, "")
		}
	}
	return strings.Join(lines, "\n")
}

// availableHeight — высота, доступная контенту внутри рамки и паддинга
// контейнера. Возвращает 0, если высота окна ещё неизвестна.
func (m Model) availableHeight() int {
	if m.height <= 0 {
		return 0
	}
	h := m.height - containerChromeV
	if h < 4 {
		h = 4
	}
	return h
}

// renderMain собирает главный экран с учётом доступной высоты. Блоки
// добавляются по убыванию важности; при нехватке места второстепенные
// (прогресс, перерывы, параметры, результат) отбрасываются, чтобы
// заголовок, hero и футер всегда оставались на экране.
func (m Model) renderMain() string {
	budget := m.availableHeight()

	header := m.renderHeader()
	controls := m.renderControls()
	status := m.renderStatusMessage()
	timer := m.renderTimerRow()

	// Блоки в порядке убывания важности. Хвост отбрасывается первым,
	// когда итоговый блок не влезает в доступную высоту.
	tiers := [][]string{
		// 0: полный экран
		{m.renderHero(true), timer, m.renderBreaks(), m.renderParams(), m.renderResult()},
		// 1: без прогресс-бара
		{m.renderHero(false), timer, m.renderBreaks(), m.renderParams(), m.renderResult()},
		// 2: без результата
		{m.renderHero(false), timer, m.renderBreaks(), m.renderParams()},
		// 3: без перерывов
		{m.renderHero(false), timer, m.renderParams()},
		// 4: минимум — заголовок, hero, время, футер
		{m.renderHero(false), timer},
	}

	if budget == 0 {
		return m.clipWidth(m.assembleMain(header, tiers[0], status, controls), m.contentWidth())
	}

	for _, tier := range tiers {
		block := m.assembleMain(header, tier, status, controls)
		if fits(block, budget) {
			return m.clipWidth(block, m.contentWidth())
		}
	}
	// Ничего не влезает — отдаём минимальный вариант (обрежется сверху).
	return m.clipWidth(m.assembleMain(header, tiers[len(tiers)-1], status, controls), m.contentWidth())
}

// assembleMain склеивает главный экран из заголовка, набора блоков, статуса
// и футера.
func (m Model) assembleMain(header string, blocks []string, status, controls string) string {
	var b strings.Builder
	b.WriteString(header)
	for _, blk := range blocks {
		b.WriteString(blk)
	}
	b.WriteString(status)
	b.WriteString(controls)
	return b.String()
}

// fits сообщает, помещается ли блок в доступную высоту.
func fits(block string, budget int) bool {
	return lipgloss.Height(block) <= budget
}

// blockLines — сколько строк занимает блок при склейке с другими блоками.
// Каждый перевод строки завершает ровно одну строку; висячий хвост без "\n"
// считается отдельной строкой.
func blockLines(s string) int {
	if s == "" {
		return 0
	}
	n := strings.Count(s, "\n")
	if !strings.HasSuffix(s, "\n") {
		n++
	}
	return n
}

// renderBox рендерит тело в рамке, отбрасывая хвостовые переводы строки, из-за
// которых lipgloss добавляет лишнюю пустую строку внутри рамки.
func renderBox(style lipgloss.Style, body string) string {
	return style.Render(strings.TrimRight(body, "\n"))
}

// listWindow возвращает диапазон [start, end) видимых элементов списка так,
// чтобы курсор был в окне, а окно не выходило за пределы доступных строк.
// avail <= 0 означает отсутствие ограничения.
func listWindow(total, cursor, offset, avail int) (start, end int) {
	if total <= 0 {
		return 0, 0
	}
	if avail < 1 {
		avail = 1
	}
	if avail >= total {
		return 0, total
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= total {
		cursor = total - 1
	}
	start = offset
	// Держим курсор внутри окна.
	if cursor < start {
		start = cursor
	}
	if cursor >= start+avail {
		start = cursor - avail + 1
	}
	if start < 0 {
		start = 0
	}
	if start+avail > total {
		start = total - avail
	}
	return start, start + avail
}

func (m Model) renderHeader() string {
	modeStyle := modeNormalStyle
	modeStr := m.locale.ModeNormal
	if m.mode == ModeInsert {
		modeStyle = modeInsertStyle
		modeStr = m.locale.ModeInsert
	}

	clock := ""
	if !m.currentTime.IsZero() {
		clock = clockStyle.Render(m.currentTime.Format("15:04"))
	}

	// Фиксированная часть (логотип, режим, часы) всегда видна; имя файла
	// получает остаток строки и обрезается при нехватке места.
	prefix := headerStyle.Render("◉ WORK TIMER") + "  " + modeStyle.Render(modeStr)
	suffix := ""
	if clock != "" {
		suffix = "  " + clock
	}

	var filePart string
	if m.saveFile != "" {
		dot := statusDotStyle.Render("●")
		if m.isDirty {
			dot = dirtyDotStyle.Render("●")
		}
		nameAvail := m.contentWidth() - lipgloss.Width(prefix) - lipgloss.Width(suffix) - lipgloss.Width(dot) - 4
		filePart = dot + "  " + fileNameStyle.Render(truncate(filepath.Base(m.saveFile), nameAvail))
	} else {
		filePart = statusBarStyle.Render(m.locale.NoFileSelected)
	}

	line := prefix + "  " + filePart
	if clock != "" {
		line += suffix
	}
	return line + "\n"
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
// "ОСТАЛОСЬ 09:00" рядом с "ГОТОВО". withProgress=false убирает полосу
// прогресса и лишние пустые строки — для компактного режима.
func (m Model) renderHero(withProgress bool) string {
	percent, _, _, ok := m.progressInfo()
	done := m.runtimeState() == stateDone

	value := m.heroRemaining()
	badge := m.statusBadge()
	line := value
	if badge != "" {
		line = lipgloss.JoinHorizontal(lipgloss.Center, value, badge)
	}

	label := heroLabelStyle.Render(m.locale.RemainingLabel + ":")
	if done {
		label = statusSuccessStyle.Bold(true).Render(m.locale.HeroDoneLabel)
	}

	if !withProgress {
		return "\n" + label + "\n" + line + "\n"
	}

	var progress string
	if ok {
		progress = m.renderProgressTimeline(percent)
	}

	return "\n" + label + "\n" +
		line + "\n\n" +
		progress + "\n\n"
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
	// Перерывы задаются в input_timezone (как и время начала/окончания),
	// поэтому текущее время сравниваем в той же зоне.
	inputLoc := TimezoneLocation(m.config.InputTimezone)
	if inputLoc == nil {
		inputLoc = time.Local
	}
	nowStr := m.currentTime.In(inputLoc).Format("15:04")
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
			start = paramValueStyle.Underline(true).Render(m.startTime.Value())
		}
	} else {
		start = statusBarStyle.Render("—:—")
		if m.cursor == FieldStartTime {
			start = offStyle.Render("▍" + m.locale.PlaceholderTime)
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Left,
		labelCell(startLabel, m.locale.FieldStart, col),
		"  ",
		start,
		sep,
		labelCell(paramLabelStyle, m.locale.FieldEnd, col),
		"  ",
		end,
	)

	return row + "\n"
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
	// JoinHorizontal выравнивает многострочную обводку поля по левому краю,
	// чтобы отступ "  " применялся ко всем строкам рамки. Незафокусированное
	// поле продолжаем показывать как значение, а не заглушку.
	if m.mode == ModeInsert && (focusedFrom || focusedTo) {
		renderOther := func(val string) string {
			if val != "" {
				return breakTimeStyle.Render(val)
			}
			return offStyle.Render("—:—")
		}
		if focusedFrom {
			return lipgloss.JoinHorizontal(lipgloss.Left, "  ",
				fieldActiveStyle.Render(br.from.View()), "  ", renderOther(to))
		}
		return lipgloss.JoinHorizontal(lipgloss.Left, "  ",
			renderOther(from), "  ", fieldActiveStyle.Render(br.to.View()))
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
	// JoinHorizontal нужен, чтобы в Insert-режиме многострочная обводка поля
	// (бокс из fieldActiveStyle) выравнивалась: последующие строки рамки
	// получают тот же отступ, что и первая.
	return lipgloss.JoinHorizontal(lipgloss.Left, "  ", row)
}

func (m Model) renderParams() string {
	var b strings.Builder
	b.WriteString(m.renderSectionHeader(m.locale.SectionParams, sectionBreakHeaderStyle))

	col := m.labelCol(m.locale.StatMode, checkboxLabel(m))

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

	// Поля ввода обоих режимов. Всегда рендерим, чтобы у каждого поля был
	// видимый экранный фокус; неактивная группа приглушается.
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
	tzFocused := m.cursor == FieldAddTZ
	tzLabel := m.valueLabelStyle(tzFocused)
	if tzFocused {
		s := "Нет"
		if m.addTZ {
			s = "Да"
		}
		tzVal = paramValueStyle.Underline(true).Render(s)
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, labelCell(tzLabel, checkboxLabel(m), col), "  ", tzVal) + "\n")

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
		labelCell(paramLabelStyle, label, col),
		"  ",
		value,
	)
}

// renderValueFieldAt рисует поле ввода. В Insert-режиме на активном поле
// показывает настоящий textinput для ввода, иначе — значение с фокусной
// подсветкой.
func (m Model) renderValueFieldAt(index int, label string, input textinput.Model, isDuration bool, col int) string {
	focused := m.cursor == index
	value := input.Value()
	labelStyle := m.valueLabelStyle(focused)

	if m.mode == ModeInsert && focused {
		return lipgloss.JoinHorizontal(lipgloss.Left,
			labelCell(labelStyle, label, col),
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
			labelCell(labelStyle, label, col),
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
		valStyle = paramValueStyle.Underline(true)
	}
	if invalid {
		valStyle = statusErrorStyle
	}

	return lipgloss.JoinHorizontal(lipgloss.Left,
		labelCell(labelStyle, label, col),
		"  ",
		valStyle.Render(value),
	)
}

func (m Model) valueLabelStyle(focused bool) lipgloss.Style {
	s := paramLabelStyle
	if focused {
		s = s.Bold(true).Underline(true)
	}
	return s
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

	// Статус действия выводится над футером, сразу под горизонтальным
	// разделителем, отделяясь от статичных подсказок клавиш.
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
		labelCell(paramLabelStyle, label, col),
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
	pct := percent
	if pct > 1.0 {
		pct = 1.0
	}
	if pct < 0 {
		pct = 0
	}

	barWidth := min(max(m.contentWidth()-8, 10), 50)
	if avail > 0 {
		barWidth = min(avail, barWidth)
	}
	if barWidth > m.dividerWidth() {
		barWidth = m.dividerWidth()
	}
	barWidth = max(barWidth, 4)

	// Одна формула: доля → количество заполненных сегментов и процент.
	filled := int(pct * float64(barWidth))
	filled = min(filled, barWidth)
	filled = max(filled, 0)
	pctText := fmt.Sprintf("%.0f%%", pct*100)

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	return lipgloss.NewStyle().Foreground(colorResult).Render(bar) + " " + statusBarStyle.Render(pctText)
}

// renderProgressTimeline — прогресс-бар рабочего дня. Время начала/окончания
// выводится только в отдельном блоке renderTimerRow (без дублирования).
func (m Model) renderProgressTimeline(percent float64) string {
	total := max(m.dividerWidth(), 20)
	return m.renderProgressBar(percent, total)
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

	// totalWork — фактическое рабочее время (окно минус перерывы).
	// percent и remaining считаются в одних единицах (работа), поэтому
	// прогресс-бар корректно доходит до 100% к концу дня.
	totalWork := max(totalDuration-breaksDur, 0)

	if totalWork <= 0 {
		return 0, 0, 0, false
	}

	p := float64(elapsedDuration) / float64(totalWork)
	if p > 1.0 {
		p = 1.0
	}
	if p < 0 {
		p = 0
	}

	remainingDuration := max(totalWork-elapsedDuration, 0)

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
	divW := m.dividerWidth()

	if m.mode == ModeInsert {
		return controlsBarStyle.Width(divW).Render(k("esc") + "  " + d(m.locale.CtrlEdit))
	}

	items := []string{
		k("j/k") + " " + d(m.locale.CtrlNav),
		k("i") + " " + d(m.locale.CtrlEdit),
		k("s") + " " + d(m.locale.CtrlSave),
		k("o") + " " + d(m.locale.CtrlOpen),
		k("?") + " " + d(m.locale.CtrlHelp),
		k("q") + " " + d(m.locale.CtrlQuit),
	}

	line := strings.Join(wrapItems(items, divW, "   "), "\n")
	return controlsBarStyle.Width(divW).Render(line)
}

// wrapItems упаковывает элементы подсказок в строки, не превышающие width.
// На узких терминалах футер переносится на несколько строк вместо обрезки.
func wrapItems(items []string, width int, sep string) []string {
	if width <= 0 {
		return []string{strings.Join(items, sep)}
	}
	var lines []string
	cur := ""
	for _, it := range items {
		switch {
		case cur == "":
			cur = it
		case lipgloss.Width(cur)+lipgloss.Width(sep)+lipgloss.Width(it) <= width:
			cur += sep + it
		default:
			lines = append(lines, cur)
			cur = it
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

func (m Model) renderSavePrompt() string {
	body := headerStyle.Render(m.locale.SaveTitle) +
		"\n\n" + statusBarStyle.Render(fmt.Sprintf(m.locale.SaveFolder, tildePath(m.workDir))) +
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

	// Путь — приглушённый, не доминирует. Сокращаем до ~.
	b.WriteString(statusBarStyle.Render(fmt.Sprintf(m.locale.SaveFolder, tildePath(m.workDir))) + "\n\n")

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
		return renderBox(promptStyle, b.String())
	}

	// Хвост экрана (после списка): отступ, статус, двухстрочный футер.
	fk := func(s string) string { return controlKeyStyle.Render(s) }
	fd := func(s string) string { return statusBarStyle.Render(s) }
	var tail strings.Builder
	if m.renaming {
		tail.WriteString("\n" + statusBarStyle.Render(m.locale.RenamePrompt+" ") + m.renameInput.View() + "\n")
		tail.WriteString(statusBarStyle.Render("[Enter] rename  [Esc] cancel") + "\n")
	} else if m.confirmDelete && m.statusMessage != "" {
		tail.WriteString("\n" + statusWarnStyle.Render(m.statusMessage) + "\n")
	}
	tail.WriteString("\n" + fk("j/k") + " " + fd(m.locale.CtrlNav) + "   " + fk("Enter") + " " + fd(m.locale.CtrlOpen) +
		"   " + fk("/") + " " + fd(m.locale.CtrlSearch) + "\n")
	tail.WriteString(fk("d") + " " + fd(m.locale.CtrlDelete) + "   " + fk("r") + " " + fd(m.locale.CtrlRename) +
		"   " + fk("n") + " " + fd(m.locale.CtrlNew) + "   " + fk("Esc") + " " + fd(m.locale.CtrlCancel) + "\n")

	// Список файлов с прокруткой: показываем только то, что влезает в высоту.
	budget := m.availableHeight() - blockLines(b.String()) - blockLines(tail.String())
	avail := budget
	if len(m.availableFiles) > avail {
		avail -= 2 // место под индикаторы ↑/↓
	}
	start, end := listWindow(len(m.availableFiles), m.fileListCursor, m.fileListOffset, avail)

	if start > 0 {
		b.WriteString(statusBarStyle.Render(fmt.Sprintf("  ↑ %d", start)) + "\n")
	}
	for i := start; i < end; i++ {
		file := m.availableFiles[i]
		if i == m.fileListCursor {
			b.WriteString(fileListItemActiveStyle.Render("▸ "+file) + "\n")
		} else {
			b.WriteString(fileListItemStyle.Render("  "+file) + "\n")
		}
	}
	if end < len(m.availableFiles) {
		b.WriteString(statusBarStyle.Render(fmt.Sprintf("  ↓ %d", len(m.availableFiles)-end)) + "\n")
	}

	b.WriteString(tail.String())

	return renderBox(promptStyle, b.String())
}

func (m Model) renderPresetList() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(m.locale.PresetTitle) + "\n\n")

	if len(m.config.Breaks) == 0 {
		b.WriteString("\n  " + statusBarStyle.Render(m.locale.PresetEmpty) + "\n\n")
	} else {
		tail := "\n" + sectionDividerStyle.Render(strings.Repeat("─", m.dividerWidth())) + "\n\n" +
			statusBarStyle.Render(m.locale.PresetHint) + "\n"
		budget := m.availableHeight() - blockLines(b.String()) - blockLines(tail)
		avail := budget
		if len(m.config.Breaks) > avail {
			avail -= 2
		}
		start, end := listWindow(len(m.config.Breaks), m.presetCursor, m.presetOffset, avail)

		if start > 0 {
			b.WriteString(statusBarStyle.Render(fmt.Sprintf("  ↑ %d", start)) + "\n")
		}
		for i := start; i < end; i++ {
			preset := m.config.Breaks[i]
			label := fmt.Sprintf("%s  (%s – %s)", preset.Name, preset.From, preset.To)
			if i == m.presetCursor {
				b.WriteString(fileListItemActiveStyle.Render("▸ "+label) + "\n")
			} else {
				b.WriteString(fileListItemStyle.Render("  "+label) + "\n")
			}
		}
		if end < len(m.config.Breaks) {
			b.WriteString(statusBarStyle.Render(fmt.Sprintf("  ↓ %d", len(m.config.Breaks)-end)) + "\n")
		}
		b.WriteString(tail)
		return renderBox(promptStyle, b.String())
	}

	divider := sectionDividerStyle.Render(strings.Repeat("─", m.dividerWidth()))
	b.WriteString(divider + "\n\n")
	b.WriteString(statusBarStyle.Render(m.locale.PresetHint))
	return renderBox(promptStyle, b.String())
}

func (m Model) renderHistory() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(m.locale.HistoryTitle) + "\n\n")

	if len(m.historyEntries) == 0 {
		b.WriteString("\n  " + statusBarStyle.Render(m.locale.HistoryEmpty) + "\n\n")
		b.WriteString(m.historyFooter())
		return renderBox(promptStyle, b.String())
	}

	// Список записей в том же стиле, что и список файлов: маркер "▸" у
	// выбранной записи, фиксированные колонки.
	// Заголовок колонок: Дата | Начало | Окончание | Перерывы.
	colDate := m.labelCol(m.locale.HistoryColDate, "00-00-0000")
	colTime := m.labelCol(m.locale.HistoryColStart, m.locale.HistoryColEnd, "00:00")
	headerLine := fmt.Sprintf("%-*s  %-*s  %-*s  %s",
		colDate, m.locale.HistoryColDate,
		colTime, m.locale.HistoryColStart,
		colTime, m.locale.HistoryColEnd,
		m.locale.HistoryColBreak)
	b.WriteString(statusBarStyle.Render("  "+headerLine) + "\n")
	b.WriteString(statusBarStyle.Render("  "+strings.Repeat("─", lipgloss.Width(headerLine))) + "\n")

	// Прокрутка: видимое окно записей зависит от высоты терминала.
	tail := m.historyFooter()
	budget := m.availableHeight() - blockLines(b.String()) - blockLines(tail)
	avail := budget
	if len(m.historyEntries) > avail {
		avail -= 2
	}
	start, end := listWindow(len(m.historyEntries), m.historyCursor, m.historyOffset, avail)

	if start > 0 {
		b.WriteString(statusBarStyle.Render(fmt.Sprintf("  ↑ %d", start)) + "\n")
	}
	for i := start; i < end; i++ {
		e := m.historyEntries[i]
		startStr := e.StartTime
		if startStr == "" {
			startStr = "—:—"
		}
		endStr := e.EndTime
		if endStr == "" {
			endStr = "—:—"
		}
		// Перерывы: количество × суммарная длительность, как на главном экране.
		// Без перерывов — "—".
		breaksInfo := "—"
		if e.Breaks > 0 {
			d := parseBreakDur(e.BreaksDur)
			breaksInfo = fmt.Sprintf("%d × %s", e.Breaks, formatShortDuration(d, m.locale))
		}
		dateStr := formatDisplayDate(e.Date)
		line := fmt.Sprintf("%-*s  %-*s  %-*s  %s",
			colDate, dateStr,
			colTime, startStr,
			colTime, endStr,
			breaksInfo)
		if i == m.historyCursor {
			b.WriteString(fileListItemActiveStyle.Render("▸ "+line) + "\n")
		} else {
			b.WriteString(fileListItemStyle.Render("  "+line) + "\n")
		}
	}
	if end < len(m.historyEntries) {
		b.WriteString(statusBarStyle.Render(fmt.Sprintf("  ↓ %d", len(m.historyEntries)-end)) + "\n")
	}

	b.WriteString(tail)

	return renderBox(promptStyle, b.String())
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
	// Справка в той же визуальной системе, что и остальные экраны:
	// заголовок + сгруппированные строки «клавиша — описание».
	fk := func(s string) string { return controlKeyStyle.Render(s) }
	fd := func(s string) string { return statusBarStyle.Render(s) }
	row := func(key, desc string) string {
		return "  " + fk(key) + "  " + fd(desc)
	}

	var b strings.Builder
	b.WriteString(headerStyle.Render(m.locale.HelpTitle) + "\n\n")

	writeSection := func(title string, rows string) {
		b.WriteString(sectionBreakHeaderStyle.Render("▾ "+title) + "\n")
		b.WriteString(rows + "\n")
	}

	writeSection(m.locale.HelpNormal,
		row("j/k ↑/↓ ←/→", m.locale.CtrlNav)+"\n"+
			row("i", m.locale.CtrlEdit)+"\n"+
			row("t", m.locale.CtrlCurrentTime)+"\n"+
			row("a", m.locale.CtrlAddBreak)+"\n"+
			row("p", m.locale.CtrlPreset)+"\n"+
			row("d", m.locale.CtrlDelBreak)+"\n"+
			row("H", m.locale.CtrlHistory)+"\n"+
			row("space", m.locale.CtrlCheckbox)+"\n"+
			row("y", m.locale.CtrlCopy)+"\n"+
			row("1-9", m.locale.CtrlQuickInput)+"\n"+
			row("s = Ctrl+S", m.locale.CtrlSave)+"\n"+
			row("o = Ctrl+O", m.locale.CtrlOpen)+"\n")

	writeSection(m.locale.HelpModes,
		row("1", m.locale.HelpMode1Desc)+"\n"+
			row("2", m.locale.HelpMode2Desc)+"\n"+
			row("x", m.locale.CtrlClear)+"\n")

	writeSection(m.locale.HelpFileList,
		row("j/k ↑/↓", m.locale.CtrlNav)+"\n"+
			row("/", m.locale.CtrlSearch)+"\n"+
			row("Enter", m.locale.CtrlSelect)+"\n"+
			row("d", m.locale.CtrlDelete)+"\n"+
			row("r", m.locale.CtrlRename)+"\n"+
			row("n", m.locale.CtrlNew)+"\n"+
			row("Esc", m.locale.CtrlCancel)+"\n")

	writeSection(m.locale.HelpGeneral,
		row("?", m.locale.CtrlHelp)+"\n"+
			row("q", m.locale.CtrlQuit)+"\n")

	b.WriteString(statusBarStyle.Render(m.locale.HowModes) + "\n\n")
	b.WriteString(fd(m.locale.HelpConfig+": ") + fk(ConfigPath) + "\n")
	b.WriteString(fd(m.locale.HelpFolder+": ") + fk(DefaultWorkDir) + "\n")

	// Прокрутка: на низком терминале показываем только часть строк справки.
	body := b.String()
	avail := m.availableHeight()
	if avail > 0 {
		lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
		if len(lines) > avail {
			offset := m.helpOffset
			if offset < 0 {
				offset = 0
			}
			if offset > len(lines)-avail {
				offset = len(lines) - avail
			}
			lines = lines[offset:]
			if len(lines) > avail {
				lines = lines[:avail]
			}
			body = strings.Join(lines, "\n") + "\n"
		}
	}

	return renderBox(helpBoxStyle, body)
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
