package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeSavePrompt
	ModeLoadPrompt
	ModeFileList
	ModePresetList
	ModeHistory
)

type HelpState int

const (
	HelpHidden HelpState = iota
	HelpVisible
)

// StatusType определяет стиль отображения статусного сообщения.
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
type tickMsg struct{}

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

	// Config
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
	filePathInput   textinput.Model
	statusMessage   string
	statusType      StatusType
	availableFiles  []string
	allFiles        []string // все файлы для поиска
	fileListCursor  int
	fileSearchInput textinput.Model
	confirmDelete   bool
	renameInput     textinput.Model
	renaming        bool
	presetCursor    int
	historyEntries  []HistoryEntry
	historyCursor   int

	// Calculation results
	result       string
	endTimeRaw   string // время окончания без конвертации TZ (для progressInfo)
	err          string
	lastSnapshot string

	// Timer
	currentTime time.Time
	dayEnded    bool
	notified    bool

	// Services
	storage    *Storage
	calculator *Calculator
}

type Break struct {
	from textinput.Model
	to   textinput.Model
}

func (m *Model) setStatus(msg string, t StatusType) {
	m.statusMessage = msg
	m.statusType = t
}

// todayFileName возвращает имя файла сохранения для текущей даты в формате DD_MM.json.
func todayFileName() string {
	return time.Now().Format("02_01") + ".json"
}

func NewModel(saveFile string) Model {
	cfg := LoadConfig()
	initStyles(cfg)
	loc := LoadLocale(cfg.Language)

	fileInput := textinput.New()
	fileInput.Placeholder = loc.PlaceholderFile
	fileInput.SetWidth(40)

	fileSearchInput := textinput.New()
	fileSearchInput.Placeholder = "search files"
	fileSearchInput.SetWidth(30)

	renameInput := textinput.New()
	renameInput.Placeholder = "new_name.json"
	renameInput.SetWidth(40)

	workDir, _ := toAbsolutePath(cfg.WorkDir)

	if saveFile != "" {
		if absPath, err := toAbsolutePath(saveFile); err == nil {
			saveFile = absPath
		}
	} else if cfg.DefaultFile != "" {
		if absPath, err := toAbsolutePath(cfg.DefaultFile); err == nil {
			saveFile = absPath
		}
	} else {
		saveFile = filepath.Join(workDir, todayFileName())
	}

	m := Model{
		mode:            ModeNormal,
		helpState:       HelpHidden,
		cursor:          0,
		saveFile:        saveFile,
		workDir:         workDir,
		config:          cfg,
		locale:          loc,
		startTime:       createTimeInput(),
		workTime:        createTimeInput(),
		worked:          createDurationInput(),
		plan:            createDurationInput(),
		breaks:          []Break{},
		filePathInput:   fileInput,
		fileSearchInput: fileSearchInput,
		renameInput:     renameInput,
		storage:         NewStorage(saveFile),
		calculator:      NewCalculator(loc),
	}

	// Плейсхолдеры полей ввода из локали
	m.startTime.Placeholder = loc.PlaceholderTime
	m.workTime.Placeholder = loc.PlaceholderTime
	m.worked.Placeholder = loc.PlaceholderTime
	m.plan.Placeholder = loc.PlaceholderTime

	m.startTime.Focus()

	if err := os.MkdirAll(workDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not create work directory %s: %v\n", workDir, err)
	}

	if saveFile != "" {
		if _, err := os.Stat(saveFile); err == nil {
			m.loadState()
		}
	}

	return m
}

func createTimeInput() textinput.Model {
	ti := textinput.New()
	ti.SetWidth(6)
	ti.CharLimit = 5 // формат чч:мм
	return ti
}

func createDurationInput() textinput.Model {
	ti := textinput.New()
	ti.SetWidth(7)
	ti.CharLimit = 6 // формат ЧЧЧ:мм (до 999 часов)
	return ti
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(func() tea.Msg { return textinput.Blink() }, tick())
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// --- Non-key messages ---------------------------------------------------
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case clearStatusMsg:
		m.statusMessage = ""
		m.confirmQuit = false
		return m, nil

	case clipboardCopiedMsg:
		m.setStatus(m.locale.StatusCopied, StatusSuccess)
		return m, clearStatusAfter(time.Duration(m.config.UI.Timeouts.Clipboard) * time.Second)

	case tickMsg:
		m.currentTime = time.Now()
		m.checkDayEnd()
		return m, tick()
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// --- Global hotkeys (work in all modes except file prompts) -------------
	if m.mode != ModeSavePrompt && m.mode != ModeLoadPrompt && m.mode != ModeFileList && m.mode != ModePresetList && m.mode != ModeHistory {
		switch keyMsg.String() {
		case "q":
			// If help is visible, just close it
			if m.helpState == HelpVisible {
				m.helpState = HelpHidden
				return m, nil
			}
			// Dirty check: first q warns, second q quits
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

	// --- Help screen: eat all other keys ------------------------------------
	if m.helpState == HelpVisible {
		return m, nil
	}

	// --- Mode-specific handling ---------------------------------------------
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
	case ModePresetList:
		m, cmd = m.updatePresetList(keyMsg)
	case ModeHistory:
		m, cmd = m.updateHistoryMode(keyMsg)
	}

	m.recalculate()

	return m, cmd
}

func (m Model) updateNormalMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	// Any key resets the quit-confirm banner
	if m.confirmQuit {
		m.confirmQuit = false
		m.statusMessage = ""
	}

	switch msg.String() {
	case "i":
		m.mode = ModeInsert
		m.focusCurrent()

	case "j", "down":
		m.moveCursor(1)

	case "k", "up":
		m.moveCursor(-1)

	case "alt+right":
		m.jumpToNextBreak()

	case "alt+left":
		m.jumpToPrevBreak()

	case "a":
		m.addBreak()

	case "p":
		if len(m.config.Breaks) > 0 {
			m.mode = ModePresetList
			m.presetCursor = 0
		}

	case "H":
		m.enterHistoryMode()

	case "d":
		m.deleteCurrentBreak()

	case "space":
		if m.cursor == FieldAddTZ {
			m.addTZ = !m.addTZ
			m.isDirty = true
		}

	case "t":
		m.fillCurrentWithNow()

	case "x":
		m.clearCurrentField()

	case "s":
		m.enterSaveMode()

	case "o":
		m.enterFileListMode()

	case "y":
		if m.result != "" {
			return m, copyToClipboard(m.result)
		}

	default:
		// Быстрый ввод из конфига
		if value, ok := m.config.QuickInputs[msg.String()]; ok {
			m.fillCurrentWithValue(value)
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

			if strings.HasPrefix(m.statusMessage, "✅") {
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
	// --- Rename mode ---
	if m.renaming {
		return m.updateRenameMode(msg)
	}

	// --- Delete confirmation ---
	if m.confirmDelete {
		switch msg.String() {
		case "y", "enter":
			if len(m.availableFiles) > 0 && m.fileListCursor < len(m.availableFiles) {
				selectedFile := m.availableFiles[m.fileListCursor]
				filePath := filepath.Join(m.workDir, selectedFile)
				if err := os.Remove(filePath); err != nil {
					m.setStatus(fmt.Sprintf(m.locale.StatusDeleteError, err.Error()), StatusError)
				} else {
					m.setStatus(fmt.Sprintf(m.locale.StatusDeleted, selectedFile), StatusSuccess)
					m.loadAvailableFiles()
					if m.fileListCursor >= len(m.availableFiles) && m.fileListCursor > 0 {
						m.fileListCursor--
					}
				}
			}
			m.confirmDelete = false
			return m, clearStatusAfter(time.Duration(m.config.UI.Timeouts.Status) * time.Second)
		case "n", "esc":
			m.confirmDelete = false
			m.statusMessage = ""
			return m, nil
		}
		return m, nil
	}

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

	case "d":
		if len(m.availableFiles) > 0 && m.fileListCursor < len(m.availableFiles) {
			m.confirmDelete = true
			m.setStatus(fmt.Sprintf(m.locale.ConfirmDelete, m.availableFiles[m.fileListCursor]), StatusWarn)
		}

	case "r":
		if len(m.availableFiles) > 0 && m.fileListCursor < len(m.availableFiles) {
			m.renaming = true
			m.renameInput.SetValue(m.availableFiles[m.fileListCursor])
			m.renameInput.Focus()
		}

	case "/":
		m.fileSearchInput.SetValue("")
		m.fileSearchInput.Focus()

	default:
		// Обработка ввода символов для поиска
		if m.fileSearchInput.Focused() {
			var cmd tea.Cmd
			m.fileSearchInput, cmd = m.fileSearchInput.Update(msg)
			m.filterFiles()
			m.fileListCursor = 0
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) updateRenameMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		newName := m.renameInput.Value()
		if newName == "" {
			m.renaming = false
			m.renameInput.Blur()
			return m, nil
		}
		if filepath.Ext(newName) != ".json" {
			newName += ".json"
		}
		if len(m.availableFiles) > 0 && m.fileListCursor < len(m.availableFiles) {
			oldPath := filepath.Join(m.workDir, m.availableFiles[m.fileListCursor])
			newPath := filepath.Join(m.workDir, newName)
			if err := os.Rename(oldPath, newPath); err != nil {
				m.setStatus(fmt.Sprintf(m.locale.StatusRenameError, err.Error()), StatusError)
			} else {
				m.setStatus(fmt.Sprintf(m.locale.StatusRenamed, m.availableFiles[m.fileListCursor], newName), StatusSuccess)
				m.loadAvailableFiles()
				for i, f := range m.availableFiles {
					if f == newName {
						m.fileListCursor = i
						break
					}
				}
			}
		}
		m.renaming = false
		m.renameInput.Blur()
		return m, clearStatusAfter(time.Duration(m.config.UI.Timeouts.Status) * time.Second)

	case "esc":
		m.renaming = false
		m.renameInput.Blur()
		m.statusMessage = ""
		return m, nil
	}

	var cmd tea.Cmd
	m.renameInput, cmd = m.renameInput.Update(msg)
	return m, cmd
}

func (m Model) updatePresetList(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.presetCursor < len(m.config.Breaks) {
			preset := m.config.Breaks[m.presetCursor]
			m.addBreakPreset(preset)
			m.mode = ModeNormal
			m.focusCurrent()
		}
		return m, nil

	case "esc":
		m.mode = ModeNormal
		m.statusMessage = ""
		m.focusCurrent()
		return m, nil

	case "j", "down":
		if m.presetCursor < len(m.config.Breaks)-1 {
			m.presetCursor++
		}

	case "k", "up":
		if m.presetCursor > 0 {
			m.presetCursor--
		}
	}

	return m, nil
}

func (m *Model) addBreakPreset(preset BreakPreset) {
	br := Break{
		from: createTimeInput(),
		to:   createTimeInput(),
	}
	br.from.SetValue(preset.From)
	br.to.SetValue(preset.To)
	br.from.Blur()
	br.to.Blur()
	m.breaks = append(m.breaks, br)
	m.isDirty = true
}

func (m *Model) enterHistoryMode() {
	m.mode = ModeHistory
	m.historyCursor = 0
	histStorage := NewHistoryStorage(m.workDir)
	entries, err := histStorage.Load()
	if err != nil {
		m.historyEntries = []HistoryEntry{}
	} else {
		m.historyEntries = entries
	}
}

func (m Model) updateHistoryMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if len(m.historyEntries) > 0 && m.historyCursor < len(m.historyEntries) {
			entry := m.historyEntries[m.historyCursor]
			if entry.FileName != "" {
				filePath := filepath.Join(m.workDir, entry.FileName)
				if _, err := os.Stat(filePath); err == nil {
					m.saveFile = filePath
					m.storage = NewStorage(filePath)
					m.loadState()
					m.exitFileMode()
					return m, clearStatusAfter(time.Duration(m.config.UI.Timeouts.Status) * time.Second)
				}
			}
		}
		return m, nil

	case "esc", "q":
		m.mode = ModeNormal
		m.statusMessage = ""
		m.focusCurrent()
		return m, nil

	case "j", "down":
		if m.historyCursor < len(m.historyEntries)-1 {
			m.historyCursor++
		}

	case "k", "up":
		if m.historyCursor > 0 {
			m.historyCursor--
		}
	}

	return m, nil
}

// saveStateCmd выполняет сохранение и возвращает команду очистки статуса.
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
	m.recordHistory()
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
		m.filePathInput.SetValue(todayFileName())
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
		m.allFiles = []string{}
		return
	}

	type fileInfo struct {
		name    string
		modTime time.Time
	}
	files := make([]fileInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		if entry.Name() == HistoryFileName {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, fileInfo{name: entry.Name(), modTime: info.ModTime()})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.After(files[j].modTime)
	})

	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.name
	}

	m.allFiles = names
	m.availableFiles = names
}

func (m *Model) filterFiles() {
	searchTerm := strings.ToLower(m.fileSearchInput.Value())
	if searchTerm == "" {
		m.availableFiles = m.allFiles
		return
	}

	filtered := []string{}
	for _, file := range m.allFiles {
		if strings.Contains(strings.ToLower(file), searchTerm) {
			filtered = append(filtered, file)
		}
	}
	m.availableFiles = filtered
}

func (m *Model) exitFileMode() {
	m.mode = ModeNormal
	m.filePathInput.Blur()
	m.filePathInput.SetValue("")
	m.fileSearchInput.Blur()
	m.fileSearchInput.SetValue("")
	m.renameInput.Blur()
	m.renameInput.SetValue("")
	m.confirmDelete = false
	m.renaming = false
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

func (m *Model) jumpToNextBreak() {
	if len(m.breaks) == 0 {
		return
	}
	// Если уже на последнем поле, переходим к первому перерыву
	if m.cursor >= FieldBreaksStart {
		// Если это последнее поле перерыва, переходим к началу
		if m.cursor >= FieldBreaksStart+len(m.breaks)*2-1 {
			m.cursor = FieldBreaksStart
		} else {
			// Иначе переходим к следующему полю
			m.cursor++
		}
		m.blurAll()
		m.focusCurrent()
	} else {
		// Если в основном блоке, переходим к первому перерыву
		m.cursor = FieldBreaksStart
		m.blurAll()
		m.focusCurrent()
	}
}

func (m *Model) jumpToPrevBreak() {
	if m.cursor < FieldBreaksStart {
		return
	}
	// Если это первое поле перерывов, возвращаемся в основной блок
	if m.cursor == FieldBreaksStart {
		m.cursor = FieldStartTime
		m.blurAll()
		m.focusCurrent()
	} else {
		// Иначе переходим к предыдущему полю
		m.cursor--
		m.blurAll()
		m.focusCurrent()
	}
}

// fillCurrentWithNow вставляет текущее время в любое time-поле под курсором.
func (m *Model) fillCurrentWithNow() {
	if m.cursor == FieldAddTZ {
		return
	}
	field := m.getCurrentField()
	if field == nil {
		return
	}
	now := CurrentTimeInZone(m.config.InputTimezone)
	field.SetValue(now)
	m.isDirty = true
}

// fillCurrentWithValue вставляет заданное значение в поле под курсором.
func (m *Model) fillCurrentWithValue(value string) {
	if m.cursor == FieldAddTZ {
		return
	}
	field := m.getCurrentField()
	if field == nil {
		return
	}
	field.SetValue(value)
	m.isDirty = true
}

// clearCurrentField очищает значение поля под курсором.
// Не действует на чекбокс AddTZ.
func (m *Model) clearCurrentField() {
	if m.cursor == FieldAddTZ {
		return
	}
	field := m.getCurrentField()
	if field == nil {
		return
	}
	if field.Value() == "" {
		return
	}
	field.SetValue("")
	m.isDirty = true
}

// checkDayEnd проверяет, перешло ли текущее время через расчётное время окончания.
// Если день окончен — ставит флаг dayEnded и (однократно) уведомляет.
func (m *Model) checkDayEnd() {
	if m.endTimeRaw == "" {
		return
	}

	percent, _, _, ok := m.progressInfo()
	if !ok {
		return
	}

	if percent >= 1.0 {
		m.dayEnded = true
		if !m.notified {
			m.notified = true
			m.setStatus(m.locale.StatusDayEnded, StatusSuccess)
			fmt.Print("\a")
		}
	} else {
		m.dayEnded = false
		m.notified = false
	}
}

func (m *Model) recalculate() {
	// Snapshot всех входных значений для пропуска лишних вычислений
	var sb strings.Builder
	sb.WriteString(m.startTime.Value())
	sb.WriteByte('|')
	sb.WriteString(m.workTime.Value())
	sb.WriteByte('|')
	sb.WriteString(m.worked.Value())
	sb.WriteByte('|')
	sb.WriteString(m.plan.Value())
	sb.WriteByte('|')
	if m.addTZ {
		sb.WriteByte('1')
	} else {
		sb.WriteByte('0')
	}
	for _, br := range m.breaks {
		sb.WriteByte('|')
		sb.WriteString(br.from.Value())
		sb.WriteByte('-')
		sb.WriteString(br.to.Value())
	}
	snap := sb.String()
	if snap == m.lastSnapshot {
		return
	}
	m.lastSnapshot = snap

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

	// endTimeRaw — время окончания без конвертации TZ, для progressInfo/checkDayEnd.
	// Считаем даже если конвертация TZ дала ошибку.
	rawInput := input
	rawInput.AddTZ = false
	rawResult, rawErr := m.calculator.Calculate(rawInput)
	if rawErr == nil {
		m.endTimeRaw = rawResult
	} else {
		m.endTimeRaw = ""
	}

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
	m.addTZ = data.AddTZ

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

// clearStatusAfter возвращает команду, которая через d очистит statusMessage.
func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

// copyToClipboard копирует текст через OSC 52 (работает в большинстве терминалов).
func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		encoded := base64.StdEncoding.EncodeToString([]byte(text))
		f, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
		if err == nil {
			_, _ = fmt.Fprintf(f, "\033]52;c;%s\007", encoded)
			_ = f.Close()
		}
		return clipboardCopiedMsg{}
	}
}
