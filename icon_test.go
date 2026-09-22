package main

import (
	"image/color"
	"testing"
)

func TestParseConfigColor(t *testing.T) {
	fallback := color.NRGBA{R: 0x5f, G: 0xaf, B: 0xff, A: 0xff}

	tests := []struct {
		name  string
		input string
		want  color.NRGBA
	}{
		{"empty fallback", "", fallback},
		{"hex full", "#ff0000", color.NRGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}},
		{"hex short", "#0f0", color.NRGBA{R: 0x00, G: 0xff, B: 0x00, A: 0xff}},
		{"invalid hex", "#zzzzzz", fallback},
		{"ansi black", "0", color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}},
		{"ansi white", "15", color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}},
		{"ansi out of range", "300", fallback},
		{"ansi negative", "-1", fallback},
		{"garbage", "nope", fallback},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseConfigColor(tt.input)
			if got != tt.want {
				t.Errorf("parseConfigColor(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAnsiToNRGBA(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want color.NRGBA
	}{
		{"base red", 9, color.NRGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}},
		{"cube start", 16, color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}},
		{"cube white", 231, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}},
		{"gray start", 232, color.NRGBA{R: 8, G: 8, B: 8, A: 0xff}},
		{"gray end", 255, color.NRGBA{R: 238, G: 238, B: 238, A: 0xff}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ansiToNRGBA(tt.n)
			if got != tt.want {
				t.Errorf("ansiToNRGBA(%d) = %+v, want %+v", tt.n, got, tt.want)
			}
		})
	}
}