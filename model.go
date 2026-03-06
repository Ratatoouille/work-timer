package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

// StatusType используется вместо проверки emoji-префиксов в View.
type StatusType int

const (
	StatusNeutral StatusType = iota
	StatusSuccess
	StatusError
	StatusWarn
)

const (
	FieldStartTime = iota
	FieldWorkTime
	FieldWorked
	FieldPlan
	FieldAddTZ
	FieldBreaksStart
)

const DefaultWorkDir = "~/work_timer"

// --- Message types ----------------------------------------------------------

type clearStatusMsg struct{}
type clipboardCopiedMsg struct{}

// ----------------------------------------------------------------------------

type Model struct {
	mode      Mode
	helpState HelpState
	cursor    int
	saveFile  string
	workDir   string

	// Terminal size
	width  int
	height int

	// Dirty tracking
	isDirty     bool
	confirmQuit bool

	// Config & locale
	config Config
	locale Locale

	// Input fields
	startTime textinput.Model
	workTime  textinput.Model
	worked    textinput.Model
	plan      textinput.Model
	addTZ     bool
	breaks    []Break

	// File operation fields
	filePathInput  textinput.Model
	statusMessage  string
	statusType     StatusType
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

// setStatus устанавливает статусное сообщение с типом — вместо проверки emoji в View.
func (m *Model) setStatus(msg string, t StatusType) {
	m.statusMessage = msg
	m.statusType = t
}

// NewModel — основной конструктор. Читает конфиг с диска.
func NewModel(saveFile string) Model {
	cfg := LoadConfig()
	loc := LoadLocale(cfg.Language)
	return newModelWithConfig(saveFile, cfg, loc)
}

// NewModelForTesting создаёт модель с дефолтным конфигом без обращения к диску.
// Используется в тестах вместо NewModel.
func NewModelForTesting(saveFile string) Model {
	return newModelWithConfig(saveFile, defaultConfig(), localeEN)
}

// newModelWithConfig — внутренний конструктор, принимает готовые cfg и loc.
func newModelWithConfig(saveFile string, cfg Config, loc Locale) Model {
	initStyles(cfg)

	fileInput := textinput.New()
	fileInput.Placeholder = loc.PlaceholderFile
	fileInput.Width = 40

	workDir, _ := toAbsolutePath(cfg.WorkDir)

	// Если аргумент не задан — пробуем default_file из конфига
	if saveFile == "" && cfg.DefaultFile != "" {
		saveFile = cfg.DefaultFile
	}

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
		config:        cfg,
		locale:        loc,
		startTime:     createTimeInput(loc.PlaceholderTime),
		workTime:      createTimeInput(loc.PlaceholderTime),
		worked:        createTimeInput(loc.PlaceholderTime),
		plan:          createTimeInput(loc.PlaceholderTime),
		breaks:        []Break{},
		filePathInput: fileInput,
		storage:       NewStorage(saveFile),
		calculator:    NewCalculator(loc),
	}

	m.startTime.Focus()

	os.MkdirAll(workDir, 0o755)

	if saveFile != "" {
		m.loadState()
	}

	return m
}

func createTimeInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Width = 6
	return ti
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case clearStatusMsg:
		m.statusMessage = ""
		m.statusType = StatusNeutral
		m.confirmQuit = false
		return m, nil

	case clipboardCopiedMsg:
		m.setStatus(m.locale.StatusCopied, StatusSuccess)
		return m, clearStatusAfter(time.Duration(m.config.UI.Timeouts.Clipboard) * time.Second)
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// --- Global hotkeys -----------------------------------------------------
	if m.mode != ModeSavePrompt && m.mode != ModeLoadPrompt && m.mode != ModeFileList {
		switch keyMsg.String() {
		case "q":
			if m.helpState == HelpVisible {
				m.helpState = HelpHidden
				return m, nil
			}
			if m.isDirty && !m.confirmQuit {
				m.confirmQuit = true
				m.setStatus(m.locale.StatusUnsaved, StatusWarn)
				return m, clearStatusAfter(time.Duration(m.config.UI.Timeouts.Warning) * time.Second)
			}
			return m, tea.Quit

		case "?":
			m.helpState = toggleHelpState(m.helpState)
			return m, nil

		case "ctrl+s":
			if m.saveFile != "" {
				return m, m.saveStateCmd()
			}
			m.enterSaveMode()
			return m, nil

		case "ctrl+o":
			m.enterFileListMode()
			return m, nil
		}
	}

	if m.helpState == HelpVisible {
		return m, nil
	}

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
	if m.confirmQuit {
		m.confirmQuit = false
		m.statusMessage = ""
		m.statusType = StatusNeutral
	}

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
		if m.cursor == FieldAddTZ {
			m.addTZ = !m.addTZ
			m.isDirty = true
		}

	case "t":
		// Вставляем текущее время в поле "Начало"
		if m.cursor == FieldStartTime {
			now := CurrentTimeInZone(m.config.InputTimezone)
			m.startTime.SetValue(now)
			m.isDirty = true
		}

	case "s":
		m.enterSaveMode()

	case "o":
		m.enterFileListMode()

	case "y":
		if m.result != "" {
			return m, copyToClipboard(m.result)
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
		before := field.Value()
		*field, cmd = field.Update(msg)
		if field.Value() != before {
			m.isDirty = true
		}
	}

	return m, cmd
}

func (m Model) updateSavePrompt(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		fileName := m.filePathInput.Value()
		if fileName != "" {
			var filePath string
			if filepath.IsAbs(fileName) {
				filePath = fileName
			} else {
				filePath = filepath.Join(m.workDir, fileName)
			}

			absPath, err := toAbsolutePath(filePath)
			if err != nil {
				m.setStatus(fmt.Sprintf(m.locale.StatusPathError, err.Error()), StatusError)
				return m, nil
			}

			m.saveFile = absPath
			m.storage = NewStorage(absPath)
			return m, m.saveStateCmd()
		}
		return m, nil

	case "esc":
		m.setStatus(m.locale.StatusSaveCancelled, StatusNeutral)
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
				m.setStatus(fmt.Sprintf(m.locale.StatusPathError, err.Error()), StatusError)
				return m, nil
			}

			m.saveFile = absPath
			m.storage = NewStorage(absPath)
			m.loadState()

			if m.statusType == StatusSuccess {
				m.exitFileMode()
				return m, clearStatusAfter(time.Duration(m.config.UI.Timeouts.Status) * time.Second)
			}
		}
		return m, nil

	case "esc":
		m.setStatus(m.locale.StatusLoadCancelled, StatusNeutral)
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
			return m, clearStatusAfter(time.Duration(m.config.UI.Timeouts.Status) * time.Second)
		}
		return m, nil

	case "esc":
		m.setStatus(m.locale.StatusLoadCancelled, StatusNeutral)
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
		m.mode = ModeLoadPrompt
		m.filePathInput.SetValue("")
		m.filePathInput.Focus()
	}

	return m, nil
}

func (m *Model) saveStateCmd() tea.Cmd {
	data := SaveData{
		StartTime: m.startTime.Value(),
		WorkTime:  m.workTime.Value(),
		Worked:    m.worked.Value(),
		Plan:      m.plan.Value(),
		AddTZ:     m.addTZ,
		Breaks:    m.getBreaksForSave(),
	}

	if err := m.storage.Save(data); err != nil {
		m.setStatus(fmt.Sprintf(m.locale.StatusSaveError, err.Error()), StatusError)
		return clearStatusAfter(time.Duration(m.config.UI.Timeouts.Warning) * time.Second)
	}

	m.isDirty = false
	m.setStatus(fmt.Sprintf(m.locale.StatusSavedAs, filepath.Base(m.saveFile)), StatusSuccess)

	if m.mode == ModeSavePrompt {
		m.exitFileMode()
	}

	return clearStatusAfter(time.Duration(m.config.UI.Timeouts.Status) * time.Second)
}

func (m *Model) enterSaveMode() {
	m.mode = ModeSavePrompt
	if m.saveFile != "" {
		m.filePathInput.SetValue(filepath.Base(m.saveFile))
	} else {
		m.filePathInput.SetValue("")
	}
	m.filePathInput.Focus()
	m.statusMessage = ""
	m.statusType = StatusNeutral
}

func (m *Model) enterFileListMode() {
	m.mode = ModeFileList
	m.fileListCursor = 0
	m.statusMessage = ""
	m.statusType = StatusNeutral
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
	from := createTimeInput(m.locale.PlaceholderTime)
	from.Blur()

	to := createTimeInput(m.locale.PlaceholderTime)
	to.Blur()

	m.breaks = append(m.breaks, Break{from: from, to: to})
	m.isDirty = true
}

func (m *Model) deleteCurrentBreak() {
	if idx, ok := m.currentBreakIndex(); ok && idx < len(m.breaks) {
		m.breaks = append(m.breaks[:idx], m.breaks[idx+1:]...)
		if m.cursor > FieldAddTZ {
			m.cursor--
		}
		m.isDirty = true
	}
}

func (m *Model) recalculate() {
	input := CalculationInput{
		StartTime:     m.startTime.Value(),
		WorkTime:      m.workTime.Value(),
		Worked:        m.worked.Value(),
		Plan:          m.plan.Value(),
		AddTZ:         m.addTZ,
		InputTimezone: m.config.InputTimezone,
		Timezone:      m.config.Timezone,
		Breaks:        m.getBreaksData(),
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

func (m *Model) loadState() {
	data, err := m.storage.Load()
	if err != nil {
		m.setStatus(fmt.Sprintf(m.locale.StatusLoadError, err.Error()), StatusError)
		return
	}

	m.startTime.SetValue(data.StartTime)
	m.workTime.SetValue(data.WorkTime)
	m.worked.SetValue(data.Worked)
	m.plan.SetValue(data.Plan)
	// Поддержка старых файлов с полем "add_four"
	m.addTZ = data.AddTZ || data.LegacyAddFour

	m.breaks = make([]Break, 0, len(data.Breaks))
	for _, bd := range data.Breaks {
		from := createTimeInput(m.locale.PlaceholderTime)
		from.SetValue(bd.From)
		from.Blur()

		to := createTimeInput(m.locale.PlaceholderTime)
		to.SetValue(bd.To)
		to.Blur()

		m.breaks = append(m.breaks, Break{from: from, to: to})
	}

	m.isDirty = false
	m.setStatus(fmt.Sprintf(m.locale.StatusLoadedFrom, filepath.Base(m.saveFile)), StatusSuccess)
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
	case FieldAddTZ:
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

// --- Commands ---------------------------------------------------------------

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		encoded := base64.StdEncoding.EncodeToString([]byte(text))
		f, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
		if err == nil {
			fmt.Fprintf(f, "\033]52;c;%s\007", encoded)
			f.Close()
		}
		return clipboardCopiedMsg{}
	}
}
