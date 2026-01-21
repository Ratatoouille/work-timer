package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
		mode:      ModeInsert,
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
	m.result = end.Format("15:04")
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

	// Ставим фокус только на текущее поле
	switch m.cursor {
	case 0:
		m.startTime.Focus()
	case 1:
		m.workTime.Focus()
	case 2:
		m.worked.Focus()
	case 3:
		m.plan.Focus()
	default:
		idx := (m.cursor - 4) / 2

		if (m.cursor-4)%2 == 0 {
			m.breaks[idx].from.Focus()
		} else {
			m.breaks[idx].to.Focus()
		}
	}
}

func (m model) View() string {
	if m.help == HelpVisible {
		return m.viewHelp()
	}

	var b strings.Builder

	modeStr := "Insert"
	if m.mode == ModeVim {
		modeStr = "Vim"
	}

	b.WriteString(fmt.Sprintf("🕒 Режим: %s\n? — помощь   q — выход\n\n", modeStr))
	b.WriteString(fmt.Sprintf("Начало: %s\n", m.startTime.View()))
	b.WriteString(fmt.Sprintf("Оставшееся время: %s\n", m.workTime.View()))
	b.WriteString(fmt.Sprintf("Отработано: %s\n", m.worked.View()))
	b.WriteString(fmt.Sprintf("План: %s\n", m.plan.View()))

	for n, br := range m.breaks {
		b.WriteString(fmt.Sprintf("Перерыв %d — ушёл: %s\n", n+1, br.from.View()))
		b.WriteString(fmt.Sprintf("Перерыв %d — вернулся: %s\n", n+1, br.to.View()))
	}

	if m.err != "" {
		b.WriteString("❌ " + m.err + "\n")
	}
	if m.result != "" {
		b.WriteString("✅ Время окончания: " + m.result + "\n")
	}

	return b.String()
}

func (m model) viewHelp() string {
	help := `
🛠 Комбинации клавиш

Vim-режим:
  j/k или стрелки вверх/вниз — перемещение
  i — перейти в Insert режим (редактирование)
  a — добавить перерыв
  d — удалить текущий перерыв

Insert-режим:
  редактирование полей
  Esc — вернуться в Vim режим

Общие:
  ? — показать/скрыть эту подсказку
  q — выйти

Формат времени всегда чч:мм
`
	return help
}

/* ---------- helpers ---------- */

func (m model) fieldCount() int {
	return 4 + len(m.breaks)*2
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
