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
			"HeaderHints":          loc.HeaderHints,
			"NoFileSelected":       loc.NoFileSelected,
			"SectionMain":          loc.SectionMain,
			"SectionBreaks":        loc.SectionBreaks,
			"FieldStart":           loc.FieldStart,
			"FieldRemainingTime":   loc.FieldRemainingTime,
			"FieldWorked":          loc.FieldWorked,
			"FieldPlan":            loc.FieldPlan,
			"CheckboxAddTZ":        loc.CheckboxAddTZ,
			"CheckboxShowIn":       loc.CheckboxShowIn,
			"BreakLeft":            loc.BreakLeft,
			"BreakRight":           loc.BreakRight,
			"ResultLabel":          loc.ResultLabel,
			"CtrlNav":              loc.CtrlNav,
			"CtrlNow":              loc.CtrlNow,
			"CtrlSave":             loc.CtrlSave,
			"CtrlOpen":             loc.CtrlOpen,
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
			"HelpText":             loc.HelpText,
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
