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
	SectionResult string

	// --- Поля ввода ---
	FieldStart         string
	FieldMode1Label    string
	FieldRemainingTime string
	FieldMode2Label    string
	FieldWorked        string
	FieldPlan          string
	CheckboxAddTZ      string // когда timezone не задан
	CheckboxShowIn     string // "Показать в %s"

	// --- Перерывы ---
	BreakLeft  string // "Перерыв %d — ушёл:"
	BreakRight string // "Перерыв %d — вернулся:"

	// --- Результат ---
	ResultLabel string
	Remaining   string

	// --- Панель управления ---
	CtrlNav    string
	CtrlEdit   string
	CtrlSave   string
	CtrlOpen   string
	CtrlHelp   string
	CtrlQuit   string

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
	ConfirmDelete   string
	RenamePrompt    string

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
	SectionResult: "Результат",

	FieldStart:         "Начало",
	FieldMode1Label:    "режим 1: оставшееся время",
	FieldRemainingTime: "Оставшееся время",
	FieldMode2Label:    "режим 2: отработано / план",
	FieldWorked:        "Отработано",
	FieldPlan:          "План",
	CheckboxAddTZ:      "Конвертировать timezone",
	CheckboxShowIn:     "Показать в %s",

	BreakLeft:  "Перерыв %d — ушел",
	BreakRight: "Перерыв %d — вернулся",

	ResultLabel: "Время окончания",
	Remaining:   "ост.",

	CtrlNav:    "навиг.",
	CtrlEdit:   "редакт.",
	CtrlSave:   "сохранить",
	CtrlOpen:   "открыть",
	CtrlHelp:   "справка",
	CtrlQuit:   "выход",

	SaveTitle:       "💾 Сохранить как",
	SaveFolder:      "Папка: %s",
	SaveHint:        "[Enter] сохранить  [Esc] отмена",
	LoadTitle:       "📂 Загрузить из файла",
	LoadHint:        "[Enter] загрузить  [Esc] отмена",
	FileListTitle:   "📂 Выберите файл",
	FileListEmpty:   "Нет сохранённых файлов",
	FileListHintNew: "[n] создать новый  [Esc] отмена",
	FileListHint:    "[j/k ↑↓] навигация  [Enter] выбрать  [/] поиск  [d] удалить  [r] переименовать  [n] новый  [Esc] отмена",
	ConfirmDelete:   "⚠  Удалить %s? [y] да  [n/Esc] отмена",
	RenamePrompt:    "✏  Новое имя:",

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
  d            — удалить перерыв
  space        — переключить чекбокс
  y            — скопировать результат
  s            — сохранить
  o            — открыть список файлов

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
	SectionResult: "Result",

	FieldStart:         "Start",
	FieldMode1Label:    "mode 1: remaining time",
	FieldRemainingTime: "Remaining time",
	FieldMode2Label:    "mode 2: worked / plan",
	FieldWorked:        "Worked",
	FieldPlan:          "Plan",
	CheckboxAddTZ:      "Convert timezone",
	CheckboxShowIn:     "Show in %s",

	BreakLeft:  "Break %d — left",
	BreakRight: "Break %d — returned",

	ResultLabel: "End time",
	Remaining:   "left",

	CtrlNav:    "nav",
	CtrlEdit:   "edit",
	CtrlSave:   "save",
	CtrlOpen:   "open",
	CtrlHelp:   "help",
	CtrlQuit:   "quit",

	SaveTitle:       "💾 Save as",
	SaveFolder:      "Folder: %s",
	SaveHint:        "[Enter] save  [Esc] cancel",
	LoadTitle:       "📂 Load from file",
	LoadHint:        "[Enter] load  [Esc] cancel",
	FileListTitle:   "📂 Select file",
	FileListEmpty:   "No saved files",
	FileListHintNew: "[n] create new  [Esc] cancel",
	FileListHint:    "[j/k ↑↓] navigate  [Enter] select  [/] search  [d] delete  [r] rename  [n] new  [Esc] cancel",
	ConfirmDelete:   "⚠  Delete %s? [y] yes  [n/Esc] cancel",
	RenamePrompt:    "✏  New name:",

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
  d             — delete break
  space         — toggle checkbox
  y             — copy result
  s             — save
  o             — open file list

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
