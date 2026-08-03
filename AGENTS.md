# Agent Guidelines for work-timer

## Project Overview

TUI work timer application built with Go and Bubble Tea v2. Calculates work end times based on start time, work duration, and breaks with timezone conversion support.

## Essential Commands

```bash
go build -o work-timer .        # Build
go test ./...                    # Run tests
go run .                         # Run application
go run . today.json              # Run with specific save file
```

## Project Structure

```
work-timer/
├── main.go           # Entry point, CLI parsing, tea.Program initialization
├── model.go          # Core model: state machine, Update logic, field management
├── view.go           # Rendering: header, fields, breaks, result, controls
├── calculator.go     # Time calculations, break validation, timezone conversion
├── storage.go        # JSON file I/O for save/load operations
├── config.go         # TOML config loading, defaults, timezone helpers
├── locale.go         # i18n: Russian/English UI strings
└── *_test.go         # Unit tests for calculator, model, storage, config, locale
```

## Architecture

**State Machine**: Application operates in modes:
- `ModeNormal` - Navigation and command keys (j/k, i, a, d, s, o, y, q)
- `ModeInsert` - Text input for time fields
- `ModeSavePrompt` / `ModeLoadPrompt` - File path input
- `ModeFileList` - Browse saved files

**Update Flow**: `model.go:Update()` → mode-specific handlers → `recalculate()` → `View()`

**Key Files**:
- `model.go:121-177` - Model initialization with config and locale
- `model.go:202-287` - Update loop handling all messages and mode dispatch
- `calculator.go:34-58` - Core calculation: start + work_time + breaks = end_time
- `config.go:41-118` - Config struct with defaults, loaded from `~/.config/work_timer/config.toml`

## Code Patterns

**Field Indexing** (model.go:43-49):
```go
const (
    FieldStartTime = iota  // 0
    FieldWorkTime          // 1
    FieldWorked            // 2
    FieldPlan              // 3
    FieldAddTZ             // 4
    FieldBreaksStart       // 5, breaks use pairs: [5,6], [7,8], ...
)
```

**Time Parsing** (calculator.go:152-189):
- Format: `HH:MM` (hours 0-23, minutes 0-59)
- Duration parsing allows 999+ hours for `worked`/`plan` fields

**Dirty Tracking**: `isDirty` flag tracks unsaved changes, triggers quit confirmation

**Input Validation**: `isInvalidTimeValue()` and `isInvalidDurationValue()` in `view.go` validate format on-the-fly

## Configuration

Config path: `~/.config/work_timer/config.toml`

Key settings:
- `work_dir` - Save file location (default: `~/work_timer`)
- `default_file` - File to open on startup without CLI args (default: `""`)
- `language` - "ru" or "en"
- `input_timezone` - IANA timezone for input times
- `timezone` - Target timezone for conversion (empty = no conversion)
- `ui.colors` - ANSI color codes (0-255) or hex
- `ui.timeouts` - Status message timeouts in seconds

## Testing Patterns

Test files follow `*_test.go` convention with table-driven tests:

```go
func TestXxx(t *testing.T) {
    tests := []struct {
        name    string
        input   Type
        want    Type
        wantErr bool
    }{...}
    
    calc := NewCalculator(localeEN)
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := calc.Method(tt.input)
            if (err != nil) != tt.wantErr { ... }
            if got != tt.want { ... }
        })
    }
}
```

**Test coverage**: calculator (time parsing, duration, breaks), model (dirty tracking, field navigation, save/load), storage (JSON I/O), config (defaults, loading), locale (non-empty fields, RU/EN differ)

## Gotchas

1. **Tilde expansion**: Config paths use `~/` which must be expanded via `toAbsolutePath()` (model.go:862-877)

2. **Break validation**: Breaks must have `from < to` and cannot be before work start time (calculator.go:86-116)

3. **Two calculation modes**:
   - Mode 1: `start_time + work_time + breaks = end_time`
   - Mode 2: `start_time + (plan - worked) + breaks = end_time`

4. **Timezone conversion**: Only applies if `addTZ=true` AND `timezone` is set in config. Uses real date for accurate DST handling (calculator.go:118-150)

5. **Snapshot optimization**: `recalculate()` skips if input snapshot unchanged (model.go:687-735)

6. **Clipboard copy**: Uses OSC 52 escape sequence via `/dev/tty` (model.go:889-901)

7. **Config defaults**: Zero values are filled from defaults after TOML decode (config.go:83-118)

8. **Progress info**: `progressInfo()` in `view.go` calculates elapsed/remaining time for display, handling overnight shifts by trying multiple day offsets for the work window

## File Format

Save files are JSON (storage.go:14-27):
```json
{
  "start_time": "06:44",
  "work_time": "09:00",
  "worked": "",
  "plan": "",
  "add_tz": false,
  "breaks": [{"from": "12:00", "to": "13:00"}]
}
```

## Key Dependencies

- `charm.land/bubbletea/v2` - TUI framework (v2)
- `charm.land/bubbles/v2` - Text input components (v2)
- `charm.land/lipgloss/v2` - Styling (v2)
- `BurntSushi/toml` - Config parsing
