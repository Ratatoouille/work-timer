package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
)

type HelpState int

const (
	HelpHidden HelpState = iota
	HelpVisible
)

const (
	FieldStartTime = iota
	FieldWorkTime
	FieldWorked
	FieldPlan
	FieldAddFour
	FieldBreaksStart
)

type Model struct {
	mode      Mode
	helpState HelpState
	cursor    int
	saveFile  string

	// Input fields
	startTime textinput.Model
	workTime  textinput.Model
	worked    textinput.Model
	plan      textinput.Model
	addFour   bool
	breaks    []Break

	// Calculation results
	result string
	err    string

	// Services
	storage    *Storage
	calculator *Calculator
}

type Break struct {
	from textinput.Model
	to   textinput.Model
}

func NewModel(saveFile string) Model {
	m := Model{
		mode:       ModeNormal,
		helpState:  HelpHidden,
		cursor:     0,
		saveFile:   saveFile,
		startTime:  createTimeInput(),
		workTime:   createTimeInput(),
		worked:     createTimeInput(),
		plan:       createTimeInput(),
		breaks:     []Break{},
		storage:    NewStorage(saveFile),
		calculator: NewCalculator(),
	}

	m.startTime.Focus()

	if saveFile != "" {
		m.loadState()
	}

	return m
}

func createTimeInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "чч:мм"
	ti.Width = 5

	return ti
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// Global hotkeys
	switch keyMsg.String() {
	case "q":
		return m, tea.Quit
	case "?":
		m.helpState = toggleHelpState(m.helpState)

		return m, nil
	case "ctrl+s":
		m.saveState()

		return m, nil
	}

	// Help screen handling
	if m.helpState == HelpVisible {
		return m, nil
	}

	// Mode-specific handling
	var cmd tea.Cmd
	switch m.mode {
	case ModeNormal:
		m, cmd = m.updateNormalMode(keyMsg)
	case ModeInsert:
		m, cmd = m.updateInsertMode(keyMsg)
	}

	m.recalculate()

	return m, cmd
}

func (m Model) updateNormalMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "i":
		m.mode = ModeInsert
		m.focusCurrent()

	case "j", "down":
		m.moveCursor(1)

	case "k", "up":
		m.moveCursor(-1)

	case "a":
		m.addBreak()

	case "d":
		m.deleteCurrentBreak()

	case " ":
		if m.cursor == FieldAddFour {
			m.addFour = !m.addFour
		}
	}

	return m, nil
}

func (m Model) updateInsertMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.mode = ModeNormal

		m.blurAll()
		m.focusCurrent()

		return m, nil
	}

	var cmd tea.Cmd

	field := m.getCurrentField()
	if field != nil {
		*field, cmd = field.Update(msg)
	}

	return m, cmd
}

func (m *Model) moveCursor(delta int) {
	newCursor := m.cursor + delta

	if newCursor >= 0 && newCursor < m.totalFields() {
		m.cursor = newCursor
		m.blurAll()
		m.focusCurrent()
	}
}

func (m *Model) addBreak() {
	br := Break{
		from: createTimeInput(),
		to:   createTimeInput(),
	}

	br.from.Blur()
	br.to.Blur()
	m.breaks = append(m.breaks, br)
}

func (m *Model) deleteCurrentBreak() {
	if idx, ok := m.currentBreakIndex(); ok && idx < len(m.breaks) {
		m.breaks = append(m.breaks[:idx], m.breaks[idx+1:]...)
		if m.cursor > FieldAddFour {
			m.cursor--
		}
	}
}

func (m *Model) recalculate() {
	input := CalculationInput{
		StartTime: m.startTime.Value(),
		WorkTime:  m.workTime.Value(),
		Worked:    m.worked.Value(),
		Plan:      m.plan.Value(),
		AddFour:   m.addFour,
		Breaks:    m.getBreaksData(),
	}

	result, err := m.calculator.Calculate(input)
	if err != nil {
		m.err = err.Error()
		m.result = ""

		return
	}

	m.result = result
	m.err = ""
}

func (m Model) getBreaksData() []BreakTime {
	breaks := make([]BreakTime, 0, len(m.breaks))
	for _, br := range m.breaks {
		breaks = append(breaks, BreakTime{
			From: br.from.Value(),
			To:   br.to.Value(),
		})
	}

	return breaks
}

func (m *Model) saveState() {
	data := SaveData{
		StartTime: m.startTime.Value(),
		WorkTime:  m.workTime.Value(),
		Worked:    m.worked.Value(),
		Plan:      m.plan.Value(),
		AddFour:   m.addFour,
		Breaks:    m.getBreaksForSave(),
	}

	if err := m.storage.Save(data); err != nil {
		m.err = "Ошибка сохранения: " + err.Error()
	}
}

func (m *Model) loadState() {
	data, err := m.storage.Load()
	if err != nil {
		m.err = "Ошибка загрузки: " + err.Error()
		return
	}

	m.startTime.SetValue(data.StartTime)
	m.workTime.SetValue(data.WorkTime)
	m.worked.SetValue(data.Worked)
	m.plan.SetValue(data.Plan)
	m.addFour = data.AddFour

	m.breaks = make([]Break, 0, len(data.Breaks))
	for _, bd := range data.Breaks {
		from := createTimeInput()
		from.SetValue(bd.From)
		from.Blur()

		to := createTimeInput()
		to.SetValue(bd.To)
		to.Blur()

		m.breaks = append(m.breaks, Break{from: from, to: to})
	}
}

func (m Model) getBreaksForSave() []BreakData {
	breaks := make([]BreakData, 0, len(m.breaks))
	for _, br := range m.breaks {
		breaks = append(breaks, BreakData{
			From: br.from.Value(),
			To:   br.to.Value(),
		})
	}

	return breaks
}

func (m *Model) getCurrentField() *textinput.Model {
	switch m.cursor {
	case FieldStartTime:
		return &m.startTime
	case FieldWorkTime:
		return &m.workTime
	case FieldWorked:
		return &m.worked
	case FieldPlan:
		return &m.plan
	case FieldAddFour:
		return nil
	default:
		if idx := (m.cursor - FieldBreaksStart) / 2; idx < len(m.breaks) {
			if (m.cursor-FieldBreaksStart)%2 == 0 {
				return &m.breaks[idx].from
			}

			return &m.breaks[idx].to
		}
	}

	return nil
}

func (m *Model) focusCurrent() {
	field := m.getCurrentField()
	if field != nil {
		field.Focus()
	}
}

func (m *Model) blurAll() {
	m.startTime.Blur()
	m.startTime.CursorEnd()
	m.workTime.Blur()
	m.workTime.CursorEnd()
	m.worked.Blur()
	m.worked.CursorEnd()
	m.plan.Blur()
	m.plan.CursorEnd()

	for i := range m.breaks {
		m.breaks[i].from.Blur()
		m.breaks[i].from.CursorEnd()
		m.breaks[i].to.Blur()
		m.breaks[i].to.CursorEnd()
	}
}

func (m Model) totalFields() int {
	return FieldBreaksStart + len(m.breaks)*2
}

func (m Model) currentBreakIndex() (int, bool) {
	if m.cursor < FieldBreaksStart {
		return 0, false
	}

	return (m.cursor - FieldBreaksStart) / 2, true
}

func toggleHelpState(state HelpState) HelpState {
	if state == HelpHidden {
		return HelpVisible
	}

	return HelpHidden
}
