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
	colorBreak   color.Color
	colorResult  color.Color
	colorWarn    color.Color

	fieldBoxStyle      lipgloss.Style
	fieldActiveStyle   lipgloss.Style
	fieldInactiveStyle lipgloss.Style
	fieldErrorStyle    lipgloss.Style
	containerStyle     lipgloss.Style
	headerStyle        lipgloss.Style
	modeNormalStyle    lipgloss.Style
	modeInsertStyle    lipgloss.Style
	headerDividerStyle lipgloss.Style
	statusBarStyle     lipgloss.Style
	fileNameStyle      lipgloss.Style
	dirtyDotStyle      lipgloss.Style

	sectionHeaderStyle      lipgloss.Style
	sectionBreakHeaderStyle lipgloss.Style
	sectionDividerStyle     lipgloss.Style

	labelStyle lipgloss.Style

	resultLabelStyle lipgloss.Style
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
	colorBreak = lipgloss.Color(cfg.UI.Colors.Break)
	colorWarn = lipgloss.Color(cfg.UI.Colors.Warn)

	fieldBoxStyle = lipgloss.NewStyle().Padding(0, 1)
	fieldActiveStyle = fieldBoxStyle.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent)
	fieldInactiveStyle = fieldBoxStyle
	fieldErrorStyle = fieldBoxStyle.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorError)

	containerStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorMuted).
		Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorAccent)

	modeNormalStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("0")).
		Background(colorSuccess).
		Padding(0, 1)

	modeInsertStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("0")).
		Background(colorWarn).
		Padding(0, 1)

	headerDividerStyle = lipgloss.NewStyle().Foreground(colorMuted)
	statusBarStyle = lipgloss.NewStyle().Foreground(colorMuted)

	fileNameStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Bold(true)

	dirtyDotStyle = lipgloss.NewStyle().Foreground(colorWarn)

	sectionHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorAccent)

	sectionBreakHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorBreak)

	sectionDividerStyle = lipgloss.NewStyle().Foreground(colorMuted)

	labelStyle = lipgloss.NewStyle().
		Width(cfg.UI.LabelWidth).
		Italic(true).
		Foreground(lipgloss.Color("7"))

	resultLabelStyle = lipgloss.NewStyle().Foreground(colorMuted)
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
	b.WriteString(m.renderMainFields())
	b.WriteString(m.renderBreaks())
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

	divW := m.dividerWidth()
	title := headerStyle.Render("Work Timer")
	badge := modeStyle.Render(" " + modeStr + " ")

	var filePart string
	if m.saveFile != "" {
		dot := statusSuccessStyle.Render("◆")
		if m.isDirty {
			dot = dirtyDotStyle.Render("◆")
		}
		filePart = headerDividerStyle.Render("  │  ") + dot + "  " + fileNameStyle.Render(filepath.Base(m.saveFile))
	} else {
		filePart = headerDividerStyle.Render("  │  ") + statusBarStyle.Render(m.locale.NoFileSelected)
	}

	topLine := "🕒  " + title + headerDividerStyle.Render("  │  ") + badge + filePart

	divider := headerDividerStyle.Render(strings.Repeat("─", divW))
	return topLine + "\n" + divider + "\n"
}

func (m Model) renderSectionHeader(icon, title string, style lipgloss.Style) string {
	text := style.Render(icon + " " + title + " ")
	lineLen := max(m.dividerWidth()-lipgloss.Width(text), 2)
	line := sectionDividerStyle.Render(strings.Repeat("─", lineLen))
	return "\n" + lipgloss.JoinHorizontal(lipgloss.Left, text, line) + "\n"
}

func (m Model) renderSubsectionHeader(title string, style lipgloss.Style) string {
	text := style.Render(title + " ")
	lineLen := max(m.dividerWidth()-lipgloss.Width(text)-2, 2)
	line := sectionDividerStyle.Render(strings.Repeat("─", lineLen))
	return lipgloss.JoinHorizontal(lipgloss.Left, "  ", text, line)
}

func (m Model) renderMainFields() string {
	var b strings.Builder
	b.WriteString(m.renderSectionHeader("▸", m.locale.SectionMain, sectionHeaderStyle))
	b.WriteString(m.renderField(FieldStartTime, m.locale.FieldStart, m.startTime) + "\n")

	// Режим 1: оставшееся время
	mode1Style := sectionDividerStyle
	if m.workTime.Value() != "" {
		mode1Style = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	}
	b.WriteString("\n" + m.renderSubsectionHeader(m.locale.FieldMode1Label, mode1Style) + "\n")
	b.WriteString(m.renderField(FieldWorkTime, m.locale.FieldRemainingTime, m.workTime) + "\n")

	// Режим 2: отработано / план
	mode2Style := sectionDividerStyle
	if m.workTime.Value() == "" && m.worked.Value() != "" && m.plan.Value() != "" {
		mode2Style = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	}
	b.WriteString("\n" + m.renderSubsectionHeader(m.locale.FieldMode2Label, mode2Style) + "\n")
	b.WriteString(m.renderFieldDuration(FieldWorked, m.locale.FieldWorked, m.worked) + "\n")
	b.WriteString(m.renderFieldDuration(FieldPlan, m.locale.FieldPlan, m.plan) + "\n")

	b.WriteString("\n")

	checkboxLabel := m.locale.CheckboxAddTZ
	if m.config.Timezone != "" {
		if label := TimezoneLabel(m.config.Timezone); label != "" {
			checkboxLabel = fmt.Sprintf(m.locale.CheckboxShowIn, label)
		}
	}
	b.WriteString(m.renderCheckbox(FieldAddTZ, checkboxLabel) + "\n")

	return b.String()
}

func (m Model) renderBreaks() string {
	if len(m.breaks) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(m.renderSectionHeader("▸", m.locale.SectionBreaks, sectionBreakHeaderStyle))

	for i, br := range m.breaks {
		baseIndex := FieldBreaksStart + i*2
		if i > 0 {
			dotLen := max(m.dividerWidth()-4, 4)
			b.WriteString(sectionDividerStyle.Render("  "+strings.Repeat("·", dotLen)) + "\n")
		}
		b.WriteString(m.renderField(baseIndex, fmt.Sprintf(m.locale.BreakLeft, i+1), br.from) + "\n")
		b.WriteString(m.renderField(baseIndex+1, fmt.Sprintf(m.locale.BreakRight, i+1), br.to) + "\n")
	}

	return b.String()
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
	if m.err == "" && m.result == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString(m.renderSectionHeader("▸", m.locale.SectionResult, resultValueStyle))

	if m.err != "" {
		b.WriteString(errorStyle.Render("✗  "+m.err) + "\n")
	}
	if m.result != "" {
		label := resultLabelStyle.Render(m.locale.ResultLabel + "  ")
		value := resultValueStyle.Render("⏰  " + m.result)

		_, elapsed, remaining, ok := m.progressInfo()
		if ok {
			elapsedStr := formatDuration(elapsed)
			remainingStr := formatDuration(remaining)
			info := statusBarStyle.Render(fmt.Sprintf("  [%s / %s %s]", elapsedStr, m.locale.Remaining, remainingStr))
			b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, label, value, info) + "\n")
		} else {
			b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, label, value) + "\n")
		}
	}

	return b.String()
}

func (m Model) progressInfo() (percent float64, elapsed, remaining time.Duration, ok bool) {
	if m.err != "" || m.result == "" {
		return 0, 0, 0, false
	}

	now := m.currentTime
	if now.IsZero() {
		now = time.Now()
	}

	endTimeStr := strings.TrimSpace(strings.TrimPrefix(m.result, "⏰ "))
	endTimeStr = strings.TrimSpace(endTimeStr)

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

	var endLoc *time.Location
	if m.addTZ && m.config.Timezone != "" {
		endLoc = TimezoneLocation(m.config.Timezone)
		if endLoc == nil {
			endLoc = time.Local
		}
	} else {
		endLoc = TimezoneLocation(m.config.InputTimezone)
		if endLoc == nil {
			endLoc = time.Local
		}
	}

	endTime := time.Date(now.Year(), now.Month(), now.Day(), endHour, endMin, 0, 0, endLoc)

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

	// Строим кандидатов: start и end могут быть на разных днях.
	// Перебираем смещения дней для start относительно end и выбираем окно,
	// в которое попадает now. Если now после всех вариантов — работа окончена.
	endCandidates := []time.Time{endTime, endTime.AddDate(0, 0, 1)}
	nowIn := now.In(endLoc)

	var bestStart, bestEnd time.Time
	var found bool

	for _, ec := range endCandidates {
		for _, dayOffset := range []int{0, -1, -2} {
			sc := time.Date(ec.Year(), ec.Month(), ec.Day(), startHour, startMin, 0, 0, startLoc).In(endLoc)
			sc = sc.AddDate(0, 0, dayOffset)
			if !sc.Before(ec) {
				continue
			}
			if !nowIn.Before(sc) && !nowIn.After(ec) {
				bestStart = sc
				bestEnd = ec
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		// now после всех окон — рабочий день окончен, 100%
		// Берём последнее окно (конец сегодня) для расчёта elapsed.
		bestEnd = endTime
		bestStart = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), startHour, startMin, 0, 0, startLoc).In(endLoc)
		if !bestStart.Before(bestEnd) {
			bestStart = bestStart.AddDate(0, 0, -1)
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
	return fmt.Sprintf("%dч %dмин", hours, minutes)
}

func (m Model) renderControls() string {
	k := func(s string) string { return controlKeyStyle.Render(s) }
	d := func(s string) string { return statusBarStyle.Render(s) }
	dot := headerDividerStyle.Render("  ·  ")
	divW := m.dividerWidth()

	if m.mode == ModeInsert {
		return controlsBarStyle.Width(divW).Render(k("esc") + "  " + d(m.locale.CtrlEdit))
	}

	line := strings.Join([]string{
		k("j/k") + "  " + d(m.locale.CtrlNav),
		k("i") + "  " + d(m.locale.CtrlEdit),
		k("s") + "  " + d(m.locale.CtrlSave),
		k("o") + "  " + d(m.locale.CtrlOpen),
		k("?") + "  " + d(m.locale.CtrlHelp),
		k("q") + "  " + d(m.locale.CtrlQuit),
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
	b.WriteString(headerStyle.Render(m.locale.FileListTitle) + "\n\n")
	b.WriteString(statusBarStyle.Render(fmt.Sprintf(m.locale.SaveFolder, m.workDir)) + "\n\n")

	divider := sectionDividerStyle.Render(strings.Repeat("─", 40))
	b.WriteString(divider + "\n")

	// Поле поиска
	searchLabel := statusBarStyle.Render("Search: ")
	b.WriteString(searchLabel + m.fileSearchInput.View() + "\n\n")
	b.WriteString(divider + "\n")

	if len(m.availableFiles) == 0 {
		b.WriteString("\n  " + statusBarStyle.Render(m.locale.FileListEmpty) + "\n\n")
		b.WriteString(divider + "\n\n")
		b.WriteString(statusBarStyle.Render(m.locale.FileListHintNew))
	} else {
		b.WriteString("\n")
		for i, file := range m.availableFiles {
			if i == m.fileListCursor {
				b.WriteString(fileListItemActiveStyle.Render("▶ "+file) + "\n")
			} else {
				b.WriteString(fileListItemStyle.Render("  "+file) + "\n")
			}
		}
		b.WriteString("\n" + divider + "\n\n")
		b.WriteString(statusBarStyle.Render(m.locale.FileListHint))
	}
	return promptStyle.Render(b.String())
}

func (m Model) renderField(index int, label string, input textinput.Model) string {
	var style lipgloss.Style
	switch {
	case m.cursor == index:
		style = fieldActiveStyle
	case isInvalidTimeValue(input.Value()):
		style = fieldErrorStyle
	default:
		style = fieldInactiveStyle
	}
	return lipgloss.JoinHorizontal(lipgloss.Left,
		labelStyle.Render(label+"  "),
		style.Render(input.View()),
	)
}

func (m Model) renderFieldDuration(index int, label string, input textinput.Model) string {
	var style lipgloss.Style
	switch {
	case m.cursor == index:
		style = fieldActiveStyle
	case isInvalidDurationValue(input.Value()):
		style = fieldErrorStyle
	default:
		style = fieldInactiveStyle
	}
	return lipgloss.JoinHorizontal(lipgloss.Left,
		labelStyle.Render(label+"  "),
		style.Render(input.View()),
	)
}

func (m Model) renderCheckbox(index int, label string) string {
	checkbox := "[ ]"
	if m.addTZ {
		checkbox = "[x]"
	}
	style := fieldInactiveStyle
	if m.cursor == index {
		style = fieldActiveStyle
	}
	return lipgloss.JoinHorizontal(lipgloss.Left,
		labelStyle.Render(label+"  "),
		style.Render(checkbox),
	)
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
	switch {
	case colonIdx == -1:
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
	case colonIdx == 1:
		if s[0] < '0' || s[0] > '9' {
			return true
		}
		return invalidMinutesSuffix(s[2:])
	case colonIdx == 2:
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
