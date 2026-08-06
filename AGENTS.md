# Agent Guidelines for work-timer

## Project Overview

TUI work timer application built with Go and Bubble Tea v2. Calculates work end times based on start time, work duration, and breaks with timezone conversion support. Interface is bilingual (Russian/English). All source files are in the package `main` at the repository root — there are no subdirectories of Go code.

## Essential Commands

```bash
make build              # Build binary → ./work-timer
make run                # go run .
make test               # go test ./...
make fmt                # go fmt ./...
make vet                # go vet ./...
make lint               # golangci-lint run ./... (skips if not installed)
make clean              # rm binary + dist/
make release            # goreleaser release --clean --snapshot (requires goreleaser)
go run .                # Run app; opens DD_MM.json for today (auto-created name)
go run . today.json     # Run with a specific save file
go run . --version      # Print version (also -v)
go run . --help         # Print usage (also -h)
```

`make build` extracts the version from `const version` in `main.go`. GoReleaser injects the release version via `-X main.version=...` ldflag (see `.goreleaser.yaml`); for local builds the const is the source of truth.

## Project Structure

All files live at the repo root in `package main`:

```
work-timer/
├── main.go           # Entry point, CLI flag parsing (--version/--help/[file]), tea.Program init
├── model.go          # Core Model, state machine, Update loop, file/history/preset handlers
├── view.go           # Rendering: header, fields, breaks, result, controls, progress bar, list views
├── calculator.go     # Time/duration parsing, break validation, TZ conversion, core calc
├── storage.go        # JSON I/O for save files (single-file load/save)
├── history.go        # history.json journal of past work days (append/upsert by date, sorted desc)
├── config.go         # TOML config loading, defaults, BreakPreset, timezone helpers
├── locale.go         # i18n: Russian/English UI strings (Locale struct, not external files)
└── *_test.go         # Unit tests for each module (table-driven)
```

`model.go` (~1200 lines) is the central hub — most behavior lives there. `view.go` (~820 lines) is large because it renders every mode. When making changes, expect most edits to touch `model.go` and `view.go` together.

## Architecture

### Bubble Tea v2 architecture

Standard Elm-architecture: `NewModel` (model.go:137) builds the initial state; `Update` (model.go) handles `tea.KeyMsg`, `tea.WindowSizeMsg`, and internal messages (`tickMsg`, `clearStatusMsg`, `clipboardCopiedMsg`); `View` (view.go) renders based on `m.mode`. After most input changes, `recalculate()` is called before the next render.

### State machine (Mode)

Defined in `model.go:16-26`:

- `ModeNormal` — Navigation and command keys (j/k, i, a, p, d, t, H, y, s, o, /, ?, q)
- `ModeInsert` — Text input for the currently focused field
- `ModeSavePrompt` / `ModeLoadPrompt` — File path text input
- `ModeFileList` — Browse saved files (with search via `/`, delete with `d`, rename with `r`)
- `ModePresetList` — Pick a break preset from `config.toml` `[[breaks]]`
- `ModeHistory` — Browse `history.json` entries

`updateNormal`, `updateInsert`, `updateFileList`, `updatePresetList`, etc. dispatch from `Update`. The tick loop drives `progressInfo()` and end-of-day notification.

### Update flow

`Update()` → mode-specific handler → (on input change) `recalculate()` → `View()`.

`recalculate()` uses a `lastSnapshot` string of all inputs to skip recomputation when nothing changed (model.go ~687-735). If you add a new input field, update the snapshot or recalcs will be stale.

### Two calculation modes

Determined by which fields are filled (calculator.go:34-58):

- Mode 1: `start_time + work_time + breaks = end_time`
- Mode 2: `start_time + (plan - worked) + breaks = end_time`

Only one mode should be filled at a time; the UI does not enforce this but the calc branches on whether `worked`/`plan` are populated.

## Code Patterns

### Field indexing (model.go:45-52)

```go
const (
    FieldStartTime = iota  // 0
    FieldWorkTime          // 1
    FieldWorked            // 2
    FieldPlan              // 3
    FieldAddTZ             // 4 — checkbox, not a textinput
    FieldBreaksStart       // 5 — breaks use pairs: [5,6], [7,8], ...
)
```

`FieldAddTZ` is a boolean checkbox, not a `textinput.Model`. Breaks are pairs of `textinput.Model` stored in `m.breaks []Break`; their cursor indices are computed as `FieldBreaksStart + 2*i` / `+1`.

### Input creation

`createTimeInput()` and `createDurationInput()` (model.go) construct textinputs with appropriate validators. Time fields accept `HH:MM` (0-23 / 0-59); duration fields (`worked`, `plan`) allow up to 999h. On-the-fly validation is done by `isInvalidTimeValue()` / `isInvalidDurationValue()` in `view.go`, which turn the field red on bad input.

### Status messages

`m.setStatus(msg, StatusType)` sets `statusMessage` + `statusType` (Neutral/Success/Error/Warn). Timed clear is driven by `clearStatusMsg` and timeouts from `[ui.timeouts]` in config.

### Dirty tracking

`m.isDirty` is set on any input change; an unsaved quit requires confirmation (`confirmQuit`). Saving clears the flag.

### History

On save, `recordHistory()` (history.go:91) writes/updates an entry in `<work_dir>/history.json` keyed by today's date (`YYYY-MM-DD`). Entries are sorted newest-first. `history.json` is separate from the per-day save files.

## Configuration

Config path: `~/.config/work_timer/config.toml` (constant `ConfigPath` in config.go). Auto-created on first run.

Top-level keys:
- `work_dir` — Save file location (default `~/work_timer`). Tilde-expanded via `toAbsolutePath()`.
- `default_file` — File to open on startup when no CLI arg given (default `""`). Takes priority over the auto-generated `DD_MM.json` name.
- `language` — `"ru"` or `"en"`.
- `input_timezone` — IANA tz for input times (start/breaks).
- `timezone` — Target tz for result conversion (empty = no conversion).

`[[breaks]]` array — break presets, selected via `p` key in Normal mode:
```toml
[[breaks]]
name = "lunch"
from = "12:00"
to = "13:00"
```

`[quick_inputs]` map — keys `1`-`9` insert a value into the current field in Normal mode.

`[ui.colors]` — `accent`, `result`, `break`, `warn`; values are ANSI 0-255 or hex `#RRGGBB`. Applied via `initStyles()`.

`[ui.timeouts]` — `clipboard`, `status`, `warning` (seconds).

Config defaults: zero values are filled from defaults after TOML decode (config.go:83-118). Do not rely on a key being present.

## File Formats

### Save files (storage.go)

JSON in `work_dir`, one per day. Default name `DD_MM.json` (e.g. `04_08.json`) when no `default_file` and no CLI arg.

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

### history.json (history.go)

Array of `HistoryEntry` in `work_dir`, sorted by `date` descending, upserted by date:

```json
[
  {
    "date": "2026-08-04",
    "file": "04_08.json",
    "start": "06:44",
    "end": "15:44",
    "breaks": 1,
    "breaks_dur": "1h0m",
    "worked": "09:00",
    "saved_at": "15:44"
  }
]
```

## Testing Patterns

Tests are table-driven and live alongside the code as `*_test.go`. Typical shape:

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

Coverage by file:
- `calculator_test.go` — time/duration parsing, break validation, TZ conversion, both calc modes
- `model_test.go` — dirty tracking, field navigation, save/load, mode transitions
- `storage_test.go` — JSON round-trip, missing file handling
- `config_test.go` — defaults, TOML loading, preset parsing
- `locale_test.go` — non-empty fields, RU/EN differ
- `history_test.go` — append/upsert, sort order

Tests use real temp directories for storage/config tests. No external mocking framework — plain `testing` only.

## Gotchas

1. **Tilde expansion**: Config and CLI paths use `~/` which must be expanded via `toAbsolutePath()` (model.go). Never pass a `~/...` path directly to `os` calls.

2. **Break validation**: Breaks must have `from < to` and cannot be before the work start time (calculator.go:86-116). Invalid breaks are surfaced as `m.err`, not panics.

3. **Timezone conversion**: Only applies if `addTZ=true` AND `timezone` is set in config. Uses a real date (today) for accurate DST handling (calculator.go:118-150). `m.endTimeRaw` keeps the un-converted time for `progressInfo()`.

4. **Snapshot optimization**: `recalculate()` skips if `lastSnapshot` is unchanged. When adding a new input field, include it in the snapshot or the result will not refresh.

5. **Clipboard copy**: Uses OSC 52 escape sequence written to `/dev/tty` (model.go), not a clipboard library. May not work in all terminals / over SSH without OSC 52 support.

6. **Progress info**: `progressInfo()` in `view.go` computes elapsed/remaining time and handles overnight shifts by trying multiple day offsets to find the work window. Do not assume the work day is within the same calendar day.

7. **End-of-day notification**: `m.dayEnded` / `m.notified` track whether the timer has crossed the computed end time. The tick loop fires the notification once.

8. **Autogenerated filename**: Without CLI args and without `default_file`, the app opens `<work_dir>/DD_MM.json` (today's day/month). The file need not exist — it's created on save.

9. **Single Go package**: Everything is `package main` at the root. There is no `pkg/`, `internal/`, or `cmd/`. New code goes in the root as a new `.go` file or extends an existing one.

10. **v2 import paths**: Dependencies use `charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2` (not the old `github.com/charmbracelet/...` paths). The module path itself is `github.com/Ratatoouille/work-timer/v2`.

## Key Dependencies

- `charm.land/bubbletea/v2` — TUI framework (Elm architecture)
- `charm.land/bubbles/v2` — `textinput` component used for all editable fields
- `charm.land/lipgloss/v2` — Styling / layout
- `github.com/BurntSushi/toml` — Config parsing

Go 1.25+ required (see `go.mod`).
