package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

var (
	fieldBoxStyle = lipgloss.NewStyle().Padding(0, 1)

	fieldActiveStyle = fieldBoxStyle.
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("8"))

	fieldInactiveStyle = fieldBoxStyle

	resultBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2).
			MarginTop(1)

	containerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Width(22).
			Italic(true)

	resultStyle = lipgloss.NewStyle().
			Bold(true).
			Underline(true)

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Italic(true)

	statusBarStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginBottom(1)

	helpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)

	sectionStyle = lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(1)
)

func (m Model) View() string {
	if m.helpState == HelpVisible {
		return m.renderHelp()
	}

	return containerStyle.Render(m.renderMain())
}

func (m Model) renderMain() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString(m.renderMainFields())
	b.WriteString(m.renderBreaks())
	b.WriteString(m.renderStatus())
	b.WriteString(m.renderControls())

	return b.String()
}

func (m Model) renderHeader() string {
	modeStr := "NORMAL"
	if m.mode == ModeInsert {
		modeStr = "INSERT"
	}

	header := headerStyle.Render("🕒 Work Timer")
	statusText := fmt.Sprintf(" %s │ ? help │ q quit", modeStr)

	if m.saveFile != "" {
		statusText += fmt.Sprintf(" │ ctrl+s save to %s", m.saveFile)
	}

	return header + "\n" + statusBarStyle.Render(statusText) + "\n"
}

func (m Model) renderMainFields() string {
	var b strings.Builder

	b.WriteString(sectionStyle.Render("━━━ Основные параметры ━━━") + "\n")
	b.WriteString(m.renderField(FieldStartTime, "Начало:", m.startTime) + "\n")
	b.WriteString(m.renderField(FieldWorkTime, "Оставшееся время:", m.workTime) + "\n")
	b.WriteString(m.renderField(FieldWorked, "Отработано:", m.worked) + "\n")
	b.WriteString(m.renderField(FieldPlan, "План:", m.plan) + "\n")
	b.WriteString(m.renderCheckbox(FieldAddFour, "Добавить +4 часа") + "\n")

	return b.String()
}

func (m Model) renderBreaks() string {
	if len(m.breaks) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n" + sectionStyle.Render("━━━ Перерывы ━━━") + "\n")

	for i, br := range m.breaks {
		baseIndex := FieldBreaksStart + i*2
		b.WriteString(m.renderField(baseIndex, fmt.Sprintf("Перерыв %d — ушёл:", i+1), br.from) + "\n")
		b.WriteString(m.renderField(baseIndex+1, fmt.Sprintf("Перерыв %d — вернулся:", i+1), br.to) + "\n")

		if i < len(m.breaks)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m Model) renderStatus() string {
	var b strings.Builder

	if m.err != "" {
		b.WriteString("\n" + errorStyle.Render("❌ "+m.err) + "\n")
	}

	if m.result != "" {
		b.WriteString("\n" + resultBoxStyle.Render(
			resultStyle.Render("⏰ Время окончания: "+m.result),
		))
	}

	return b.String()
}

func (m Model) renderControls() string {
	controls := "[j/k ↑↓] nav  [i] edit  [a] add  [d] del  [space] toggle  [esc] normal"

	return "\n\n" + statusBarStyle.Render(controls) + "\n"
}

func (m Model) renderField(index int, label string, input textinput.Model) string {
	style := fieldInactiveStyle
	if m.cursor == index {
		style = fieldActiveStyle
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render(label),
		style.Render(input.View()),
	)
}

func (m Model) renderCheckbox(index int, label string) string {
	checkbox := "[ ]"
	if m.addFour {
		checkbox = "[x]"
	}

	style := fieldInactiveStyle
	if m.cursor == index {
		style = fieldActiveStyle
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render(label),
		style.Render(checkbox),
	)
}

func (m Model) renderHelp() string {
	help := `🛠 Комбинации клавиш

Normal-режим:
  j/k или ↑/↓  — перемещение
  i            — Insert режим
  a            — добавить перерыв
  d            — удалить перерыв
  space        — toggle чекбокс

Insert-режим:
  ввод текста
  Esc          — обратно в Normal

Общие:
  ?            — показать/скрыть help
  Ctrl+S       — сохранить в файл
  q            — выход

Формат времени: чч:мм

Использование:
  ./program              - запуск без сохранения
  ./program data.json    - загрузка/сохранение в data.json
`

	return helpBoxStyle.Render(help)
}
