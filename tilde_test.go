package main

import (
	"os"
	"path/filepath"
	"testing"
)

// tildePath заменяет префикс домашней папки на "~".
func TestTildePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	p := tildePath(filepath.Join(home, ".config", "work_timer", "dates"))
	want := "~/" + filepath.Join(".config", "work_timer", "dates")
	if p != want {
		t.Errorf("got %q, want %q", p, want)
	}
}
