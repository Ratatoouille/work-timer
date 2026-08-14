package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

const version = "2.2.2"

func main() {
	saveFile := ""

	for _, arg := range os.Args[1:] {
		switch arg {
		case "--version", "-v":
			fmt.Printf("work-timer %s\n", version)
			os.Exit(0)
		case "--help", "-h":
			fmt.Fprintf(os.Stderr, "Usage: work-timer [file.json]\n\n")
			fmt.Fprintf(os.Stderr, "  file.json   load this save file on startup\n\n")
			fmt.Fprintf(os.Stderr, "Flags:\n")
			fmt.Fprintf(os.Stderr, "  -v, --version   print version and exit\n")
			fmt.Fprintf(os.Stderr, "  -h, --help      show this help\n\n")
			fmt.Fprintf(os.Stderr, "Config: %s\n", ConfigPath)
			os.Exit(0)
		default:
			if saveFile == "" {
				saveFile = arg
			}
		}
	}

	p := tea.NewProgram(NewModel(saveFile))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
