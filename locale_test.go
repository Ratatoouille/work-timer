package main

import "testing"

func TestLoadLocale(t *testing.T) {
	tests := []struct {
		lang   string
		wantRU bool // true = ожидаем русский
	}{
		{lang: "ru", wantRU: true},
		{lang: "", wantRU: true},   // пустой → русский
		{lang: "xx", wantRU: true}, // неизвестный → русский
		{lang: "en", wantRU: false},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			loc := LoadLocale(tt.lang)
			isRU := loc.SectionMain == localeRU.SectionMain

			if isRU != tt.wantRU {
				t.Errorf("LoadLocale(%q).SectionMain = %q, wantRU=%v",
					tt.lang, loc.SectionMain, tt.wantRU)
			}
		})
	}
}

func TestLocaleAllFieldsNonEmpty(t *testing.T) {
	check := func(name string, loc Locale) {
		fields := map[string]string{
			"ModeNormal":           loc.ModeNormal,
			"ModeInsert":           loc.ModeInsert,
			"NoFileSelected":       loc.NoFileSelected,
			"SectionMain":          loc.SectionMain,
			"SectionBreaks":        loc.SectionBreaks,
			"SectionParams":        loc.SectionParams,
			"SectionResult":        loc.SectionResult,
			"FieldStart":           loc.FieldStart,
			"FieldEnd":             loc.FieldEnd,
			"RemainingLabel":       loc.RemainingLabel,
			"HeroDoneLabel":        loc.HeroDoneLabel,
			"StateRunning":         loc.StateRunning,
			"StatePaused":          loc.StatePaused,
			"StateDone":            loc.StateDone,
			"StateOvertime":        loc.StateOvertime,
			"StatMode":             loc.StatMode,
			"StatMode1":            loc.StatMode1,
			"StatMode2":            loc.StatMode2,
			"StatOn":               loc.StatOn,
			"StatOff":              loc.StatOff,
			"BreakWord":            loc.BreakWord,
			"DurFormatHours":       loc.DurFormatHours,
			"DurFormatMins":        loc.DurFormatMins,
			"FieldRemainingTime":   loc.FieldRemainingTime,
			"FieldWorked":          loc.FieldWorked,
			"FieldPlan":            loc.FieldPlan,
			"CheckboxAddTZ":        loc.CheckboxAddTZ,
			"CheckboxShowIn":       loc.CheckboxShowIn,
			"BreakLeft":            loc.BreakLeft,
			"BreakRight":           loc.BreakRight,
			"ResultLabel":          loc.ResultLabel,
			"Remaining":            loc.Remaining,
			"RemainingCap":         loc.RemainingCap,
			"Status":               loc.Status,
			"CtrlNav":              loc.CtrlNav,
			"CtrlSave":             loc.CtrlSave,
			"CtrlOpen":             loc.CtrlOpen,
			"CtrlHelp":             loc.CtrlHelp,
			"CtrlQuit":             loc.CtrlQuit,
			"CtrlDelete":           loc.CtrlDelete,
			"CtrlRename":           loc.CtrlRename,
			"CtrlNew":              loc.CtrlNew,
			"CtrlCancel":           loc.CtrlCancel,
			"CtrlSearch":           loc.CtrlSearch,
			"FileListHint2":        loc.FileListHint2,
			"FileSearchLabel":      loc.FileSearchLabel,
			"SaveTitle":            loc.SaveTitle,
			"SaveHint":             loc.SaveHint,
			"StatusCopied":         loc.StatusCopied,
			"StatusUnsaved":        loc.StatusUnsaved,
			"StatusSavedAs":        loc.StatusSavedAs,
			"StatusLoadedFrom":     loc.StatusLoadedFrom,
			"StatusSaveError":      loc.StatusSaveError,
			"StatusLoadError":      loc.StatusLoadError,
			"ErrInvalidStartTime":  loc.ErrInvalidStartTime,
			"ErrWorkedExceedsPlan": loc.ErrWorkedExceedsPlan,
			"ErrNoTimeInput":       loc.ErrNoTimeInput,
			"ErrTimezoneNotSet":    loc.ErrTimezoneNotSet,
			"PlaceholderFile":      loc.PlaceholderFile,
			"PlaceholderTime":      loc.PlaceholderTime,
			"HelpTitle":            loc.HelpTitle,
			"HelpNormal":           loc.HelpNormal,
			"HowModes":             loc.HowModes,
			"CtrlAddBreak":         loc.CtrlAddBreak,
			"HelpMode1Desc":        loc.HelpMode1Desc,
		}

		for field, value := range fields {
			if value == "" {
				t.Errorf("locale %s: field %s is empty", name, field)
			}
		}
	}

	check("RU", localeRU)
	check("EN", localeEN)
}

func TestLocaleRUAndENDiffer(t *testing.T) {
	// Хотя бы основные строки должны отличаться
	if localeRU.SectionMain == localeEN.SectionMain {
		t.Error("SectionMain should differ between RU and EN")
	}
	if localeRU.FieldStart == localeEN.FieldStart {
		t.Error("FieldStart should differ between RU and EN")
	}
	if localeRU.NoFileSelected == localeEN.NoFileSelected {
		t.Error("NoFileSelected should differ between RU and EN")
	}
	if localeRU.StatusCopied == localeEN.StatusCopied {
		t.Error("StatusCopied should differ between RU and EN")
	}
}
