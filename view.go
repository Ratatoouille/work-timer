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

	promptStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2).
			BorderForeground(lipgloss.Color("12"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14"))

	fileListItemStyle = lipgloss.NewStyle().
				Padding(0, 2)

	fileListItemActiveStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Background(lipgloss.Color("8")).
				Bold(true)
)

func (m Model) View() string {
	if m.helpState == HelpVisible {
		return m.renderHelp()
	}

	if m.mode == ModeSavePrompt {
		return m.renderSavePrompt()
	}

	if m.mode == ModeLoadPrompt {
		return m.renderLoadPrompt()
	}

	if m.mode == ModeFileList {
		return m.renderFileList()
	}

	return containerStyle.Render(m.renderMain())
}

func (m Model) renderMain() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString(m.renderMainFields())
	b.WriteString(m.renderBreaks())
	b.WriteString(m.renderStatusMessage())
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
	statusText := fmt.Sprintf(" %s │ ? help │ q quit │ s save │ o open", modeStr)

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

func (m Model) renderStatusMessage() string {
	if m.statusMessage == "" {
		return ""
	}

	return "\n" + infoStyle.Render(m.statusMessage) + "\n"
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

func (m Model) renderSavePrompt() string {
	prompt := "💾 Сохранить как:\n\n" +
		"Рабочая папка: " + m.workDir + "\n\n" +
		m.filePathInput.View()

	if m.statusMessage != "" {
		prompt += "\n\n" + m.statusMessage
	}

	prompt += "\n\n[Enter] сохранить  [Esc] отмена"

	return promptStyle.Render(prompt)
}

func (m Model) renderLoadPrompt() string {
	prompt := "📂 Загрузить из файла:\n\n" +
		m.filePathInput.View()

	if m.statusMessage != "" {
		prompt += "\n\n" + m.statusMessage
	}

	prompt += "\n\n[Enter] загрузить  [Esc] отмена"

	return promptStyle.Render(prompt)
}

func (m Model) renderFileList() string {
	var b strings.Builder

	b.WriteString("📂 Выберите файл для загрузки\n\n")
	b.WriteString("Папка: " + m.workDir + "\n\n")

	if len(m.availableFiles) == 0 {
		b.WriteString("  Нет сохраненных файлов\n\n")
		b.WriteString("[n] создать новый  [Esc] отмена")
	} else {
		for i, file := range m.availableFiles {
			if i == m.fileListCursor {
				b.WriteString(fileListItemActiveStyle.Render("▶ "+file) + "\n")
			} else {
				b.WriteString(fileListItemStyle.Render("  "+file) + "\n")
			}
		}
		b.WriteString("\n[j/k ↑↓] навигация  [Enter] выбрать  [n] новый  [Esc] отмена")
	}

	return promptStyle.Render(b.String())
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
  s            — сохранить
  o            — открыть список файлов

Insert-режим:
  ввод текста
  Esc          — обратно в Normal

Выбор файла:
  j/k или ↑/↓  — навигация по списку
  Enter        — загрузить выбранный файл
  n            — создать новый файл
  Esc          — отмена

Общие:
  ?            — показать/скрыть help
  Ctrl+S       — быстрое сохранение
  Ctrl+O       — открыть список файлов
  q            — выход

Формат времени: чч:мм

Рабочая папка: %s
  - Все файлы сохраняются в эту папку
  - Папка создается автоматически

Использование:
  ./program              - запуск с выбором файла
  ./program file.json    - запуск с конкретным файлом
`

	return helpBoxStyle.Render(fmt.Sprintf(help, DefaultWorkDir))
}
