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

type Break struct {
	from textinput.Model
	to   textinput.Model
}

type model struct {
	mode      Mode
	startTime textinput.Model
	workTime  textinput.Model
	breaks    []Break
	cursor    int
	result    string
	err       string
}

func initialModel() model {
	start := textinput.New()
	start.Placeholder = "чч:мм"
	start.Focus()

	work := textinput.New()
	work.Placeholder = "чч:мм"

	return model{
		mode:      ModeInsert,
		startTime: start,
		workTime:  work,
		breaks:    []Break{},
		cursor:    0,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) recalc() {
	start, err1 := parseClock(m.startTime.Value())
	work, err2 := parseDuration(m.workTime.Value())
	if err1 != nil || err2 != nil {
		m.err = "Неверный формат времени"
		m.result = ""
		return
	}

	totalBreak := time.Duration(0)
	for _, br := range m.breaks {
		f, err1 := parseClock(br.from.Value())
		t, err2 := parseClock(br.to.Value())
		if err1 != nil || err2 != nil {
			m.err = "Неверный формат перерыва"
			m.result = ""
			return
		}
		totalBreak += t.Sub(f)
	}

	end := start.Add(work).Add(totalBreak)
	m.result = end.Format("15:04")
	m.err = ""
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {

		case ModeVim:
			switch msg.String() {
			case "i":
				m.mode = ModeInsert
				if m.cursor == 0 {
					m.startTime.Focus()
				} else if m.cursor == 1 {
					m.workTime.Focus()
				} else {
					idx := (m.cursor - 2) / 2
					if (m.cursor-2)%2 == 0 {
						m.breaks[idx].from.Focus()
					} else {
						m.breaks[idx].to.Focus()
					}
				}
			case "j":
				if m.cursor < m.fieldCount()-1 {
					m.cursor++
				}
			case "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "a":
				// добавить перерыв
				br1 := textinput.New()
				br1.Placeholder = "чч:мм"
				br2 := textinput.New()
				br2.Placeholder = "чч:мм"
				m.breaks = append(m.breaks, Break{from: br1, to: br2})
			case "d":
				idx, ok := m.cursorBreakIndex()
				if ok {
					m.breaks = append(m.breaks[:idx], m.breaks[idx+1:]...)
					if m.cursor > 1 {
						m.cursor--
					}
				}
			}

		case ModeInsert:
			var cmd tea.Cmd
			if m.cursor == 0 {
				m.startTime, cmd = m.startTime.Update(msg)
			} else if m.cursor == 1 {
				m.workTime, cmd = m.workTime.Update(msg)
			} else {
				idx := (m.cursor - 2) / 2
				if (m.cursor-2)%2 == 0 {
					m.breaks[idx].from, cmd = m.breaks[idx].from.Update(msg)
				} else {
					m.breaks[idx].to, cmd = m.breaks[idx].to.Update(msg)
				}
			}
			cmds = append(cmds, cmd)
			if msg.String() == "esc" {
				m.mode = ModeVim
			}
		}
	}

	m.recalc()
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var b strings.Builder

	modeStr := "Insert"
	if m.mode == ModeVim {
		modeStr = "Vim"
	}

	b.WriteString(fmt.Sprintf("🕒 Режим: %s\n\n", modeStr))

	cursor := func(i int) string {
		if i == m.cursor && m.mode == ModeVim {
			return "👉 "
		}
		return "   "
	}

	b.WriteString(cursor(0) + "Начало: " + m.startTime.View() + "\n")
	b.WriteString(cursor(1) + "Рабочее время: " + m.workTime.View() + "\n")

	for n, br := range m.breaks {
		b.WriteString(cursor(2+n*2) + fmt.Sprintf("Перерыв %d — ушёл: %s\n", n+1, br.from.View()))
		b.WriteString(cursor(3+n*2) + fmt.Sprintf("Перерыв %d — вернулся: %s\n", n+1, br.to.View()))
	}

	b.WriteString("\nИнструкции:\n")
	b.WriteString("Vim: j/k перемещение, i — редактирование, a — добавить перерыв, d — удалить\n")
	b.WriteString("Insert: редактирование, Esc — вернуться в Vim\n")
	b.WriteString("q — выход\n\n")

	if m.err != "" {
		b.WriteString("❌ " + m.err + "\n")
	}
	if m.result != "" {
		b.WriteString("✅ Время окончания: " + m.result + "\n")
	}

	return b.String()
}

/* ---------- helpers ---------- */

func (m model) fieldCount() int {
	return 2 + len(m.breaks)*2
}

func (m model) cursorBreakIndex() (int, bool) {
	if m.cursor < 2 {
		return 0, false
	}
	return (m.cursor - 2) / 2, true
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
	if err := p.Start(); err != nil {
		panic(err)
	}
}
