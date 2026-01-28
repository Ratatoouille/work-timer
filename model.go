package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeSavePrompt
	ModeLoadPrompt
	ModeFileList
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

const DefaultWorkDir = "~/work_timer"

type Model struct {
	mode      Mode
	helpState HelpState
	cursor    int
	saveFile  string
	workDir   string

	// Input fields
	startTime textinput.Model
	workTime  textinput.Model
	worked    textinput.Model
	plan      textinput.Model
	addFour   bool
	breaks    []Break

	// File operation fields
	filePathInput  textinput.Model
	statusMessage  string
	availableFiles []string
	fileListCursor int

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
	fileInput := textinput.New()
	fileInput.Placeholder = "имя_файла.json"
	fileInput.Width = 40

	// Получаем абсолютный путь рабочей директории
	workDir, _ := toAbsolutePath(DefaultWorkDir)

	// Конвертируем путь в абсолютный, если он передан
	if saveFile != "" {
		if absPath, err := toAbsolutePath(saveFile); err == nil {
			saveFile = absPath
		}
	}

	m := Model{
		mode:          ModeNormal,
		helpState:     HelpHidden,
		cursor:        0,
		saveFile:      saveFile,
		workDir:       workDir,
		startTime:     createTimeInput(),
		workTime:      createTimeInput(),
		worked:        createTimeInput(),
		plan:          createTimeInput(),
		breaks:        []Break{},
		filePathInput: fileInput,
		storage:       NewStorage(saveFile),
		calculator:    NewCalculator(),
	}

	m.startTime.Focus()

	// Создаем рабочую директорию если её нет
	os.MkdirAll(workDir, 0o755)

	if saveFile != "" {
		m.loadState()
	}

	return m
}

func createTimeInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "чч:мм"
	ti.Width = 6

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
	if m.mode != ModeSavePrompt && m.mode != ModeLoadPrompt && m.mode != ModeFileList {
		switch keyMsg.String() {
		case "q":
			return m, tea.Quit
		case "?":
			m.helpState = toggleHelpState(m.helpState)

			return m, nil
		case "ctrl+s":
			if m.saveFile != "" {
				m.saveState()
			} else {
				m.enterSaveMode()
			}

			return m, nil
		case "ctrl+o":
			m.enterFileListMode()

			return m, nil
		}
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
	case ModeSavePrompt:
		m, cmd = m.updateSavePrompt(keyMsg)
	case ModeLoadPrompt:
		m, cmd = m.updateLoadPrompt(keyMsg)
	case ModeFileList:
		m, cmd = m.updateFileList(keyMsg)
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

	case "s":
		m.enterSaveMode()

	case "o":
		m.enterFileListMode()
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

func (m Model) updateSavePrompt(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		fileName := m.filePathInput.Value()
		if fileName != "" {
			// Если путь не абсолютный, сохраняем в рабочую директорию
			var filePath string
			if filepath.IsAbs(fileName) {
				filePath = fileName
			} else {
				filePath = filepath.Join(m.workDir, fileName)
			}

			absPath, err := toAbsolutePath(filePath)
			if err != nil {
				m.statusMessage = "❌ Ошибка пути: " + err.Error()

				return m, nil
			}

			m.saveFile = absPath
			m.storage = NewStorage(absPath)
			m.saveState()

			if strings.HasPrefix(m.statusMessage, "✅") {
				m.exitFileMode()
			}
		}

		return m, nil

	case "esc":
		m.statusMessage = "Сохранение отменено"
		m.exitFileMode()

		return m, nil
	}

	m.filePathInput, cmd = m.filePathInput.Update(msg)

	return m, cmd
}

func (m Model) updateLoadPrompt(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		filePath := m.filePathInput.Value()
		if filePath != "" {
			absPath, err := toAbsolutePath(filePath)
			if err != nil {
				m.statusMessage = "❌ Ошибка пути: " + err.Error()

				return m, nil
			}

			m.saveFile = absPath
			m.storage = NewStorage(absPath)
			m.loadState()

			if strings.HasPrefix(m.statusMessage, "✅") {
				m.exitFileMode()
			}
		}

		return m, nil

	case "esc":
		m.statusMessage = "Загрузка отменена"
		m.exitFileMode()

		return m, nil
	}

	m.filePathInput, cmd = m.filePathInput.Update(msg)

	return m, cmd
}

func (m Model) updateFileList(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if len(m.availableFiles) > 0 && m.fileListCursor < len(m.availableFiles) {
			selectedFile := m.availableFiles[m.fileListCursor]
			filePath := filepath.Join(m.workDir, selectedFile)
			m.saveFile = filePath
			m.storage = NewStorage(filePath)
			m.loadState()
			m.exitFileMode()
		}

		return m, nil

	case "esc":
		m.statusMessage = "Загрузка отменена"
		m.exitFileMode()

		return m, nil

	case "j", "down":
		if m.fileListCursor < len(m.availableFiles)-1 {
			m.fileListCursor++
		}

	case "k", "up":
		if m.fileListCursor > 0 {
			m.fileListCursor--
		}

	case "n":
		// Создать новый файл
		m.mode = ModeLoadPrompt
		m.filePathInput.SetValue("")
		m.filePathInput.Focus()
	}

	return m, nil
}

func (m *Model) enterSaveMode() {
	m.mode = ModeSavePrompt
	// Если файл уже есть, показываем только имя
	if m.saveFile != "" {
		m.filePathInput.SetValue(filepath.Base(m.saveFile))
	} else {
		m.filePathInput.SetValue("")
	}
	m.filePathInput.Focus()
	m.statusMessage = ""
}

func (m *Model) enterFileListMode() {
	m.mode = ModeFileList
	m.fileListCursor = 0
	m.statusMessage = ""
	m.loadAvailableFiles()
}

func (m *Model) loadAvailableFiles() {
	entries, err := os.ReadDir(m.workDir)
	if err != nil {
		m.availableFiles = []string{}

		return
	}

	files := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			files = append(files, entry.Name())
		}
	}

	m.availableFiles = files
}

func (m *Model) exitFileMode() {
	m.mode = ModeNormal
	m.filePathInput.Blur()
	m.filePathInput.SetValue("")
	m.focusCurrent()
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
		m.statusMessage = "❌ Ошибка сохранения: " + err.Error()
	} else {
		m.statusMessage = "✅ Сохранено в " + filepath.Base(m.saveFile)
	}
}

func (m *Model) loadState() {
	data, err := m.storage.Load()
	if err != nil {
		m.statusMessage = "❌ Ошибка загрузки: " + err.Error()

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

	m.statusMessage = "✅ Загружено из " + filepath.Base(m.saveFile)
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

func toAbsolutePath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		path = filepath.Join(homeDir, path[2:])
	}

	if filepath.IsAbs(path) {
		return path, nil
	}

	return filepath.Abs(path)
}
