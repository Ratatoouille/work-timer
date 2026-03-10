package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// Стили инициализируются из конфига через initStyles().
var (
	colorAccent  lipgloss.Color
	colorMuted   = lipgloss.Color("8")
	colorSuccess = lipgloss.Color("10")
	colorError   = lipgloss.Color("9")
	colorBreak   lipgloss.Color
	colorResult  lipgloss.Color
	colorWarn    lipgloss.Color

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

	resultBoxStyle   lipgloss.Style
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

	resultBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorResult).
		Padding(0, 2).
		MarginTop(1)

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

func (m Model) View() string {
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
		return content
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
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
	topLine := "🕒  " + title + headerDividerStyle.Render("  │  ") + badge

	var fileLine string
	if m.saveFile != "" {
		dot := statusSuccessStyle.Render("◆")
		if m.isDirty {
			dot = dirtyDotStyle.Render("◆")
		}
		fileLine = dot + "  " + fileNameStyle.Render(filepath.Base(m.saveFile))
	} else {
		fileLine = statusBarStyle.Render(m.locale.NoFileSelected)
	}

	hintsLine := statusBarStyle.Render(m.locale.HeaderHints)
	divider := headerDividerStyle.Render(strings.Repeat("─", divW))
	return topLine + "\n" + fileLine + "\n" + hintsLine + "\n" + divider + "\n"
}

func (m Model) renderSectionHeader(icon, title string, style lipgloss.Style) string {
	text := style.Render(icon + " " + title)
	lineLen := m.dividerWidth() - lipgloss.Width(text) - 1
	if lineLen < 2 {
		lineLen = 2
	}
	line := sectionDividerStyle.Render(" " + strings.Repeat("─", lineLen))
	return "\n" + lipgloss.JoinHorizontal(lipgloss.Left, text, line) + "\n"
}

func (m Model) renderMainFields() string {
	var b strings.Builder
	b.WriteString(m.renderSectionHeader("▸", m.locale.SectionMain, sectionHeaderStyle))
	b.WriteString("\n")
	b.WriteString(m.renderField(FieldStartTime, m.locale.FieldStart, m.startTime) + "\n")

	// Режим 1: оставшееся время
	b.WriteString("\n" + sectionDividerStyle.Render(m.locale.FieldMode1Label) + "\n")
	b.WriteString(m.renderField(FieldWorkTime, m.locale.FieldRemainingTime, m.workTime) + "\n")

	// Режим 2: отработано / план
	b.WriteString("\n" + sectionDividerStyle.Render(m.locale.FieldMode2Label) + "\n")
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
	b.WriteString("\n")

	for i, br := range m.breaks {
		baseIndex := FieldBreaksStart + i*2
		if i > 0 {
			b.WriteString(sectionDividerStyle.Render("  "+strings.Repeat("·", 30)) + "\n")
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
	divider := sectionDividerStyle.Render(strings.Repeat("─", m.dividerWidth()))
	b.WriteString("\n" + divider + "\n")

	if m.err != "" {
		b.WriteString(errorStyle.Render("✗  "+m.err) + "\n")
	}
	if m.result != "" {
		label := resultLabelStyle.Render(m.locale.ResultLabel)
		value := resultValueStyle.Render("⏰  " + m.result)
		b.WriteString(resultBoxStyle.Render(label+value) + "\n")
	}

	return b.String()
}

func (m Model) renderControls() string {
	k := func(s string) string { return controlKeyStyle.Render(s) }
	d := func(s string) string { return statusBarStyle.Render(s) }
	dot := headerDividerStyle.Render("  ·  ")
	divW := m.dividerWidth()

	if m.mode == ModeInsert {
		return controlsBarStyle.Width(divW).Render(k("esc") + "  " + d(m.locale.CtrlNormal))
	}

	line1 := strings.Join([]string{
		k("j/k") + "  " + d(m.locale.CtrlNav),
		k("i") + "  " + d(m.locale.CtrlEdit),
		k("esc") + "  " + d(m.locale.CtrlNormal),
		k("space") + "  " + d(m.locale.CtrlToggle),
		k("t") + "  " + d(m.locale.CtrlNow),
		k("x") + "  " + d(m.locale.CtrlClear),
	}, dot)

	row2 := []string{
		k("a") + " / " + k("d") + "  " + d(m.locale.CtrlAdd+"/"+m.locale.CtrlDel),
		k("s") + "  " + d(m.locale.CtrlSave),
		k("o") + "  " + d(m.locale.CtrlOpen),
	}
	if m.result != "" {
		row2 = append(row2, k("y")+"  "+d(m.locale.CtrlCopy))
	}
	line2 := strings.Join(row2, dot)

	return controlsBarStyle.Width(divW).Render(line1 + "\n" + line2)
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
		labelStyle.Render(label),
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
		labelStyle.Render(label),
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
		labelStyle.Render(label),
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
