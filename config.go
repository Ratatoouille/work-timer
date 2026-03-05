package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

const ConfigPath = "~/.config/work_timer/config.toml"

type Config struct {
	WorkDir       string   `toml:"work_dir"`
	Language      string   `toml:"language"`       // "ru" (default) или "en"
	InputTimezone string   `toml:"input_timezone"`
	Timezone      string   `toml:"timezone"`
	UI            ConfigUI `toml:"ui"`
}

type ConfigUI struct {
	LabelWidth int            `toml:"label_width"`
	Colors     ConfigColors   `toml:"colors"`
	Timeouts   ConfigTimeouts `toml:"timeouts"`
}

type ConfigColors struct {
	Accent string `toml:"accent"`
	Result string `toml:"result"`
	Break  string `toml:"break"`
	Warn   string `toml:"warn"`
}

type ConfigTimeouts struct {
	Clipboard int `toml:"clipboard"`
	Status    int `toml:"status"`
	Warning   int `toml:"warning"`
}

func defaultConfig() Config {
	return Config{
		WorkDir:       DefaultWorkDir,
		Language:      "ru",
		InputTimezone: "Europe/Moscow",
		Timezone:      "",
		UI: ConfigUI{
			LabelWidth: 22,
			Colors: ConfigColors{
				Accent: "12",
				Result: "14",
				Break:  "13",
				Warn:   "11",
			},
			Timeouts: ConfigTimeouts{
				Clipboard: 2,
				Status:    3,
				Warning:   4,
			},
		},
	}
}

func LoadConfig() Config {
	cfg := defaultConfig()

	absPath, err := toAbsolutePath(ConfigPath)
	if err != nil {
		return cfg
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		_ = writeDefaultConfig(absPath)
		return cfg
	}

	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return defaultConfig()
	}

	// Заполняем нулевые значения дефолтами
	def := defaultConfig()
	if cfg.UI.LabelWidth == 0 {
		cfg.UI.LabelWidth = def.UI.LabelWidth
	}
	if cfg.UI.Colors.Accent == "" {
		cfg.UI.Colors.Accent = def.UI.Colors.Accent
	}
	if cfg.UI.Colors.Result == "" {
		cfg.UI.Colors.Result = def.UI.Colors.Result
	}
	if cfg.UI.Colors.Break == "" {
		cfg.UI.Colors.Break = def.UI.Colors.Break
	}
	if cfg.UI.Colors.Warn == "" {
		cfg.UI.Colors.Warn = def.UI.Colors.Warn
	}
	if cfg.UI.Timeouts.Clipboard == 0 {
		cfg.UI.Timeouts.Clipboard = def.UI.Timeouts.Clipboard
	}
	if cfg.UI.Timeouts.Status == 0 {
		cfg.UI.Timeouts.Status = def.UI.Timeouts.Status
	}
	if cfg.UI.Timeouts.Warning == 0 {
		cfg.UI.Timeouts.Warning = def.UI.Timeouts.Warning
	}
	if cfg.WorkDir == "" {
		cfg.WorkDir = def.WorkDir
	}
	if cfg.Language == "" {
		cfg.Language = def.Language
	}
	if cfg.InputTimezone == "" {
		cfg.InputTimezone = def.InputTimezone
	}

	return cfg
}

func writeDefaultConfig(absPath string) error {
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return err
	}

	content := `# Work Timer — конфиг
# https://en.wikipedia.org/wiki/List_of_tz_database_time_zones

# Рабочая папка для сохранений
work_dir = "~/work_timer"

# Язык интерфейса: "ru" или "en"
language = "ru"

# Часовой пояс в котором вводится время (начало, перерывы и т.д.)
input_timezone = "Europe/Moscow"

# Целевой часовой пояс для конвертации результата.
# Оставьте пустым чтобы не конвертировать.
# Примеры: "Asia/Krasnoyarsk", "Europe/Moscow", "America/New_York"
timezone = ""

[ui]
# Ширина колонки лейблов (символов)
label_width = 22

[ui.colors]
# Цвета задаются номерами ANSI (0–255) или hex "#RRGGBB"
accent = "12"   # синий  — заголовки, рамки активного поля
result = "14"   # cyan   — блок результата
break  = "13"   # маджента — секция перерывов
warn   = "11"   # жёлтый — предупреждения, грязная точка

[ui.timeouts]
# Время (секунды) до автоматического скрытия статусных сообщений
clipboard = 2   # "скопировано в буфер"
status    = 3   # сохранение / загрузка
warning   = 4   # предупреждения и ошибки
`

	return os.WriteFile(absPath, []byte(content), 0o644)
}

// TimezoneLocation загружает *time.Location по имени из конфига.
func TimezoneLocation(timezone string) *time.Location {
	if timezone == "" {
		return nil
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil
	}

	return loc
}

// TimezoneLabel возвращает аббревиатуру зоны (напр. "KRAT") или "".
func TimezoneLabel(timezone string) string {
	loc := TimezoneLocation(timezone)
	if loc == nil {
		return ""
	}

	name, _ := time.Now().In(loc).Zone()
	return name
}
