package main

// Locale содержит все пользовательские строки интерфейса.
type Locale struct {
	// --- Заголовок ---
	ModeNormal     string
	ModeInsert     string
	NoFileSelected string

	// --- Секции ---
	SectionMain   string
	SectionBreaks string
	SectionParams string
	SectionResult string

	// --- Поля ввода ---
	FieldStart         string
	FieldEnd           string
	FieldMode1Label    string
	FieldRemainingTime string
	FieldMode2Label    string
	FieldWorked        string
	FieldPlan          string
	CheckboxAddTZ      string // когда timezone не задан
	CheckboxShowIn     string // "Показать в %s"

	// --- Большой блок "оставшееся время" ---
	RemainingLabel string

	// --- Статус таймера ---
	StateRunning  string
	StatePaused   string
	StateDone     string
	StateOvertime string

	// --- Параметры: режим ---
	StatMode    string
	StatMode1   string
	StatMode2   string
	StatModeOff string
	StatOn      string
	StatOff     string

	// --- Перерывы ---
	BreakWord      string
	DurFormatHours string // "%dч %02dм" — часы и минуты
	DurFormatMins  string // "%dм" — только минуты

	// --- Перерывы ---
	BreakLeft  string // "Перерыв %d — ушёл:"
	BreakRight string // "Перерыв %d — вернулся:"

	// --- Результат ---
	ResultLabel  string
	Remaining    string
	RemainingCap string // "Осталось" (капитализированный/полный вариант для секции Результат)
	Status       string // "Статус"

	// --- Панель управления ---
	CtrlNav    string
	CtrlEdit   string
	CtrlSave   string
	CtrlOpen   string
	CtrlHelp   string
	CtrlQuit   string
	CtrlDelete string
	CtrlRename string
	CtrlNew    string
	CtrlCancel string
	CtrlSearch string

	// --- Промпты ---
	SaveTitle       string
	SaveFolder      string
	SaveHint        string
	LoadTitle       string
	LoadHint        string
	FileListTitle   string
	FileListEmpty   string
	FileListHintNew string
	FileListHint    string
	FileListHint2   string
	FileSearchLabel string
	ConfirmDelete   string
	RenamePrompt    string
	PresetTitle     string
	PresetEmpty     string
	PresetHint      string
	HistoryTitle    string
	HistoryEmpty    string
	HistoryHint     string
	HistoryColDate  string
	HistoryColStart string
	HistoryColEnd   string
	HistoryColBreak string
	HistoryColSaved string

	// --- Статусные сообщения ---
	StatusCopied        string
	StatusUnsaved       string
	StatusSaveCancelled string
	StatusLoadCancelled string
	StatusSavedAs       string // "✅ Сохранено: %s"
	StatusLoadedFrom    string // "✅ Загружено: %s"
	StatusSaveError     string // "❌ Ошибка сохранения: %s"
	StatusLoadError     string // "❌ Ошибка загрузки: %s"
	StatusPathError     string // "❌ Ошибка пути: %s"
	StatusDeleted       string // "✅ Удалено: %s"
	StatusDeleteError   string // "❌ Ошибка удаления: %s"
	StatusRenamed       string // "✅ Переименовано: %s → %s"
	StatusRenameError   string // "❌ Ошибка переименования: %s"
	StatusDayEnded      string

	// --- Ошибки калькулятора ---
	ErrInvalidStartTime    string
	ErrInvalidWorked       string
	ErrInvalidPlan         string
	ErrWorkedExceedsPlan   string // "отработано (%s) больше плана (%s)"
	ErrNoTimeInput         string
	ErrBreakInvalidFrom    string // "перерыв %d: ..."
	ErrBreakInvalidTo      string
	ErrBreakEndBeforeStart string
	ErrBreakBeforeWorkday  string // "перерыв %d: начало (%s) раньше начала рабочего дня (%s)"
	ErrTimezoneNotSet      string
	ErrTimezoneInvalid     string // "неверный timezone в конфиге: %q"
	ErrInvalidTimeFormat   string
	ErrInvalidHours        string
	ErrInvalidMinutes      string
	ErrInvalidDuration     string

	// --- Плейсхолдеры полей ввода ---
	PlaceholderFile string
	PlaceholderTime string
	HelpText        string
}

var localeRU = Locale{
	ModeNormal:     "NORMAL",
	ModeInsert:     "INSERT",
	NoFileSelected: "○ файл не выбран",

	SectionMain:   "Основные параметры",
	SectionBreaks: "Перерывы",
	SectionParams: "Параметры",
	SectionResult: "Результат",

	FieldStart:         "Начало",
	FieldEnd:           "Окончание",
	FieldMode1Label:    "режим 1: оставшееся время",
	FieldRemainingTime: "Оставшееся время",
	FieldMode2Label:    "режим 2: отработано / план",
	FieldWorked:        "Отработано",
	FieldPlan:          "План",
	CheckboxAddTZ:      "Конвертировать timezone",
	CheckboxShowIn:     "Показать в %s",

	RemainingLabel: "ОСТАЛОСЬ",

	StateRunning:  "РАБОТАЕМ",
	StatePaused:   "ПЕРЕРЫВ",
	StateDone:     "ГОТОВО",
	StateOvertime: "ПЕРЕРАБОТКА",

	StatMode:    "Режим",
	StatMode1:   "Режим 1",
	StatMode2:   "Режим 2",
	StatModeOff: "Режим",
	StatOn:      "ВКЛ",
	StatOff:     "ВЫКЛ",

	BreakWord:      "Перерыв",
	DurFormatHours: "%dч %02dм",
	DurFormatMins:  "%dм",

	BreakLeft:  "Перерыв %d — начало",
	BreakRight: "Перерыв %d — конец",

	ResultLabel:  "Время окончания",
	Remaining:    "ост.",
	RemainingCap: "Осталось",
	Status:       "Статус",

	CtrlNav:    "навиг.",
	CtrlEdit:   "редакт.",
	CtrlSave:   "сохранить",
	CtrlOpen:   "открыть",
	CtrlHelp:   "справка",
	CtrlQuit:   "выход",
	CtrlDelete: "удалить",
	CtrlRename: "переимен.",
	CtrlNew:    "новый",
	CtrlCancel: "отмена",
	CtrlSearch: "поиск",

	SaveTitle:       "Сохранить как",
	SaveFolder:      "Папка: %s",
	SaveHint:        "[Enter] сохранить  [Esc] отмена",
	LoadTitle:       "Загрузить из файла",
	LoadHint:        "[Enter] загрузить  [Esc] отмена",
	FileListTitle:   "Выберите файл",
	FileListEmpty:   "Нет сохранённых файлов",
	FileListHintNew: "[n] создать новый  [Esc] отмена",
	FileListHint:    "[j/k ↑↓] навигация  [Enter] выбрать",
	FileListHint2:   "[/] поиск  [d] удалить  [r] переименовать  [n] новый  [Esc] отмена",
	FileSearchLabel: "Search",
	ConfirmDelete:   "⚠  Удалить %s? [y] да  [n/Esc] отмена",
	RenamePrompt:    "Новое имя:",
	PresetTitle:     "Добавить перерыв из шаблона",
	PresetEmpty:     "Шаблоны перерывов не настроены",
	PresetHint:      "[j/k ↑↓] выбрать  [Enter] добавить  [Esc] отмена",
	HistoryTitle:    "История рабочих дней",
	HistoryEmpty:    "История пуста",
	HistoryHint:     "[j/k ↑↓] навигация  [Enter] открыть файл  [Esc] отмена",
	HistoryColDate:  "Дата",
	HistoryColStart: "Начало",
	HistoryColEnd:   "Конец",
	HistoryColBreak: "Перерывы",
	HistoryColSaved: "Сохр.",

	StatusCopied:        "✅ Скопировано в буфер обмена",
	StatusUnsaved:       "⚠  Есть несохранённые изменения. Нажмите q ещё раз для выхода",
	StatusSaveCancelled: "Сохранение отменено",
	StatusLoadCancelled: "Загрузка отменена",
	StatusSavedAs:       "✅ Сохранено: %s",
	StatusLoadedFrom:    "✅ Загружено: %s",
	StatusSaveError:     "❌ Ошибка сохранения: %s",
	StatusLoadError:     "❌ Ошибка загрузки: %s",
	StatusPathError:     "❌ Ошибка пути: %s",
	StatusDeleted:       "✅ Удалено: %s",
	StatusDeleteError:   "❌ Ошибка удаления: %s",
	StatusRenamed:       "✅ Переименовано: %s → %s",
	StatusRenameError:   "❌ Ошибка переименования: %s",
	StatusDayEnded:      "🎉 Рабочий день окончен!",

	ErrInvalidStartTime:    "неверное время начала",
	ErrInvalidWorked:       "неверный формат 'отработано'",
	ErrInvalidPlan:         "неверный формат 'план'",
	ErrWorkedExceedsPlan:   "отработано (%s) больше плана (%s)",
	ErrNoTimeInput:         "введите либо оставшееся время, либо отработано/план",
	ErrBreakInvalidFrom:    "перерыв %d: неверный формат времени начала",
	ErrBreakInvalidTo:      "перерыв %d: неверный формат времени конца",
	ErrBreakEndBeforeStart: "перерыв %d: время конца раньше или равно началу",
	ErrBreakBeforeWorkday:  "перерыв %d: начало (%s) раньше начала рабочего дня (%s)",
	ErrTimezoneNotSet:      "timezone не задан в конфиге (~/.config/work_timer/config.toml)",
	ErrTimezoneInvalid:     "неверный timezone в конфиге: %q",
	ErrInvalidTimeFormat:   "неверный формат времени",
	ErrInvalidHours:        "неверные часы",
	ErrInvalidMinutes:      "неверные минуты",
	ErrInvalidDuration:     "неверный формат длительности",

	PlaceholderFile: "имя_файла.json",
	PlaceholderTime: "чч:мм",

	HelpText: `🛠  Комбинации клавиш

Normal-режим:
  j/k или ↑/↓  — перемещение
  i            — Insert режим
  a            — добавить перерыв
  p            — добавить перерыв из шаблона
  d            — удалить перерыв
  H            — история рабочих дней
  space        — переключить чекбокс
  y            — скопировать результат
  s            — сохранить
  o            — открыть список файлов

Режимы:
  Режим 1      — заполни «Оставшееся время» (старт + перерывы = конец)
  Режим 2      — заполни «Отработано» и «План», очисти «Оставшееся время»
  x            — очистить текущее поле

Insert-режим:
  ввод текста
  Esc          — обратно в Normal

Выбор файла:
  j/k или ↑/↓  — навигация по списку
  /             — поиск файлов
  Enter        — загрузить выбранный файл
  d            — удалить файл
  r            — переименовать файл
  n            — создать новый файл
  Esc          — отмена

Общие:
  ?            — показать/скрыть справку
  q            — закрыть справку / выход
  Ctrl+S       — быстрое сохранение
  Ctrl+O       — открыть список файлов

Конфиг: %s

Формат времени: чч:мм

Рабочая папка: %s
`,
}

var localeEN = Locale{
	ModeNormal:     "NORMAL",
	ModeInsert:     "INSERT",
	NoFileSelected: "○ no file selected",

	SectionMain:   "Main parameters",
	SectionBreaks: "Breaks",
	SectionParams: "Options",
	SectionResult: "Result",

	FieldStart:         "Start",
	FieldEnd:           "End",
	FieldMode1Label:    "mode 1: remaining time",
	FieldRemainingTime: "Remaining time",
	FieldMode2Label:    "mode 2: worked / plan",
	FieldWorked:        "Worked",
	FieldPlan:          "Plan",
	CheckboxAddTZ:      "Convert timezone",
	CheckboxShowIn:     "Show in %s",

	RemainingLabel: "REMAINING",

	StateRunning:  "RUNNING",
	StatePaused:   "PAUSED",
	StateDone:     "DONE",
	StateOvertime: "OVERTIME",

	StatMode:    "Mode",
	StatMode1:   "Mode 1",
	StatMode2:   "Mode 2",
	StatModeOff: "Mode",
	StatOn:      "ON",
	StatOff:     "OFF",

	BreakWord:      "Break",
	DurFormatHours: "%dh %02dm",
	DurFormatMins:  "%dm",

	BreakLeft:  "Break %d — start",
	BreakRight: "Break %d — end",

	ResultLabel:  "End time",
	Remaining:    "left",
	RemainingCap: "Remaining",
	Status:       "Status",

	CtrlNav:    "nav",
	CtrlEdit:   "edit",
	CtrlSave:   "save",
	CtrlOpen:   "open",
	CtrlHelp:   "help",
	CtrlQuit:   "quit",
	CtrlDelete: "delete",
	CtrlRename: "rename",
	CtrlNew:    "new",
	CtrlCancel: "cancel",
	CtrlSearch: "search",

	SaveTitle:       "Save as",
	SaveFolder:      "Folder: %s",
	SaveHint:        "[Enter] save  [Esc] cancel",
	LoadTitle:       "Load from file",
	LoadHint:        "[Enter] load  [Esc] cancel",
	FileListTitle:   "Select file",
	FileListEmpty:   "No saved files",
	FileListHintNew: "[n] create new  [Esc] cancel",
	FileListHint:    "[j/k ↑↓] navigate  [Enter] select",
	FileListHint2:   "[/] search  [d] delete  [r] rename  [n] new  [Esc] cancel",
	FileSearchLabel: "Search",
	ConfirmDelete:   "⚠  Delete %s? [y] yes  [n/Esc] cancel",
	RenamePrompt:    "New name:",
	PresetTitle:     "Add break from preset",
	PresetEmpty:     "No break presets configured",
	PresetHint:      "[j/k ↑↓] select  [Enter] add  [Esc] cancel",
	HistoryTitle:    "Work day history",
	HistoryEmpty:    "History is empty",
	HistoryHint:     "[j/k ↑↓] navigate  [Enter] open file  [Esc] cancel",
	HistoryColDate:  "Date",
	HistoryColStart: "Start",
	HistoryColEnd:   "End",
	HistoryColBreak: "Breaks",
	HistoryColSaved: "Saved",

	StatusCopied:        "✅ Copied to clipboard",
	StatusUnsaved:       "⚠  Unsaved changes. Press q again to quit",
	StatusSaveCancelled: "Save cancelled",
	StatusLoadCancelled: "Load cancelled",
	StatusSavedAs:       "✅ Saved: %s",
	StatusLoadedFrom:    "✅ Loaded: %s",
	StatusSaveError:     "❌ Save error: %s",
	StatusLoadError:     "❌ Load error: %s",
	StatusPathError:     "❌ Path error: %s",
	StatusDeleted:       "✅ Deleted: %s",
	StatusDeleteError:   "❌ Delete error: %s",
	StatusRenamed:       "✅ Renamed: %s → %s",
	StatusRenameError:   "❌ Rename error: %s",
	StatusDayEnded:      "🎉 Work day is over!",

	ErrInvalidStartTime:    "invalid start time",
	ErrInvalidWorked:       "invalid format for 'worked'",
	ErrInvalidPlan:         "invalid format for 'plan'",
	ErrWorkedExceedsPlan:   "worked (%s) exceeds plan (%s)",
	ErrNoTimeInput:         "enter either remaining time or worked/plan",
	ErrBreakInvalidFrom:    "break %d: invalid start time format",
	ErrBreakInvalidTo:      "break %d: invalid end time format",
	ErrBreakEndBeforeStart: "break %d: end time is before or equal to start",
	ErrBreakBeforeWorkday:  "break %d: start (%s) is before workday start (%s)",
	ErrTimezoneNotSet:      "timezone not set in config (~/.config/work_timer/config.toml)",
	ErrTimezoneInvalid:     "invalid timezone in config: %q",
	ErrInvalidTimeFormat:   "invalid time format",
	ErrInvalidHours:        "invalid hours",
	ErrInvalidMinutes:      "invalid minutes",
	ErrInvalidDuration:     "invalid duration format",

	PlaceholderFile: "filename.json",
	PlaceholderTime: "hh:mm",

	HelpText: `🛠  Key bindings

Normal mode:
  j/k or ↑/↓   — navigate
  i             — Insert mode
  a             — add break
  p             — add break from preset
  d             — delete break
  H             — work day history
  space         — toggle checkbox
  y             — copy result
  s             — save
  o             — open file list

Modes:
  Mode 1       — fill "Remaining time" (start + breaks = end)
  Mode 2       — fill "Worked" and "Plan", clear "Remaining time"
  x            — clear current field

Insert mode:
  type text
  Esc           — back to Normal

File selection:
  j/k or ↑/↓   — navigate list
  /             — search files
  Enter         — load selected file
  d             — delete file
  r             — rename file
  n             — create new file
  Esc           — cancel

General:
  ?             — show/hide help
  q             — close help / quit
  Ctrl+S        — quick save
  Ctrl+O        — open file list

Config: %s

Time format: HH:MM

Work folder: %s
`,
}

// LoadLocale возвращает локаль по коду языка. Неизвестный код → русский.
func LoadLocale(lang string) Locale {
	switch lang {
	case "en":
		return localeEN
	default:
		return localeRU
	}
}
