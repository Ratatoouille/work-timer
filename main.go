package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Break struct {
	from string
	to   string
}

type model struct {
	startTime string
	workTime  string
	breaks    []Break

	cursor  int
	editing bool
	input   string

	result string
	err    string
}

func initialModel() model {
	return model{
		breaks: []Break{},
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		// ---------- navigation ----------
		case "up":
			if !m.editing && m.cursor > 0 {
				m.cursor--
			}

		case "down":
			if !m.editing && m.cursor < m.fieldCount()-1 {
				m.cursor++
			}

		// ---------- edit mode ----------
		case "enter":
			if !m.editing {
				m.editing = true
				m.input = m.getFieldValue(m.cursor)
			} else {
				m.setFieldValue(m.cursor, m.input)
				m.editing = false
				m.input = ""
				m.recalc()
			}

		case "esc":
			m.editing = false
			m.input = ""

		case "backspace":
			if m.editing && len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}

		default:
			if m.editing && len(msg.String()) == 1 {
				m.input += msg.String()
			}

		// ---------- breaks ----------
		case "a":
			if !m.editing {
				m.breaks = append(m.breaks, Break{})
				m.recalc()
			}

		case "d":
			if !m.editing {
				idx, ok := m.cursorBreakIndex()
				if ok {
					m.breaks = append(m.breaks[:idx], m.breaks[idx+1:]...)
					if m.cursor > 0 {
						m.cursor--
					}
					m.recalc()
				}
			}
		}
	}

	return m, nil
}

func (m *model) recalc() {
	end, err := calcEndTime(m.startTime, m.workTime, m.breaks)
	if err != nil {
		m.err = err.Error()
		m.result = ""
		return
	}
	m.err = ""
	m.result = end
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("🕒 Рабочее время\n\n")

	draw := func(i int, label, value string) {
		cursor := "   "
		if i == m.cursor {
			cursor = "👉 "
		}
		if m.editing && i == m.cursor {
			b.WriteString(fmt.Sprintf("%s%s: %s_\n", cursor, label, m.input))
		} else {
			b.WriteString(fmt.Sprintf("%s%s: %s\n", cursor, label, value))
		}
	}

	i := 0
	draw(i, "Начало", m.startTime)
	i++
	draw(i, "Работать", m.workTime)
	i++

	for n, br := range m.breaks {
		draw(i, fmt.Sprintf("Перерыв %d — ушёл", n+1), br.from)
		i++
		draw(i, fmt.Sprintf("Перерыв %d — вернулся", n+1), br.to)
		i++
	}

	b.WriteString("\n⬆⬇ — навигация   Enter — редактировать\n")
	b.WriteString("a — добавить перерыв   d — удалить\n")
	b.WriteString("Esc — отмена   q — выход\n\n")

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

func (m model) getFieldValue(i int) string {
	if i == 0 {
		return m.startTime
	}
	if i == 1 {
		return m.workTime
	}

	i -= 2
	br := i / 2
	if i%2 == 0 {
		return m.breaks[br].from
	}
	return m.breaks[br].to
}

func (m *model) setFieldValue(i int, v string) {
	if i == 0 {
		m.startTime = v
		return
	}
	if i == 1 {
		m.workTime = v
		return
	}

	i -= 2
	br := i / 2
	if i%2 == 0 {
		m.breaks[br].from = v
	} else {
		m.breaks[br].to = v
	}
}

func (m model) cursorBreakIndex() (int, bool) {
	if m.cursor < 2 {
		return 0, false
	}
	return (m.cursor - 2) / 2, true
}

/* ---------- time logic ---------- */

func calcEndTime(start, work string, breaks []Break) (string, error) {
	st, err := parseClock(start)
	if err != nil {
		return "", fmt.Errorf("начало")
	}
	w, err := parseDuration(work)
	if err != nil {
		return "", fmt.Errorf("рабочее время")
	}

	totalBreak := time.Duration(0)
	for _, br := range breaks {
		if br.from == "" || br.to == "" {
			continue
		}
		f, _ := parseClock(br.from)
		t, _ := parseClock(br.to)
		totalBreak += t.Sub(f)
	}

	end := st.Add(w).Add(totalBreak)
	return end.Format("15:04"), nil
}

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
