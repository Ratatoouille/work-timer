package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Mode int

const (
	ModeVim Mode = iota
	ModeInsert
)

type HelpState int

const (
	HelpHidden HelpState = iota
	HelpVisible
)

type Break struct {
	from textinput.Model
	to   textinput.Model
}

type model struct {
	mode Mode

	startTime textinput.Model
	workTime  textinput.Model // оставшееся рабочее время (чч:мм)
	worked    textinput.Model // уже отработано (чч:мм)
	plan      textinput.Model // сколько должно быть отработано (чч:мм)
	addFour   bool

	breaks []Break

	cursor int
	result string
	err    string

	help HelpState
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func initialModel() model {
	start := textinput.New()
	start.Placeholder = "чч:мм"
	start.Width = 5
	start.Focus()

	work := textinput.New()
	work.Width = 5
	work.Placeholder = "чч:мм"

	worked := textinput.New()
	worked.Width = 5
	worked.Placeholder = "чч:мм"

	plan := textinput.New()
	plan.Width = 5
	plan.Placeholder = "чч:мм"

	return model{
		mode:      ModeVim,
		startTime: start,
		workTime:  work,
		worked:    worked,
		plan:      plan,
		breaks:    []Break{},
		cursor:    0,
		help:      HelpHidden,
	}
}

func (m *model) recalc() {
	start, err := parseClock(m.startTime.Value())
	if err != nil {
		m.err = "Неверное время начала"
		m.result = ""

		return
	}

	var remain time.Duration
	if m.workTime.Value() != "" {
		remain, err = parseDuration(m.workTime.Value())

		if err != nil {
			m.err = "Неверное оставшееся время"
			m.result = ""

			return
		}
	} else if m.worked.Value() != "" && m.plan.Value() != "" {
		workedDur, err1 := parseDuration(m.worked.Value())
		planDur, err2 := parseDuration(m.plan.Value())

		if err1 != nil || err2 != nil {
			m.err = "Неверный формат отработано/план"
			m.result = ""

			return
		}

		remain = planDur - workedDur

		if remain < 0 {
			remain = 0
		}
	} else {
		m.err = "Введите либо оставшееся время, либо отработано/план"
		m.result = ""

		return
	}

	totalBreak := time.Duration(0)
	for _, br := range m.breaks {
		if br.from.Value() == "" || br.to.Value() == "" {
			continue
		}

		f, err1 := parseClock(br.from.Value())
		t, err2 := parseClock(br.to.Value())

		if err1 != nil || err2 != nil {
			m.err = "Неверный формат перерыва"
			m.result = ""
			return
		}

		totalBreak += t.Sub(f)
	}

	end := start.Add(remain).Add(totalBreak)

	if m.addFour {
		end = end.Add(4 * time.Hour)

		m.result = end.Format("15:04") + " KRSK"
	} else {
		m.result = end.Format("15:04")
	}

	m.err = ""
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "?":
			if m.help == HelpHidden {
				m.help = HelpVisible
			} else {
				m.help = HelpHidden
			}
		}

		if m.help == HelpVisible {
			break
		}

		switch m.mode {
		case ModeVim:
			switch msg.String() {
			case "i":
				m.mode = ModeInsert
				m.focusCurrent()
			case "j", "down":
				if m.cursor < m.fieldCount()-1 {
					m.cursor++

					m.blurAll()
					m.focusCurrent()
				}
			case "k", "up":
				if m.cursor > 0 {
					m.cursor--

					m.blurAll()
					m.focusCurrent()
				}
			case "a":
				br1 := textinput.New()
				br1.Placeholder = "чч:мм"
				br1.Width = 5
				br1.Blur()

				br2 := textinput.New()
				br2.Width = 5
				br2.Placeholder = "чч:мм"
				br1.Blur()

				m.breaks = append(m.breaks, Break{from: br1, to: br2})
			case "d":
				idx, ok := m.cursorBreakIndex()

				if ok {
					m.breaks = append(m.breaks[:idx], m.breaks[idx+1:]...)
					if m.cursor > 3 {
						m.cursor--
					}
				}
			case " ":
				if m.cursor == 4 {
					m.addFour = !m.addFour
				}
			}
		case ModeInsert:
			var cmd tea.Cmd

			switch m.cursor {
			case 0:
				m.startTime, cmd = m.startTime.Update(msg)
			case 1:
				m.workTime, cmd = m.workTime.Update(msg)
			case 2:
				m.worked, cmd = m.worked.Update(msg)
			case 3:
				m.plan, cmd = m.plan.Update(msg)
			default:
				idx := (m.cursor - 4) / 2

				if (m.cursor-4)%2 == 0 {
					m.breaks[idx].from, cmd = m.breaks[idx].from.Update(msg)
				} else {
					m.breaks[idx].to, cmd = m.breaks[idx].to.Update(msg)
				}
			}

			cmds = append(cmds, cmd)

			if msg.String() == "esc" {
				m.mode = ModeVim

				m.blurAll()
				m.focusCurrent()
			}
		}
	}

	m.recalc()

	return m, tea.Batch(cmds...)
}

func (m *model) blurAll() {
	fields := []*textinput.Model{
		&m.startTime, &m.workTime, &m.worked, &m.plan,
	}

	for i := range fields {
		fields[i].Blur()
		fields[i].CursorEnd() // сброс позиции курсора
	}

	for i := range m.breaks {
		m.breaks[i].from.Blur()
		m.breaks[i].from.CursorEnd()
		m.breaks[i].to.Blur()
		m.breaks[i].to.CursorEnd()
	}
}

func (m model) renderField(index int, label string, input textinput.Model) string {
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

func (m model) renderCheckbox(index int, label string) string {
	box := "[ ]"
	if m.addFour {
		box = "[x]"
	}

	content := label + " " + box

	style := fieldInactiveStyle
	if m.cursor == index {
		style = fieldActiveStyle
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render(""),
		style.Render(content),
	)
}

func (m *model) focusCurrent() {
	// Снимаем фокус со всех полей
	m.startTime.Blur()
	m.workTime.Blur()
	m.worked.Blur()
	m.plan.Blur()

	for _, br := range m.breaks {
		br.from.Blur()
		br.to.Blur()
	}

	// Ставим фокус только на текущее поле (textinput)
	switch {
	case m.cursor == 0:
		m.startTime.Focus()
	case m.cursor == 1:
		m.workTime.Focus()
	case m.cursor == 2:
		m.worked.Focus()
	case m.cursor == 3:
		m.plan.Focus()
	default:
		// Только если cursor >= 5 (breaks), чтобы не попасть на чекбокс
		if m.cursor >= 5 {
			idx := (m.cursor - 4) / 2

			if idx < len(m.breaks) {
				if (m.cursor-4)%2 == 0 {
					m.breaks[idx].from.Focus()
				} else {
					m.breaks[idx].to.Focus()
				}
			}
		}
	}
}

var (
	fieldBoxStyle = lipgloss.NewStyle().
			Padding(0, 1)

	fieldActiveStyle = fieldBoxStyle.Copy().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("8"))

	fieldInactiveStyle = fieldBoxStyle.Copy()

	resultBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2).
			MarginTop(1)

	box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true)

	labelStyle = lipgloss.NewStyle().
			Width(22)

	resultStyle = lipgloss.NewStyle().
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Bold(true)

	statusBarStyle = lipgloss.NewStyle().
			Padding(0, 1)

	helpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2).
			Width(60)
)

func (m model) View() string {
	if m.help == HelpVisible {
		return m.viewHelp()
	}

	var body strings.Builder

	modeStr := "INSERT"
	if m.mode == ModeVim {
		modeStr = "VIM"
	}

	// HEADER
	header := headerStyle.Render("🕒 Work Timer")
	status := statusBarStyle.Render(
		fmt.Sprintf(" mode: %s | ? help | q quit ", modeStr),
	)

	body.WriteString(header + "\n")
	body.WriteString(status + "\n\n")

	// MAIN FIELDS
	body.WriteString(
		m.renderField(0, "Начало:", m.startTime) + "\n",
	)
	body.WriteString(
		m.renderField(1, "Оставшееся время:", m.workTime) + "\n",
	)
	body.WriteString(
		m.renderField(2, "Отработано:", m.worked) + "\n",
	)
	body.WriteString(
		m.renderField(3, "План:", m.plan) + "\n\n",
	)
	body.WriteString(
		m.renderCheckbox(4, "Добавить +4 часа") + "\n\n",
	)

	// BREAKS
	for n, br := range m.breaks {
		body.WriteString(
			m.renderField(4+n*2, fmt.Sprintf("Перерыв %d — ушёл:", n+1), br.from) + "\n",
		)
		body.WriteString(
			m.renderField(5+n*2, fmt.Sprintf("Перерыв %d — вернулся:", n+1), br.to) + "\n",
		)
	}

	// RESULT / ERROR
	if m.err != "" {
		body.WriteString("\n" + errorStyle.Render("❌ "+m.err) + "\n")
	}

	if m.result != "" {
		body.WriteString(
			resultBoxStyle.Render(
				resultStyle.Render("⏰ Время окончания: " + m.result),
			),
		)
	}

	// FOOTER
	body.WriteString("\n[j/k ↑/↓] перемещение   [i] ввод   [esc] normal\n")

	return box.Render(body.String())
}

func (m model) viewHelp() string {
	help := `🛠 Комбинации клавиш

Vim-режим:
  j/k или ↑/↓  — перемещение
  i            — Insert режим
  a            — добавить перерыв
  d            — удалить перерыв

Insert-режим:
  ввод текста
  Esc          — обратно в Vim

Общие:
  ?            — показать/скрыть help
  q            — выход

Формат времени: чч:мм
`

	box := helpBoxStyle.Render(help)

	return lipgloss.Place(
		80, 24, // можно потом заменить на size окна
		lipgloss.Left,
		lipgloss.Left,
		box,
	)
}

/* ---------- helpers ---------- */
func (m model) fieldCount() int {
	return 5 + len(m.breaks)*2 // +1 чекбокс
}

func (m model) cursorBreakIndex() (int, bool) {
	if m.cursor < 4 {
		return 0, false
	}

	return (m.cursor - 4) / 2, true
}

/* ---------- time logic ---------- */

func parseClock(v string) (time.Time, error) {
	p := strings.Split(v, ":")
	if len(p) != 2 {
		return time.Time{}, fmt.Errorf("bad")
	}

	h, _ := strconv.Atoi(p[0])
	m, _ := strconv.Atoi(p[1])

	return time.Date(0, 1, 1, h, m, 0, 0, time.UTC), nil
}

func parseDuration(v string) (time.Duration, error) {
	p := strings.Split(v, ":")
	if len(p) != 2 {
		return 0, fmt.Errorf("bad")
	}

	h, _ := strconv.Atoi(p[0])
	m, _ := strconv.Atoi(p[1])

	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute, nil
}

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
